package cli

import "testing"

func TestOutputFormatVariants(t *testing.T) {
	if _, err := outputFormat(true, true); err == nil {
		t.Fatalf("expected error")
	}
	if format, err := outputFormat(true, false); err != nil || format != "json" {
		t.Fatalf("expected json")
	}
	if format, err := outputFormat(false, true); err != nil || format != "plain" {
		t.Fatalf("expected plain")
	}
	if format, err := outputFormat(false, false); err != nil || format != "human" {
		t.Fatalf("expected human")
	}
}
