package protocol

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func installCurrentExecutable(destination string) error {
	current, err := os.Executable()
	if err != nil {
		return err
	}
	current, err = filepath.Abs(current)
	if err != nil {
		return err
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return err
	}
	if strings.EqualFold(current, destination) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}

	source, err := os.Open(current)
	if err != nil {
		return err
	}
	defer source.Close()

	temporary := destination + ".new"
	target, err := os.OpenFile(temporary, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o700)
	if err != nil {
		return err
	}
	_, copyError := io.Copy(target, source)
	closeError := target.Close()
	if copyError != nil {
		_ = os.Remove(temporary)
		return copyError
	}
	if closeError != nil {
		_ = os.Remove(temporary)
		return closeError
	}
	if err := os.Chmod(temporary, 0o700); err != nil {
		_ = os.Remove(temporary)
		return err
	}
	backup := destination + ".old"
	_ = os.Remove(backup)
	hadDestination := false
	if err := os.Rename(destination, backup); err == nil {
		hadDestination = true
	} else if !os.IsNotExist(err) {
		_ = os.Remove(temporary)
		return err
	}
	if err := os.Rename(temporary, destination); err != nil {
		if hadDestination {
			_ = os.Rename(backup, destination)
		}
		_ = os.Remove(temporary)
		return err
	}
	if hadDestination {
		_ = os.Remove(backup)
	}
	return nil
}
