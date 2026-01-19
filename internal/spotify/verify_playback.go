package spotify

import (
	"context"
	"fmt"
	"time"
)

type PlaybackVerifyError struct {
	Timeout    time.Duration
	Interval   time.Duration
	LastStatus PlaybackStatus
	LastErr    error
}

func (e *PlaybackVerifyError) Error() string {
	if e == nil {
		return "playback verification failed"
	}
	missing := ""
	if e.LastStatus.Device.ID == "" {
		missing += "device "
	}
	if e.LastStatus.Item == nil || e.LastStatus.Item.URI == "" {
		missing += "item "
	}
	switch {
	case missing != "" && e.LastStatus.ProgressMS == 0:
		return fmt.Sprintf("playback verification failed after %s (missing %sand progress_ms stayed at 0)", e.Timeout, missing)
	case missing != "":
		return fmt.Sprintf("playback verification failed after %s (missing %s)", e.Timeout, missing)
	case e.LastStatus.ProgressMS == 0:
		return fmt.Sprintf("playback verification failed after %s (progress_ms stayed at 0)", e.Timeout)
	default:
		return fmt.Sprintf("playback verification failed after %s", e.Timeout)
	}
}

func (e *PlaybackVerifyError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.LastErr
}

func VerifyPlayback(ctx context.Context, client interface {
	Playback(context.Context) (PlaybackStatus, error)
}, timeout time.Duration, interval time.Duration) (PlaybackStatus, error) {
	if timeout <= 0 {
		return PlaybackStatus{}, nil
	}
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}

	verifyCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastStatus PlaybackStatus
	var lastErr error
	for {
		status, err := client.Playback(verifyCtx)
		if err == nil && status.ProgressMS > 0 && status.Device.ID != "" && status.Item != nil && status.Item.URI != "" {
			return status, nil
		}
		lastStatus = status
		lastErr = err

		select {
		case <-verifyCtx.Done():
			return lastStatus, &PlaybackVerifyError{
				Timeout:    timeout,
				Interval:   interval,
				LastStatus: lastStatus,
				LastErr:    lastErr,
			}
		case <-ticker.C:
		}
	}
}
