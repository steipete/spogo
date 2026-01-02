package cli

import (
	"context"
	"fmt"

	"github.com/steipete/spogo/internal/app"
	"github.com/steipete/spogo/internal/output"
	"github.com/steipete/spogo/internal/spotify"
)

type QueueCmd struct {
	Add   QueueAddCmd   `kong:"cmd,help='Add to queue.'"`
	Show  QueueShowCmd  `kong:"cmd,help='Show queue.'"`
	Clear QueueClearCmd `kong:"cmd,help='Clear queue (if supported).'"`
}

type QueueAddCmd struct {
	Item string `arg:"" required:"" help:"Track URI/URL/ID."`
}

type QueueShowCmd struct{}

type QueueClearCmd struct{}

func (cmd *QueueAddCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	res, err := spotify.ParseTypedID(cmd.Item, "track")
	if err != nil {
		return err
	}
	if res.URI == "" {
		return fmt.Errorf("invalid track")
	}
	if err := client.QueueAdd(context.Background(), res.URI); err != nil {
		return err
	}
	return ctx.Output.Emit(map[string]string{"status": "ok"}, []string{"ok"}, []string{"Queued"})
}

func (cmd *QueueShowCmd) Run(ctx *app.Context) error {
	client, err := ctx.Spotify()
	if err != nil {
		return err
	}
	queue, err := client.Queue(context.Background())
	if err != nil {
		return err
	}
	plain, human := renderItems(ctx.Output, queue.Queue)
	if queue.CurrentlyPlaying != nil {
		current := itemHuman(ctx.Output, *queue.CurrentlyPlaying)
		if ctx.Output.Format == output.FormatHuman {
			human = append([]string{"Now playing: " + current, "Queue:"}, human...)
		}
	}
	return ctx.Output.Emit(queue, plain, human)
}

func (cmd *QueueClearCmd) Run(ctx *app.Context) error {
	return fmt.Errorf("queue clear not supported by Spotify API")
}
