package browser

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindHonorsExplicitBrowserPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable permission semantics differ on Windows")
	}
	path := filepath.Join(t.TempDir(), "browser")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CHAT2API_BROWSER_PATH", path)
	found, err := Find()
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	if found.Path != path || found.Name != "Custom Chromium" {
		t.Fatalf("Find() = %#v", found)
	}
}

func TestFindRejectsInvalidExplicitBrowserPath(t *testing.T) {
	t.Setenv("CHAT2API_BROWSER_PATH", filepath.Join(t.TempDir(), "missing"))
	if _, err := Find(); err == nil {
		t.Fatal("Find() unexpectedly succeeded")
	}
}
