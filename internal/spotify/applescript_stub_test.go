//go:build !darwin

package spotify

import "testing"

func TestNewAppleScriptClient_NonDarwin(t *testing.T) {
	t.Parallel()

	client, err := NewAppleScriptClient(AppleScriptOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
	if client != nil {
		t.Fatal("expected nil client")
	}
}

