package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestEmitJSON(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatJSON, Out: out, Err: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(map[string]string{"status": "ok"}, nil, nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if !strings.Contains(out.String(), "\"status\"") {
		t.Fatalf("expected json output")
	}
}

func TestEmitPlain(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatPlain, Out: out, Err: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(nil, []string{"a", "b"}, nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if out.String() != "a\nb\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestEmitHuman(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Out: out, Err: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(nil, nil, []string{"hello"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if strings.TrimSpace(out.String()) != "hello" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestEmitHumanQuiet(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Out: out, Err: out, Quiet: true})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(nil, nil, []string{"hello"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("expected empty output, got %q", out.String())
	}
}

func TestErrorf(t *testing.T) {
	errOut := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Out: &bytes.Buffer{}, Err: errOut})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	w.Errorf("oops %d", 1)
	if !strings.Contains(errOut.String(), "oops 1") {
		t.Fatalf("expected error output")
	}
}

func TestNewInvalidFormat(t *testing.T) {
	if _, err := New(Options{Format: "bad"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestThemeColor(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Out: out, Err: out, Color: true})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	_ = w.Theme.Accent("hi")
}

func TestNewDefaults(t *testing.T) {
	w, err := New(Options{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if w.Format != FormatHuman {
		t.Fatalf("expected human")
	}
}

func TestEmitNilWriter(t *testing.T) {
	var w *Writer
	if err := w.Emit(nil, nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWriteLinesEmpty(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatPlain, Out: out, Err: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.WriteLines(nil); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func TestErrorfNilWriter(t *testing.T) {
	var w *Writer
	w.Errorf("oops")
}
