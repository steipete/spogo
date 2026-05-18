package app

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/steipete/spogo/internal/output"
)

func isColorEnabled(format output.Format, noColor bool) bool {
	if format != output.FormatHuman {
		return false
	}
	if noColor {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "dumb" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd())
}

func newOutputWriter(settings Settings) (*output.Writer, error) {
	format := settings.Format
	if format == "" {
		format = output.FormatHuman
	}
	return output.New(output.Options{
		Format: format,
		Color:  isColorEnabled(format, settings.NoColor),
		Quiet:  settings.Quiet,
		Out:    settings.Out,
		Err:    settings.Err,
	})
}
