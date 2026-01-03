package spotify

import "testing"

func TestParseMapLiteral(t *testing.T) {
	raw := `{1:"foo",2:"bar"}`
	m, err := parseMapLiteral(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m[1] != "foo" || m[2] != "bar" {
		t.Fatalf("unexpected map: %#v", m)
	}
}

func TestParseWebpackMaps(t *testing.T) {
	js := `var a={1:"alpha-beta",2:"beta-gamma"};var b={1:"a1b2c3d4",2:"e5f6a7b8"};`
	nameMap, hashMap, err := parseWebpackMaps(js)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if nameMap[1] != "alpha-beta" && nameMap[2] != "beta-gamma" {
		t.Fatalf("unexpected name map")
	}
	if hashMap[1] != "a1b2c3d4" && hashMap[2] != "e5f6a7b8" {
		t.Fatalf("unexpected hash map")
	}
}

func TestFindOperationHashes(t *testing.T) {
	body := `foo searchDesktop bar sha256Hash":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"`
	found := findOperationHashes(body, []string{"searchDesktop"})
	if found["searchDesktop"] == "" {
		t.Fatalf("expected hash")
	}
}
