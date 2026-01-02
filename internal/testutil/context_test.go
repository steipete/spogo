package testutil

import (
	"testing"

	"github.com/steipete/spogo/internal/output"
)

func TestNewTestContext(t *testing.T) {
	ctx, out, errOut := NewTestContext(t, output.FormatPlain)
	if ctx == nil || out == nil || errOut == nil {
		t.Fatalf("expected context")
	}
	_ = ctx.Output.Theme.Accent("hi")
}
