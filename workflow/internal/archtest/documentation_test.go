package archtest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// requiredDocuments is the module's documentation spine. A missing file here means
// a decision or a design was delivered without being written down.
func requiredDocuments() []string {
	return []string{
		"README.md",
		"docs/README.md",
		"docs/overview.md",
		"docs/architecture.md",
		"docs/actions.md",
		"docs/plan.md",
		"docs/resilience.md",
		"docs/composition.md",
		"docs/dependency-injection.md",
		"docs/transactions.md",
		"docs/modeling.md",
		"docs/examples.md",
		"docs/adr/workflow/workflow-as-reusable-effectful-composition.md",
		"docs/adr/workflow/action-as-deferred-context-bound-computation.md",
		"docs/adr/workflow/functional-dependency-injection-through-capabilities.md",
		"docs/adr/workflow/explicit-transaction-scopes.md",
		"docs/adr/workflow/modeling-integration-boundary.md",
		"docs/adr/workflow/no-distributed-transaction-abstraction.md",
		"docs/adr/workflow/functional-integration-without-effect-duplication.md",
		"docs/adr/workflow/transactional-outbox-as-application-composition.md",
		"docs/adr/workflow/composition-and-resource-safety.md",
		"docs/adr/workflow/plan-as-explicit-non-executable-workflow-description.md",
		"docs/adr/workflow/explicit-resilience-as-action-transformation.md",
		"docs/adr/workflow/no-ddd-runtime-module.md",
		"docs/adr/workflow/sequential-traverse-over-slices.md",
		"docs/adr/workflow/ast-evidence-for-exercised-surface-gate.md",
		"docs/tdd/workflow/core.md",
		"docs/tdd/workflow/action.md",
		"docs/tdd/workflow/composition.md",
		"docs/tdd/workflow/transaction.md",
		"docs/tdd/workflow/modeling-integration.md",
		"docs/tdd/workflow/plan-resilience.md",
	}
}

// moduleRoot walks up from the archtest package to the module directory.
func moduleRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve module root: %v", err)
	}
	return root
}

// TestDocumentationSpine proves every required document exists and that each ADR
// and TDD carries the frontmatter keys the AKB expects.
func TestDocumentationSpine(t *testing.T) {
	root := moduleRoot(t)

	for _, relative := range requiredDocuments() {
		path := filepath.Join(root, relative)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("missing document %s: %v", relative, err)
			continue
		}
		assertFrontmatter(t, relative, string(data))
	}
}

func assertFrontmatter(t *testing.T, relative, content string) {
	t.Helper()

	isADR := strings.Contains(relative, "/adr/")
	isTDD := strings.Contains(relative, "/tdd/")
	if !isADR && !isTDD {
		return
	}

	required := []string{"type:", "id:", "context: workflow", "title:", "status:", "tags:", "date:"}
	for _, key := range required {
		if !strings.Contains(content, key) {
			t.Errorf("%s is missing frontmatter key %q", relative, key)
		}
	}
	if isADR {
		slug := strings.TrimSuffix(filepath.Base(relative), ".md")
		if !strings.Contains(content, "\nid: adr:workflow:"+slug+"\n") {
			t.Errorf("%s must use its filename slug as its ADR ID", relative)
		}
		if !strings.Contains(content, "\nstatus: Accepted\n") {
			t.Errorf("%s must have status: Accepted", relative)
		}
	}
	if isTDD {
		if !strings.Contains(content, "\nstatus: Accepted\n") {
			t.Errorf("%s must have status: Accepted", relative)
		}
		if strings.Contains(content, "deciders:") {
			t.Errorf("%s must not carry a deciders key", relative)
		}
	}
}

func requiredDocumentationTokens() map[string][]string {
	return map[string][]string{
		"README.md": {"Contramap", "Provide", "Bracket", "plan.Plan", "resilience", "action.Run"},
		"docs/overview.md": {
			"Plan", "Resilience", "Boundary", "Run", "explicit input", "ctx.Value",
		},
		"docs/architecture.md": {
			"plan", "resilience", "resilience -> action", "WithTimeout", "preempt",
		},
		"docs/actions.md": {"errors.Join", "Attempt", "Bracket"},
		"docs/plan.md":    {"does not execute Actions", "KindEmpty", "defensive", "Specification"},
		"docs/resilience.md": {
			"Attempts", "1-based", "idempot", "cooperat", "Timeout(Retry", "Retry(Timeout",
		},
		"docs/composition.md":              {"FlatMap", "Zip", "Both", "Traverse"},
		"docs/dependency-injection.md":     {"supplies input, not dependencies", "Container"},
		"docs/examples.md":                 {"action.Run", "transaction.Within", "resilience.Retry"},
		"docs/transactions.md":             {"Bracket", "Within", "atomic"},
		"docs/tdd/workflow/action.md":      {"AcquireUseRelease", "errors.Join", "panic", "Traverse"},
		"docs/tdd/workflow/composition.md": {"Contramap", "Provide", "DI container"},
	}
}

