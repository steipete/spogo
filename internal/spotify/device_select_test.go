package spotify

import "testing"

func TestResolveDeviceIDExactID(t *testing.T) {
	devices := []Device{
		{ID: "d1", Name: "Desk"},
		{ID: "d2", Name: "Phone"},
	}
	got, err := ResolveDeviceID(devices, "d2")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "d2" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveDeviceIDExactName(t *testing.T) {
	devices := []Device{
		{ID: "d1", Name: "Desk"},
		{ID: "d2", Name: "Phone"},
	}
	got, err := ResolveDeviceID(devices, "desk")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "d1" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveDeviceIDUniqueSubstringName(t *testing.T) {
	devices := []Device{
		{ID: "d1", Name: "Desk Mac"},
		{ID: "d2", Name: "Phone"},
	}
	got, err := ResolveDeviceID(devices, "mac")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "d1" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveDeviceIDAmbiguousSubstringName(t *testing.T) {
	devices := []Device{
		{ID: "d1", Name: "Office Desk"},
		{ID: "d2", Name: "Home Desk"},
	}
	if _, err := ResolveDeviceID(devices, "desk"); err == nil {
		t.Fatalf("expected error")
	}
}
