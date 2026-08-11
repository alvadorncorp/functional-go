package archtest

import (
	"os"
	"path/filepath"
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
		"docs/composition.md",
		"docs/dependency-injection.md",
		"docs/transactions.md",
		"docs/modeling.md",
		"docs/examples.md",
		"docs/adr/workflow/adr-001-workflow-as-reusable-effectful-composition.md",
		"docs/adr/workflow/adr-002-action-as-deferred-context-bound-computation.md",
		"docs/adr/workflow/adr-003-functional-dependency-injection-through-capabilities.md",
		"docs/adr/workflow/adr-004-explicit-transaction-scopes.md",
		"docs/adr/workflow/adr-005-modeling-integration-boundary.md",
		"docs/adr/workflow/adr-006-no-distributed-transaction-abstraction.md",
		"docs/adr/workflow/adr-007-functional-integration-without-effect-duplication.md",
		"docs/adr/workflow/adr-008-transactional-outbox-as-application-composition.md",
		"docs/adr/workflow/adr-009-composition-and-resource-safety.md",
		"docs/tdd/workflow/core.md",
		"docs/tdd/workflow/action.md",
		"docs/tdd/workflow/composition.md",
		"docs/tdd/workflow/transaction.md",
		"docs/tdd/workflow/modeling-integration.md",
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
	if isADR && !strings.Contains(content, "status: Aceita") {
		t.Errorf("%s must have status: Aceita", relative)
	}
	if isTDD {
		if !strings.Contains(content, "status: Aceito") {
			t.Errorf("%s must have status: Aceito", relative)
		}
		if strings.Contains(content, "deciders:") {
			t.Errorf("%s must not carry a deciders key", relative)
		}
	}
}

func requiredDocumentationTokens() map[string][]string {
	return map[string][]string{
		"README.md":                        {"Contramap", "Provide", "Bracket"},
		"docs/actions.md":                  {"errors.Join", "Attempt", "Bracket"},
		"docs/composition.md":              {"FlatMap", "Zip", "Both"},
		"docs/dependency-injection.md":     {"fornece input", "Container"},
		"docs/examples.md":                 {"AttemptResult.Err", "action.Run", "transaction.Within"},
		"docs/transactions.md":             {"Bracket", "Within", "atomicidade"},
		"docs/tdd/workflow/action.md":      {"AcquireUseRelease", "errors.Join", "panic"},
		"docs/tdd/workflow/composition.md": {"Contramap", "Provide", "DI container"},
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
	missing := missingDocumentationTokens("Contramap and Provide", []string{"Contramap", "Bracket"})
	if len(missing) != 1 || missing[0] != "Bracket" {
		t.Fatalf("missing tokens = %v, want [Bracket]", missing)
	}
}