// forbiddenDocumentationPhrases are obsolete claims that must not reappear as
// present-tense guidance. Historical discussion under rejected alternatives or
// non-goals is out of scope for these narrow phrases.
func forbiddenDocumentationPhrases() []string {
	return []string{
		"Plan executes an Action",
		"Plan compiles to an Action",
		"Timeout preempts",
		"Timeout stops goroutines",
		"retry is built into action",
	}
}

func missingDocumentationTokens(content string, tokens []string) []string {
	var missing []string
	for _, token := range tokens {
		if !strings.Contains(content, token) {
			missing = append(missing, token)
		}
	}
	return missing
}

func foundForbiddenPhrases(content string, phrases []string) []string {
	var found []string
	for _, phrase := range phrases {
		if strings.Contains(content, phrase) {
			found = append(found, phrase)
		}
	}
	return found
}

func TestRequiredDocumentationTerminology(t *testing.T) {
	root := moduleRoot(t)
	for relative, tokens := range requiredDocumentationTokens() {
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		for _, token := range missingDocumentationTokens(string(data), tokens) {
			t.Errorf("%s is missing required token %q", relative, token)
		}
	}
}

func TestDocumentationTerminologyGateCanFail(t *testing.T) {
	missing := missingDocumentationTokens(
		"Contramap and Provide without plan.Plan",
		[]string{"Contramap", "plan.Plan", "action.Run"},
	)
	if len(missing) != 1 || missing[0] != "action.Run" {
		t.Fatalf("missing tokens = %v, want [action.Run]", missing)
	}

	found := foundForbiddenPhrases(
		"Plan executes an Action",
		[]string{"Plan executes an Action", "Timeout preempts"},
	)
	if len(found) != 1 || found[0] != "Plan executes an Action" {
		t.Fatalf("forbidden phrases = %v, want [Plan executes an Action]", found)
	}
}

func TestDocumentationADRIndexAndSupersession(t *testing.T) {
	root := moduleRoot(t)

	compositionPath := filepath.Join(root, "docs", "adr", "workflow", "composition-and-resource-safety.md")
	composition, err := os.ReadFile(compositionPath)
	if err != nil {
		t.Fatalf("read composition ADR: %v", err)
	}
	if !strings.Contains(string(composition), "explicit-resilience-as-action-transformation.md") &&
		!strings.Contains(string(composition), "adr:workflow:explicit-resilience-as-action-transformation") {
		t.Error("composition ADR must link to the resilience decision that supersedes part of it")
	}

	index, err := os.ReadFile(filepath.Join(root, "docs", "README.md"))
	if err != nil {
		t.Fatalf("read docs index: %v", err)
	}
	indexText := string(index)
	adrLink := regexp.MustCompile(`\]\((adr/workflow/[a-z0-9-]+\.md)\)`)
	seen := make(map[string]int)
	for _, match := range adrLink.FindAllStringSubmatch(indexText, -1) {
		seen[match[1]]++
	}
	for _, relative := range requiredDocuments() {
		link, isADR := strings.CutPrefix(relative, "docs/")
		if isADR && strings.HasPrefix(link, "adr/") {
			if count := seen[link]; count != 1 {
				t.Errorf("%s link count = %d, want 1", link, count)
			}
			delete(seen, link)
		}
	}
	for link := range seen {
		t.Errorf("unexpected ADR link %s", link)
	}
	if !strings.Contains(indexText, "plan.md") || !strings.Contains(indexText, "resilience.md") {
		t.Error("docs index must link Plan and Resilience guides")
	}
}

func TestDocumentationPathsStayModuleLocal(t *testing.T) {
	root := moduleRoot(t)
	for _, relative := range requiredDocuments() {
		if strings.HasPrefix(relative, "/") || strings.HasPrefix(relative, "../") {
			t.Errorf("document path escapes module: %s", relative)
		}
		if !filepath.IsLocal(relative) {
			t.Errorf("document path is not module-local: %s", relative)
		}
		if _, err := os.Stat(filepath.Join(root, relative)); err != nil {
			t.Errorf("document missing under module root: %s: %v", relative, err)
		}
	}
}

func TestDocumentationForbidsObsoleteClaims(t *testing.T) {
	root := moduleRoot(t)
	phrases := forbiddenDocumentationPhrases()
	// Keep the gate on human guides; ADRs/TDDs may quote rejected alternatives.
	guides := []string{
		"docs/README.md",
		"docs/overview.md",
		"docs/architecture.md",
		"docs/actions.md",
		"docs/plan.md",
		"docs/resilience.md",
		"docs/composition.md",
		"docs/dependency-injection.md",
		"docs/transactions.md",
		"docs/modeling.md",
		"docs/examples.md",
	}
	for _, relative := range guides {
		data, err := os.ReadFile(filepath.Join(root, relative))
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		for _, phrase := range foundForbiddenPhrases(string(data), phrases) {
			t.Errorf("%s contains forbidden obsolete claim %q", relative, phrase)
		}
	}
}
