package state_test

import (
	"testing"

	"github.com/alvadorncorp/gopackages/functional/persistent"
	"github.com/alvadorncorp/gopackages/functional/state"
)

// Store is a domain-shaped value whose fields are persistent collections.
// Mutations evolve the store by reassigning Map.Set / Set.Add / List.Append
// results; they never convert collections to native map or slice mid-pipeline.
type Store struct {
	Users persistent.Map[string, string]
	Tags  persistent.Set[string]
	Log   persistent.List[string]
}

func emptyStore() Store {
	return Store{
		Users: persistent.MapOf[string, string](nil),
		Tags:  persistent.SetOf[string](),
		Log:   persistent.ListOf[string](),
	}
}

func putUser(id, name string) state.Mutation[Store] {
	return func(s Store) Store {
		s.Users = s.Users.Set(id, name)
		s.Log = s.Log.Append("user:" + id)
		return s
	}
}

func addTag(tag string) state.Mutation[Store] {
	return func(s Store) Store {
		s.Tags = s.Tags.Add(tag)
		s.Log = s.Log.Append("tag:" + tag)
		return s
	}
}

func removeUser(id string) state.Mutation[Store] {
	return func(s Store) Store {
		s.Users = s.Users.Delete(id)
		s.Log = s.Log.Append("del:" + id)
		return s
	}
}

//nolint:gocritic // Store is a test fixture; pointer forces awkward temps for Materialize().
func assertStoreEqual(t *testing.T, got Store, wantUsers map[string]string, wantTags, wantLog []string) {
	t.Helper()

	if got.Users.Len() != len(wantUsers) {
		t.Fatalf("Users.Len() = %d, want %d", got.Users.Len(), len(wantUsers))
	}
	for k, want := range wantUsers {
		gotVal, ok := got.Users.Get(k)
		if !ok {
			t.Fatalf("Users missing key %q", k)
		}
		if gotVal != want {
			t.Fatalf("Users[%q] = %q, want %q", k, gotVal, want)
		}
	}

	if got.Tags.Len() != len(wantTags) {
		t.Fatalf("Tags.Len() = %d, want %d", got.Tags.Len(), len(wantTags))
	}
	for _, tag := range wantTags {
		if !got.Tags.Has(tag) {
			t.Fatalf("Tags missing %q", tag)
		}
	}

	if !got.Log.Equal(persistent.ListOf(wantLog...)) {
		t.Fatalf("Log = %v, want %v", got.Log.Items(), wantLog)
	}
}

func TestPersistentStoreMaterializeAfterSeveralApplies(t *testing.T) {
	t.Parallel()

	s := state.Of(emptyStore()).
		Apply(putUser("u1", "alice")).
		Apply(putUser("u2", "bob")).
		Apply(addTag("admin")).
		Apply(addTag("beta")).
		Apply(removeUser("u1"))

	got := s.Materialize()
	assertStoreEqual(t, got,
		map[string]string{"u2": "bob"},
		[]string{"admin", "beta"},
		[]string{"user:u1", "user:u2", "tag:admin", "tag:beta", "del:u1"},
	)
	if got := s.Version(); got != 5 {
		t.Fatalf("Version() = %d, want 5", got)
	}
}

func TestPersistentStoreOldVersionsRemainIntact(t *testing.T) {
	t.Parallel()

	s0 := state.Of(emptyStore())
	s1 := s0.Apply(putUser("u1", "alice"))
	s2 := s1.Apply(addTag("vip"))
	s3 := s2.Apply(putUser("u2", "bob"))

	// Materialize newest first, then prove older snapshots are unchanged.
	assertStoreEqual(t, s3.Materialize(),
		map[string]string{"u1": "alice", "u2": "bob"},
		[]string{"vip"},
		[]string{"user:u1", "tag:vip", "user:u2"},
	)
	assertStoreEqual(t, s2.Materialize(),
		map[string]string{"u1": "alice"},
		[]string{"vip"},
		[]string{"user:u1", "tag:vip"},
	)
	assertStoreEqual(t, s1.Materialize(),
		map[string]string{"u1": "alice"},
		nil,
		[]string{"user:u1"},
	)
	assertStoreEqual(t, s0.Materialize(),
		map[string]string{},
		nil,
		nil,
	)

	// Re-materialize s1 after s3 to prove structural sharing did not corrupt it.
	assertStoreEqual(t, s1.Materialize(),
		map[string]string{"u1": "alice"},
		nil,
		[]string{"user:u1"},
	)
}

