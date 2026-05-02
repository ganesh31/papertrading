package ptcli

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// RepoRoot is the monorepo root (directory containing the root go.mod).
func RepoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("ptcli: runtime.Caller failed")
	}
	// This file lives at internal/ptcli/repo.go → two levels up is repo root.
	dir := filepath.Dir(file)
	root := filepath.Clean(filepath.Join(dir, "..", ".."))
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if st, err := os.Stat(filepath.Join(abs, "go.mod")); err != nil || st.IsDir() {
		return "", errors.New("ptcli: could not locate repo root (go.mod missing)")
	}
	return abs, nil
}
