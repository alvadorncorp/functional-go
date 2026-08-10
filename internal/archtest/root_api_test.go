package archtest

import (
	"encoding/json"
	"os/exec"
	"slices"
	"testing"
)

func TestRootPackageRemainsDocOnly(t *testing.T) {
	t.Parallel()
	out, err := exec.Command("go", "list", "-json", "github.com/alvadorncorp/gopackages/functional").Output()
	if err != nil {
		t.Fatalf("go list root package: %v", err)
	}
	var listed struct{ GoFiles []string }
	if err := json.Unmarshal(out, &listed); err != nil {
		t.Fatalf("decode root package: %v", err)
	}
	if !slices.Equal(listed.GoFiles, []string{"doc.go"}) {
		t.Fatalf("root GoFiles = %v, want [doc.go]", listed.GoFiles)
	}
}
