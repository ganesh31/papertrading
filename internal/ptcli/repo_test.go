package ptcli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepoRoot(t *testing.T) {
	root, err := RepoRoot()
	if err != nil {
		t.Fatal(err)
	}
	goMod := filepath.Join(root, "go.mod")
	if st, err := os.Stat(goMod); err != nil || st.IsDir() {
		t.Fatalf("expected root go.mod at %s", goMod)
	}
	ptMain := filepath.Join(root, "cmd", "pt", "main.go")
	if st, err := os.Stat(ptMain); err != nil || st.IsDir() {
		t.Fatalf("expected cmd/pt at %s", ptMain)
	}
}
