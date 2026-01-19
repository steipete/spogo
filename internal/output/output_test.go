package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, err := New(Options{Format: "wat"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewUsesDefaults(t *testing.T) {
	w, err := New(Options{})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if w.Format != FormatHuman {
		t.Fatalf("expected human format, got %q", w.Format)
	}
	if w.Out == nil || w.Err == nil {
		t.Fatalf("expected default out/err")
	}
}

func TestThemeEnabledIncludesText(t *testing.T) {
	th := theme(true)
	if got := th.Accent("x"); got == "" || !strings.Contains(got, "x") {
		t.Fatalf("expected accent to include input, got %q", got)
	}
}

func TestEmitJSON(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatJSON, Out: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(map[string]any{"ok": true}, nil, nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if !strings.Contains(out.String(), "\"ok\"") {
		t.Fatalf("expected json output, got %q", out.String())
	}
}

func TestEmitPlain(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatPlain, Out: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(nil, []string{"a\tb", "c\td"}, nil); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if out.String() != "a\tb\nc\td\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestEmitHuman(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Out: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(nil, nil, []string{"hello"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if out.String() != "hello\n" {
		t.Fatalf("unexpected output: %q", out.String())
	}
}

func TestEmitHumanQuiet(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Out: out, Quiet: true})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.Emit(nil, nil, []string{"hello"}); err != nil {
		t.Fatalf("emit: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("expected no output, got %q", out.String())
	}
}

func TestNilWriterEmitErrors(t *testing.T) {
	var w *Writer
	if err := w.Emit(nil, nil, nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWriteLinesEmptyNoop(t *testing.T) {
	out := &bytes.Buffer{}
	w, err := New(Options{Format: FormatPlain, Out: out})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := w.WriteLines(nil); err != nil {
		t.Fatalf("write: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("expected no output, got %q", out.String())
	}
}

func TestWarnfAndErrorfWriteToErr(t *testing.T) {
	errOut := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Err: errOut})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	w.Warnf("warn %d", 1)
	w.Errorf("err %d", 2)

	s := errOut.String()
	if !strings.Contains(s, "warn 1") || !strings.Contains(s, "err 2") {
		t.Fatalf("expected warn+err, got %q", s)
	}
}

func TestWarnfAndErrorfColorizeWhenEnabled(t *testing.T) {
	errOut := &bytes.Buffer{}
	w, err := New(Options{Format: FormatHuman, Err: errOut, Color: true})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	w.Warnf("warn %d", 1)
	w.Errorf("err %d", 2)
	s := errOut.String()
	if !strings.Contains(s, "warn 1") || !strings.Contains(s, "err 2") {
		t.Fatalf("expected warn+err, got %q", s)
	}
}

func TestNilWriterWarnfAndErrorfNoop(t *testing.T) {
	var w *Writer
	w.Warnf("warn")
	w.Errorf("err")
}
