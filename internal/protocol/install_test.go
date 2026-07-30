package protocol

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInstallCurrentExecutableCreatesIndependentCopy(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the running Windows test binary cannot be replaced consistently")
	}
	destination := filepath.Join(t.TempDir(), "installed", "connector")
	if err := installCurrentExecutable(destination); err != nil {
		t.Fatalf("installCurrentExecutable() error = %v", err)
	}
	current, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	sourceBytes, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	installedBytes, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sourceBytes, installedBytes) {
		t.Fatal("installed executable differs from the running binary")
	}
	info, err := os.Stat(destination)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("installed executable permissions = %o", info.Mode().Perm())
	}
}
