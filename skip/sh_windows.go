//go:build windows

package skip

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

const (
	shName        = "sh"
	defaultShPath = `C:\Program Files\Git\bin\sh.exe`
)

var fullShPath = sync.OnceValues(func() (string, error) {
	if _, err := os.Stat(defaultShPath); err == nil {
		return defaultShPath, nil
	}

	shPath, _ := exec.LookPath("sh")
	if len(shPath) > 0 {
		return shPath, nil
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		return "", err
	}

	shPath = filepath.Join(gitPath, "..", "..", "bin", "sh.exe")
	if _, err := os.Stat(shPath); err != nil {
		return "", err
	}

	return shPath, nil
})

func shExecutable() (string, error) {
	if len(os.Getenv("GIT_INDEX_FILE")) != 0 {
		return shName, nil
	}

	shPath, err := fullShPath()
	if err != nil {
		return "", fmt.Errorf("`sh` lookup failed: %w", err)
	}

	return shPath, nil
}
