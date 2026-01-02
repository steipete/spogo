package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load default: %v", err)
	}
	if cfg.DefaultProfile != DefaultProfile {
		t.Fatalf("default profile = %q", cfg.DefaultProfile)
	}
}

func TestDefaultPath(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	path, err := DefaultPath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}
	if filepath.Base(path) != DefaultConfig {
		t.Fatalf("unexpected path: %s", path)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	cfg := Default()
	cfg.DefaultProfile = "work"
	cfg.SetProfile("work", Profile{Browser: "chrome", Market: "US"})
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	profile := loaded.Profile("work")
	if profile.Browser != "chrome" || profile.Market != "US" {
		t.Fatalf("profile mismatch: %#v", profile)
	}
}

func TestCookiePath(t *testing.T) {
	path := CookiePath("/tmp/spogo/config.toml", "default")
	if filepath.Base(path) != "default.json" {
		t.Fatalf("cookie path: %s", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

func TestCookiePathEmptyConfig(t *testing.T) {
	if CookiePath("", "default") != "" {
		t.Fatalf("expected empty")
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("not=toml=\""), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := Load(path); err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadReadError(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSaveNilConfig(t *testing.T) {
	if err := Save("", nil); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSaveDefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg := Default()
	if err := Save("", cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
}

func TestSaveInvalidDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	path := filepath.Join(file, "config.toml")
	if err := Save(path, Default()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestProfileNilConfig(t *testing.T) {
	var cfg *Config
	if p := cfg.Profile("default"); p != (Profile{}) {
		t.Fatalf("expected empty profile")
	}
}

func TestSetProfileDefaultName(t *testing.T) {
	cfg := Default()
	cfg.SetProfile("", Profile{Market: "US"})
	if cfg.Profile(DefaultProfile).Market != "US" {
		t.Fatalf("expected profile set")
	}
}

func TestSetProfileNilMap(t *testing.T) {
	cfg := &Config{}
	cfg.SetProfile("", Profile{Market: "DE"})
	if cfg.Profiles == nil {
		t.Fatalf("expected profiles map")
	}
	if cfg.Profile(DefaultProfile).Market != "DE" {
		t.Fatalf("expected profile")
	}
}

func TestSetProfileNilConfig(t *testing.T) {
	var cfg *Config
	cfg.SetProfile("default", Profile{Market: "US"})
}

func TestProfileNilMap(t *testing.T) {
	cfg := &Config{DefaultProfile: DefaultProfile}
	if cfg.Profile("default") != (Profile{}) {
		t.Fatalf("expected empty profile")
	}
}

func TestProfileFallback(t *testing.T) {
	cfg := Default()
	cfg.DefaultProfile = "primary"
	cfg.SetProfile("primary", Profile{Market: "DE"})
	if cfg.Profile("").Market != "DE" {
		t.Fatalf("expected default profile")
	}
}

func TestNormalize(t *testing.T) {
	cfg := &Config{}
	cfg.normalize()
	if cfg.DefaultProfile == "" || cfg.Profiles == nil {
		t.Fatalf("expected defaults")
	}
}
