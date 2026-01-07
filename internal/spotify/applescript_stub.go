//go:build !darwin
// +build !darwin

package spotify

import (
	"errors"
)

type AppleScriptClient struct{}

type AppleScriptOptions struct {
	Fallback API
}

func NewAppleScriptClient(opts AppleScriptOptions) (API, error) {
	return nil, errors.New("applescript engine is only available on macOS")
}
