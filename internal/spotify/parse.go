package spotify

import (
	"errors"
	"net/url"
	"path"
	"strings"
)

var ErrUnsupportedType = errors.New("unsupported spotify type")

var supportedTypes = map[string]struct{}{
	"track":    {},
	"album":    {},
	"artist":   {},
	"playlist": {},
	"show":     {},
	"episode":  {},
}

type Resource struct {
	Type string
	ID   string
	URI  string
}

func ParseResource(input string) (Resource, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Resource{}, errors.New("empty input")
	}
	if strings.HasPrefix(input, "spotify:") {
		parts := strings.Split(input, ":")
		if len(parts) < 3 {
			return Resource{}, errors.New("invalid spotify uri")
		}
		kind := parts[1]
		id := parts[2]
		if !isSupportedType(kind) {
			return Resource{}, ErrUnsupportedType
		}
		return Resource{Type: kind, ID: id, URI: "spotify:" + kind + ":" + id}, nil
	}
	if strings.HasPrefix(input, "open.spotify.com/") {
		input = "https://" + input
	}
	if strings.Contains(input, "open.spotify.com/") {
		parsed, err := url.Parse(input)
		if err != nil {
			return Resource{}, err
		}
		segments := strings.Split(strings.Trim(path.Clean(parsed.Path), "/"), "/")
		if len(segments) < 2 {
			return Resource{}, errors.New("invalid spotify url")
		}
		kind := segments[0]
		id := segments[1]
		if !isSupportedType(kind) {
			return Resource{}, ErrUnsupportedType
		}
		return Resource{Type: kind, ID: id, URI: "spotify:" + kind + ":" + id}, nil
	}
	return Resource{ID: input}, nil
}

func ParseTypedID(input, expectedType string) (Resource, error) {
	res, err := ParseResource(input)
	if err != nil {
		return Resource{}, err
	}
	if expectedType == "" {
		return res, nil
	}
	if res.Type == "" {
		res.Type = expectedType
		res.URI = "spotify:" + expectedType + ":" + res.ID
		return res, nil
	}
	if res.Type != expectedType {
		return Resource{}, errors.New("unexpected spotify type")
	}
	return res, nil
}

func isSupportedType(kind string) bool {
	_, ok := supportedTypes[kind]
	return ok
}

func isContextURI(uri string) bool {
	return strings.Contains(uri, ":album:") || strings.Contains(uri, ":playlist:") || strings.Contains(uri, ":show:")
}
