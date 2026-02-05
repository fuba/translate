package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDirUsesXDG(t *testing.T) {
	base := t.TempDir()
	if err := os.Setenv("XDG_CONFIG_HOME", base); err != nil {
		t.Fatalf("Setenv error: %v", err)
	}
	dir, err := ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir error: %v", err)
	}

	want := filepath.Join(base, "translate")
	if dir != want {
		t.Fatalf("ConfigDir got %q, want %q", dir, want)
	}
}

func TestDefaultPDFFontPathUsesXDG(t *testing.T) {
	base := t.TempDir()
	if err := os.Setenv("XDG_CONFIG_HOME", base); err != nil {
		t.Fatalf("Setenv error: %v", err)
	}
	path, err := DefaultPDFFontPath()
	if err != nil {
		t.Fatalf("DefaultPDFFontPath error: %v", err)
	}
	want := filepath.Join(base, "translate", "fonts", "LINESeedJP-Regular.ttf")
	if path != want {
		t.Fatalf("DefaultPDFFontPath got %q, want %q", path, want)
	}
}
