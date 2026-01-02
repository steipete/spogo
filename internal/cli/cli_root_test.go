package cli

import (
	"testing"

	"github.com/steipete/spogo/internal/output"
)

func TestNewCLI(t *testing.T) {
	if New() == nil {
		t.Fatalf("expected cli")
	}
}

func TestGlobalsSettings(t *testing.T) {
	settings, err := (Globals{JSON: true}).Settings()
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if settings.Format != output.FormatJSON {
		t.Fatalf("format")
	}
	if _, err := (Globals{JSON: true, Plain: true}).Settings(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestOutputFormat(t *testing.T) {
	if f, _ := outputFormat(false, false); f != output.FormatHuman {
		t.Fatalf("expected human")
	}
}

func TestVersionVars(t *testing.T) {
	vars := VersionVars()
	if vars["version"] == "" {
		t.Fatalf("expected version")
	}
}

func TestUsage(t *testing.T) {
	if Usage() == "" {
		t.Fatalf("expected usage")
	}
}
