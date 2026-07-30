//go:build !windows

package browser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestLaunchOptionsReachBrowserProcess(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "fake-browser")
	if err := os.WriteFile(
		executable,
		[]byte("#!/bin/sh\nprintf 'fake browser exited\\n' >&2\nexit 1\n"),
		0o700,
	); err != nil {
		t.Fatal(err)
	}

	rootContext, rootCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer rootCancel()
	allocatorContext, allocatorCancel := chromedp.NewExecAllocator(
		rootContext,
		launchOptions(Installed{Name: "Test Browser", Path: executable}, t.TempDir())...,
	)
	defer allocatorCancel()
	browserContext, browserCancel := chromedp.NewContext(allocatorContext)
	defer browserCancel()

	err := chromedp.Run(browserContext)
	if err == nil {
		t.Fatal("fake browser unexpectedly started")
	}
	if strings.Contains(err.Error(), "invalid exec pool flag") {
		t.Fatalf("chromedp rejected a connector launch flag: %v", err)
	}
	if !strings.Contains(err.Error(), "chrome failed to start") {
		t.Fatalf("browser process was not reached: %v", err)
	}
}
