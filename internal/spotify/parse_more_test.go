package spotify

import "testing"

func TestParseResourceErrors(t *testing.T) {
	if _, err := ParseResource(""); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := ParseResource("spotify:badtype:123"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := ParseResource("https://open.spotify.com/"); err == nil {
		t.Fatalf("expected error")
	}
	if _, err := ParseResource("spotify:track"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseTypedIDNoExpectedType(t *testing.T) {
	res, err := ParseTypedID("spotify:track:t1", "")
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if res.Type != "track" || res.ID != "t1" {
		t.Fatalf("unexpected: %#v", res)
	}
}
