package lefthook

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/tumarsal/lefthook/v2/internal/command"
	"github.com/tumarsal/lefthook/v2/internal/logger"
	"github.com/tumarsal/lefthook/v2/internal/updater"
	ver "github.com/tumarsal/lefthook/v2/internal/version"
)

// ErrNotInstalled is returned by CheckInstall when hooks are missing or stale.
var ErrNotInstalled = errors.New("hooks are not installed or stale")

// RunWithArgs executes a hook with full run options (equivalent to `lefthook run`).
func (a *App) RunWithArgs(ctx context.Context, args RunArgs) error {
	return a.inner.Run(ctx, args)
}

// Install installs Git hooks from config (equivalent to `lefthook install`).
func (a *App) Install(ctx context.Context, args InstallArgs, hooks ...string) error {
	return a.inner.Install(ctx, args, hooks)
}

// Uninstall removes installed hooks (equivalent to `lefthook uninstall`).
func (a *App) Uninstall(ctx context.Context, args UninstallArgs) error {
	return a.inner.Uninstall(ctx, args)
}

// CheckInstall verifies hooks are installed and synchronized.
// Returns ErrNotInstalled when hooks are missing or stale.
func (a *App) CheckInstall(ctx context.Context) error {
	ok, err := a.inner.CheckInstallStatus(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotInstalled
	}
	return nil
}

// Dump prints merged config to stdout (equivalent to `lefthook dump`).
func (a *App) Dump(ctx context.Context, args DumpArgs) error {
	return a.inner.Dump(ctx, args)
}

// Add creates scripts directory and installs the hook (equivalent to `lefthook add`).
func (a *App) Add(ctx context.Context, args AddArgs) error {
	return a.inner.Add(ctx, args)
}

// Validate validates lefthook config against JSON schema (equivalent to `lefthook validate`).
func (a *App) Validate(ctx context.Context, args ValidateArgs) error {
	return a.inner.Validate(ctx, args)
}

// Version returns the lefthook version string (equivalent to `lefthook version`).
func Version(verbose bool) string {
	return ver.Version(verbose)
}

// SelfUpdate updates the lefthook executable (equivalent to `lefthook self-update`).
func SelfUpdate(ctx context.Context, args SelfUpdateArgs) error {
	l := logger.New(os.Stdout)
	if os.Getenv(command.EnvVerbose) == "1" || os.Getenv(command.EnvVerbose) == "true" {
		args.Verbose = true
	}
	if args.Verbose {
		l.SetLevel(logger.LevelDebug)
		l.Debug("Verbose mode enabled")
	}

	exePath := args.ExePath
	if exePath == "" {
		var err error
		exePath, err = os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine the binary path: %w", err)
		}
	}

	return updater.New(l).SelfUpdate(ctx, updater.Options{
		Yes:     args.Yes,
		Force:   args.Force,
		ExePath: exePath,
	})
}
