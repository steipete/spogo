package cli

import (
	"os"
	"testing"
)

func withStdin(t *testing.T, contents string, fn func()) {
	t.Helper()
	old := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString(contents); err != nil {
		_ = r.Close()
		_ = w.Close()
		t.Fatalf("write: %v", err)
	}
	_ = w.Close()
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = old })
	fn()
	_ = r.Close()
}
