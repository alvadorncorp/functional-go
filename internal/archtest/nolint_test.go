package archtest

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoNolint fails if any //nolint directive reappears under this module.
func TestNoNolint(t *testing.T) {
	assertNoNolint(t, filepath.Dir(goModPath(t)))
}

func TestNolintGateCanFail(t *testing.T) {
	// Build fixtures without a contiguous banned directive in this source file.
	gocriticNolint := "//" + "nolint:" + "gocritic" + " // hugeParam: temporary suppression"
	if !lineHasNolint(gocriticNolint) {
		t.Error("lineHasNolint missed gocritic nolint directive")
	}
	multi := "f(x) //" + "nolint:errcheck,gocritic"
	if !lineHasNolint(multi) {
		t.Error("lineHasNolint missed nolint in a multi-linter list")
	}
	if !lineHasNolint("//" + "nolint:all") {
		t.Error("lineHasNolint missed all")
	}
	if !lineHasNolint("//" + "nolint:gosec // tables are validated identifiers") {
		t.Error("lineHasNolint missed gosec nolint directive")
	}
	if !lineHasNolint(`u := "http://example.com" //` + "nolint:gosec") {
		t.Error("lineHasNolint missed trailing nolint after a URL string")
	}
	if lineHasNolint("// gocritic hugeParam sizeThreshold=80. Store is a test fixture.") {
		t.Error("lineHasNolint flagged a non-nolint hugeParam comment")
	}
}

// lineHasNolint reports whether line carries a //nolint directive.
// The first // inside a double-quoted string is ignored so a trailing
// directive after a URL is still detected.
func lineHasNolint(line string) bool {
	inString := false
	escaped := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == '/' && i+1 < len(line) && line[i+1] == '/' {
			comment := strings.TrimSpace(line[i+2:])
			return strings.HasPrefix(comment, "nolint")
		}
	}
	return false
}

func assertNoNolint(t *testing.T, root string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor":
				return filepath.SkipDir
			default:
				return nil
			}
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(f)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			if lineHasNolint(scanner.Text()) {
				rel, relErr := filepath.Rel(root, path)
				if relErr != nil {
					rel = path
				}
				t.Errorf("%s:%d: nolint directive is banned", rel, lineNo)
			}
		}
		scanErr := scanner.Err()
		if closeErr := f.Close(); closeErr != nil && scanErr == nil {
			scanErr = closeErr
		}
		return scanErr
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}
