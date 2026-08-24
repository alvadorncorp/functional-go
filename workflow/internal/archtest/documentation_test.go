package archtest

import (
	"fmt"
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
		"docs/adr/workflow/adr-001-workflow-as-reusable-effectful-composition.md",
		"docs/adr/workflow/adr-002-action-as-deferred-context-bound-computation.md",
		"docs/adr/workflow/adr-003-functional-dependency-injection-through-capabilities.md",
		"docs/adr/workflow/adr-004-explicit-transaction-scopes.md",
		"docs/adr/workflow/adr-005-modeling-integration-boundary.md",
		"docs/adr/workflow/adr-006-no-distributed-transaction-abstraction.md",
		"docs/adr/workflow/adr-007-functional-integration-without-effect-duplication.md",
		"docs/adr/workflow/adr-008-transactional-outbox-as-application-composition.md",
		"docs/adr/workflow/adr-009-composition-and-resource-safety.md",
		"docs/adr/workflow/adr-010-plan-as-explicit-non-executable-workflow-description.md",
		"docs/adr/workflow/adr-011-explicit-resilience-as-action-transformation.md",
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
		"README.md": {"Contramap", "Provide", "Bracket", "plan.Plan", "resilience", "Run executa"},
		"docs/overview.md": {
			"Plan", "Resilience", "Boundary", "Run", "Dados de negócio", "ctx.Value",
		},
		"docs/architecture.md": {
			"plan", "resilience", "resilience -> action", "WithTimeout", "preempção",
		},
		"docs/actions.md": {"errors.Join", "Attempt", "Bracket"},
		"docs/plan.md":    {"Plan descreve", "não executa", "KindEmpty", "defensiva", "Specification"},
		"docs/resilience.md": {
			"Attempts", "1-based", "idempot", "cooperativo", "Timeout(Retry", "Retry(Timeout",
		},
		"docs/composition.md":          {"FlatMap", "Zip", "Both"},
		"docs/dependency-injection.md": {"fornece input", "Container"},
		"docs/examples.md": {
			"AttemptResult.Err", "action.Run", "transaction.Within",
			"CreateOrderHTTPRequest", "CreateOrderMessage", "CreateOrderPlan",
			"faults.IsUnavailable", "Timeout(30", "AppendOutbox",
		},
		"docs/transactions.md":             {"Bracket", "Within", "atomicidade"},
		"docs/tdd/workflow/action.md":      {"AcquireUseRelease", "errors.Join", "panic"},
		"docs/tdd/workflow/composition.md": {"Contramap", "Provide", "DI container"},
	}
}

// forbiddenDocumentationPhrases are obsolete claims that must not reappear as
// present-tense guidance. Historical discussion under rejected alternatives or
// non-goals is out of scope for these narrow phrases.
func forbiddenDocumentationPhrases() []string {
	return []string{
		"Plan executa Action",
		"Plan compila para Action",
		"Timeout preempta",
		"Timeout interrompe goroutine",
		"retry está embutido em action",
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
		[]string{"Contramap", "plan.Plan", "Run executa"},
	)
	if len(missing) != 1 || missing[0] != "Run executa" {
		t.Fatalf("missing tokens = %v, want [Run executa]", missing)
	}

	found := foundForbiddenPhrases(
		"legacy note: Plan executa Action when misused",
		[]string{"Plan executa Action", "Timeout preempta"},
	)
	if len(found) != 1 || found[0] != "Plan executa Action" {
		t.Fatalf("forbidden phrases = %v, want [Plan executa Action]", found)
	}
}

func TestDocumentationADRIndexAndSupersession(t *testing.T) {
	root := moduleRoot(t)

	adr009Path := filepath.Join(
		root, "docs", "adr", "workflow", "adr-009-composition-and-resource-safety.md",
	)
	adr009, err := os.ReadFile(adr009Path)
	if err != nil {
		t.Fatalf("read ADR-009: %v", err)
	}
	if !strings.Contains(string(adr009), "adr-011-explicit-resilience-as-action-transformation.md") {
		t.Error("ADR-009 must record the ADR-011 supersession note")
	}

	index, err := os.ReadFile(filepath.Join(root, "docs", "README.md"))
	if err != nil {
		t.Fatalf("read docs index: %v", err)
	}
	indexText := string(index)
	adrLink := regexp.MustCompile(`adr/workflow/adr-(\d{3})-[a-z0-9-]+\.md`)
	matches := adrLink.FindAllStringSubmatch(indexText, -1)
	if len(matches) != 12 {
		t.Fatalf("docs index ADR links = %d, want 12", len(matches))
	}
	seen := make(map[string]int, len(matches))
	for _, match := range matches {
		seen[match[1]]++
	}
	for i := 1; i <= 12; i++ {
		key := formatADRNumber(i)
		if count := seen[key]; count != 1 {
			t.Errorf("ADR-%s link count = %d, want 1", key, count)
		}
	}
	if !strings.Contains(indexText, "plan.md") || !strings.Contains(indexText, "resilience.md") {
		t.Error("docs index must link Plan and Resilience guides")
	}
	if !strings.Contains(indexText, "CreateOrder") {
		t.Error("docs index examples entry must mention CreateOrder")
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

func formatADRNumber(n int) string {
	return fmt.Sprintf("%03d", n)
}
