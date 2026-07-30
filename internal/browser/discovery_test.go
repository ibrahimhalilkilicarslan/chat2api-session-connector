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

func TestBrowserNameForProgID(t *testing.T) {
	tests := map[string]string{
		"ChromeHTML":              "Google Chrome",
		"MSEdgeHTM":               "Microsoft Edge",
		"MicrosoftEdge.Url.https": "Microsoft Edge",
		"BraveHTML":               "Brave",
		"ChromiumHTM":             "Chromium",
		"FirefoxURL-308046B0AF4A": "",
	}
	for programID, expected := range tests {
		if actual := browserNameForProgID(programID); actual != expected {
			t.Errorf("browserNameForProgID(%q) = %q, want %q", programID, actual, expected)
		}
	}
}

func TestPrioritizeCandidatesPreservesPreferredBrowserOrder(t *testing.T) {
	values := []candidate{
		{Name: "Google Chrome", Path: "chrome-system"},
		{Name: "Microsoft Edge", Path: "edge-system"},
		{Name: "Google Chrome", Path: "chrome-user"},
		{Name: "Brave", Path: "brave-system"},
	}
	actual := prioritizeCandidates(values, "Microsoft Edge")
	expectedPaths := []string{"edge-system", "chrome-system", "chrome-user", "brave-system"}
	for index, expectedPath := range expectedPaths {
		if actual[index].Path != expectedPath {
			t.Fatalf("candidate %d path = %q, want %q", index, actual[index].Path, expectedPath)
		}
	}
}
