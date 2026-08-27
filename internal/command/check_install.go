package command

import (
	"context"
	"os"
)

type installationStatus int

const (
	installed installationStatus = iota
	notInstalled
)

func (l *Lefthook) CheckInstall(_ctx context.Context) error {
	ok, err := l.CheckInstallStatus(_ctx)
	if err != nil {
		return err
	}

	if ok {
		os.Exit(0)
	}

	os.Exit(1)
	return nil
}

// CheckInstallStatus reports whether hooks are installed and synchronized.
func (l *Lefthook) CheckInstallStatus(_ctx context.Context) (bool, error) {
	check, err := l.checkInstall()
	if err != nil {
		return false, err
	}

	return check == installed, nil
}

func (l *Lefthook) checkInstall() (installationStatus, error) {
	if !l.configExists(l.repo.RootPath) {
		return notInstalled, nil
	}

	cfg, err := l.LoadConfig()
	if err != nil {
		return notInstalled, err
	}

	ok, _ := l.checkHooksSynchronized(cfg)
	if !ok {
		return notInstalled, nil
	}

	return installed, nil
}