func TestPersistentStoreMutationsUsePersistentAPIsOnly(t *testing.T) {
	t.Parallel()

	// Mutations must keep field types as persistent collections throughout
	// the plan: no conversion to native map/slice for intermediate writes.
	// The putUser/addTag helpers reassign only Map.Set / Set.Add / List.Append
	// results, so Materialize observes the same persistent types end-to-end.
	s := state.Of(emptyStore()).
		Apply(func(st Store) Store {
			st.Users = st.Users.Set("k", "v")
			return st
		}).
		Apply(func(st Store) Store {
			// Read via persistent Get (not Items map conversion) then write
			// via Set/Add/Append only.
			name, _ := st.Users.Get("k")
			st.Users = st.Users.Set("k", name+"!")
			st.Tags = st.Tags.Add("t")
			st.Log = st.Log.Append("step")
			return st
		})

	got := s.Materialize()
	if v, ok := got.Users.Get("k"); !ok || v != "v!" {
		t.Fatalf("Users[k] = (%q, %v), want (\"v!\", true)", v, ok)
	}
	if !got.Tags.Has("t") {
		t.Fatal("Tags missing t")
	}
	if got.Log.Len() != 1 || got.Log.At(0) != "step" {
		t.Fatalf("Log = %v, want [step]", got.Log.Items())
	}

	// Field types remain the persistent collection types (Store fields);
	// mutations reassign Map.Set / Set.Add / List.Append results only.
}

func TestPersistentStoreBranchingStatesIndependent(t *testing.T) {
	t.Parallel()

	parent := state.Of(emptyStore()).
		Apply(putUser("u1", "alice")).
		Apply(addTag("base"))

	left := parent.
		Apply(putUser("u2", "left")).
		Apply(addTag("left-only"))

	right := parent.
		Apply(putUser("u3", "right")).
		Apply(removeUser("u1"))

	assertStoreEqual(t, parent.Materialize(),
		map[string]string{"u1": "alice"},
		[]string{"base"},
		[]string{"user:u1", "tag:base"},
	)
	assertStoreEqual(t, left.Materialize(),
		map[string]string{"u1": "alice", "u2": "left"},
		[]string{"base", "left-only"},
		[]string{"user:u1", "tag:base", "user:u2", "tag:left-only"},
	)
	assertStoreEqual(t, right.Materialize(),
		map[string]string{"u3": "right"},
		[]string{"base"},
		[]string{"user:u1", "tag:base", "user:u3", "del:u1"},
	)

	// Further left evolution must not affect right or parent.
	left2 := left.Apply(removeUser("u1"))
	assertStoreEqual(t, left2.Materialize(),
		map[string]string{"u2": "left"},
		[]string{"base", "left-only"},
		[]string{"user:u1", "tag:base", "user:u2", "tag:left-only", "del:u1"},
	)
	assertStoreEqual(t, left.Materialize(),
		map[string]string{"u1": "alice", "u2": "left"},
		[]string{"base", "left-only"},
		[]string{"user:u1", "tag:base", "user:u2", "tag:left-only"},
	)
	assertStoreEqual(t, right.Materialize(),
		map[string]string{"u3": "right"},
		[]string{"base"},
		[]string{"user:u1", "tag:base", "user:u3", "del:u1"},
	)
	assertStoreEqual(t, parent.Materialize(),
		map[string]string{"u1": "alice"},
		[]string{"base"},
		[]string{"user:u1", "tag:base"},
	)
}

func TestPersistentStoreTransactionCommit(t *testing.T) {
	t.Parallel()

	parent := state.Of(emptyStore()).
		Apply(putUser("u1", "alice")).
		Apply(addTag("seed"))

	wantParentUsers := map[string]string{"u1": "alice"}
	wantParentTags := []string{"seed"}
	wantParentLog := []string{"user:u1", "tag:seed"}
	wantParentVer := parent.Version()

	tx := parent.Begin()
	tx.Apply(putUser("u2", "bob")).
		Apply(addTag("tx")).
		Apply(removeUser("u1"))
	committed := tx.Commit()

	// Parent frozen through Commit (lazy plan only).
	assertStoreEqual(t, parent.Materialize(), wantParentUsers, wantParentTags, wantParentLog)
	if got := parent.Version(); got != wantParentVer {
		t.Fatalf("parent.Version() after Commit = %d, want %d", got, wantParentVer)
	}

	assertStoreEqual(t, committed.Materialize(),
		map[string]string{"u2": "bob"},
		[]string{"seed", "tx"},
		[]string{"user:u1", "tag:seed", "user:u2", "tag:tx", "del:u1"},
	)
	if got := committed.Version(); got != wantParentVer+3 {
		t.Fatalf("committed.Version() = %d, want %d", got, wantParentVer+3)
	}

	// Parent still intact after materializing the committed branch.
	assertStoreEqual(t, parent.Materialize(), wantParentUsers, wantParentTags, wantParentLog)
}
