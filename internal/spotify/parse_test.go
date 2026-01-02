package spotify

import "testing"

func TestParseResource(t *testing.T) {
	res, err := ParseResource("spotify:track:abc")
	if err != nil || res.Type != "track" || res.ID != "abc" {
		t.Fatalf("unexpected: %#v %v", res, err)
	}
	res, err = ParseResource("https://open.spotify.com/album/xyz?si=123")
	if err != nil || res.Type != "album" || res.ID != "xyz" {
		t.Fatalf("unexpected: %#v %v", res, err)
	}
	res, err = ParseResource("open.spotify.com/artist/aa")
	if err != nil || res.Type != "artist" || res.ID != "aa" {
		t.Fatalf("unexpected: %#v %v", res, err)
	}
	res, err = ParseResource("rawid")
	if err != nil || res.ID != "rawid" || res.Type != "" {
		t.Fatalf("unexpected: %#v %v", res, err)
	}
}

func TestParseTypedID(t *testing.T) {
	res, err := ParseTypedID("abc", "track")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if res.URI != "spotify:track:abc" {
		t.Fatalf("uri: %s", res.URI)
	}
	if _, err := ParseTypedID("spotify:album:abc", "track"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestJoinComma(t *testing.T) {
	if joinComma([]string{"a", "b"}) != "a,b" {
		t.Fatalf("unexpected join")
	}
}

func TestIsContextURI(t *testing.T) {
	if !isContextURI("spotify:album:a1") {
		t.Fatalf("expected context uri")
	}
	if isContextURI("spotify:track:t1") {
		t.Fatalf("unexpected context uri")
	}
}
