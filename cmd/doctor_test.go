package cmd

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestCheckYTDLPInstalled(t *testing.T) {
	check := checkYTDLP(func(name string) (string, error) {
		if name != "yt-dlp" {
			t.Fatalf("unexpected binary lookup %q", name)
		}
		return "/usr/bin/yt-dlp", nil
	})

	if !check.OK {
		t.Fatalf("expected yt-dlp check to pass: %+v", check)
	}
	if check.Detail != "/usr/bin/yt-dlp" {
		t.Fatalf("detail = %q, want binary path", check.Detail)
	}
}

func TestCheckYTDLPMissing(t *testing.T) {
	check := checkYTDLP(func(string) (string, error) {
		return "", errors.New("not found")
	})

	if check.OK {
		t.Fatalf("expected yt-dlp check to fail")
	}
	if check.Fix == "" {
		t.Fatalf("expected install guidance")
	}
}

func TestCheckClipboardToolsWayland(t *testing.T) {
	have := map[string]bool{"wl-copy": true, "wl-paste": true}
	check := checkClipboardTools("linux", func(key string) string {
		if key == "WAYLAND_DISPLAY" {
			return "wayland-1"
		}
		return ""
	}, fakeLookPath(have))

	if !check.OK {
		t.Fatalf("expected Wayland clipboard check to pass: %+v", check)
	}
}

func TestCheckClipboardToolsMissing(t *testing.T) {
	check := checkClipboardTools("linux", func(string) string { return "" }, fakeLookPath(nil))

	if check.OK {
		t.Fatalf("expected missing clipboard tools to fail")
	}
	if check.Fix == "" {
		t.Fatalf("expected install guidance")
	}
}

func TestCheckDirectoryWritable(t *testing.T) {
	dir := t.TempDir()
	check := checkDirectoryWritable("test directory", dir)

	if !check.OK {
		t.Fatalf("expected writable directory check to pass: %+v", check)
	}
	if filepath.Clean(dir) == "" {
		t.Fatalf("unexpected empty temp dir")
	}
}

func fakeLookPath(commands map[string]bool) func(string) (string, error) {
	return func(name string) (string, error) {
		if commands[name] {
			return "/usr/bin/" + name, nil
		}
		return "", errors.New("not found")
	}
}
