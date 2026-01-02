package testutil

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
)

func NewTestContext(t *testing.T, format output.Format) (*app.Context, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	ctx := &app.Context{
		Settings: app.Settings{Format: format},
		Output: &output.Writer{
			Format: format,
			Out:    out,
			Err:    errOut,
			Color:  false,
			Theme:  output.Theme{Accent: sprint, Muted: sprint, Success: sprint, Warn: sprint, Error: sprint, Bold: sprint},
		},
	}
	return ctx, out, errOut
}

func sprint(a ...any) string {
	return fmt.Sprint(a...)
}
