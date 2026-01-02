package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

type Format string

const (
	FormatHuman Format = "human"
	FormatJSON  Format = "json"
	FormatPlain Format = "plain"
)

type Theme struct {
	Accent  func(a ...any) string
	Muted   func(a ...any) string
	Success func(a ...any) string
	Warn    func(a ...any) string
	Error   func(a ...any) string
	Bold    func(a ...any) string
}

type Writer struct {
	Format Format
	Out    io.Writer
	Err    io.Writer
	Color  bool
	Theme  Theme
	Quiet  bool
}

type Options struct {
	Format Format
	Color  bool
	Out    io.Writer
	Err    io.Writer
	Quiet  bool
}

func New(opts Options) (*Writer, error) {
	format := opts.Format
	if format == "" {
		format = FormatHuman
	}
	if format != FormatHuman && format != FormatJSON && format != FormatPlain {
		return nil, fmt.Errorf("unknown output format %q", format)
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}
	if opts.Err == nil {
		opts.Err = os.Stderr
	}
	w := &Writer{
		Format: format,
		Out:    opts.Out,
		Err:    opts.Err,
		Color:  opts.Color,
		Quiet:  opts.Quiet,
	}
	w.Theme = theme(opts.Color)
	return w, nil
}

func (w *Writer) Emit(value any, plainLines []string, humanLines []string) error {
	if w == nil {
		return errors.New("nil writer")
	}
	switch w.Format {
	case FormatJSON:
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w.Out, string(data))
		return err
	case FormatPlain:
		return w.WriteLines(plainLines)
	default:
		if w.Quiet {
			return nil
		}
		return w.WriteLines(humanLines)
	}
}

func (w *Writer) WriteLines(lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	_, err := fmt.Fprintln(w.Out, strings.Join(lines, "\n"))
	return err
}

func (w *Writer) Errorf(format string, args ...any) {
	if w == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	if w.Color {
		msg = w.Theme.Error(msg)
	}
	_, _ = fmt.Fprintln(w.Err, msg)
}

func theme(enable bool) Theme {
	if !enable {
		return Theme{
			Accent:  fmt.Sprint,
			Muted:   fmt.Sprint,
			Success: fmt.Sprint,
			Warn:    fmt.Sprint,
			Error:   fmt.Sprint,
			Bold:    fmt.Sprint,
		}
	}
	return Theme{
		Accent:  color.New(color.FgCyan, color.Bold).SprintFunc(),
		Muted:   color.New(color.FgHiBlack).SprintFunc(),
		Success: color.New(color.FgGreen).SprintFunc(),
		Warn:    color.New(color.FgYellow).SprintFunc(),
		Error:   color.New(color.FgRed).SprintFunc(),
		Bold:    color.New(color.Bold).SprintFunc(),
	}
}
