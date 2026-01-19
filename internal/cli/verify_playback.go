package cli

import (
	"context"
	"time"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/spotify"
)

const verifyPlaybackPollInterval = 250 * time.Millisecond

func verifyPlaybackAfterAction(ctx *app.Context, client spotify.API, action string) error {
	if ctx == nil || ctx.Settings.VerifyPlayback <= 0 {
		return nil
	}
	_, err := spotify.VerifyPlayback(context.Background(), client, ctx.Settings.VerifyPlayback, verifyPlaybackPollInterval)
	if err == nil {
		return nil
	}
	if ctx.Settings.VerifyPlaybackFail {
		return app.WrapExit(5, err)
	}
	ctx.Output.Warnf("warning: %s: %v", action, err)
	return nil
}
