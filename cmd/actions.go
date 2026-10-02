package cmd

import (
	"context"

	"github.com/tumarsal/lefthook/v2/lefthook"
)

// Run executes a hook (lefthook run).
func Run(ctx context.Context, args lefthook.RunArgs, colors string) error {
	app, err := lefthook.New(
		lefthook.WithVerbose(args.Verbose),
		lefthook.WithColors(colors),
	)
	if err != nil {
		return err
	}

	return app.RunWithArgs(ctx, args)
}

// Install installs Git hooks (lefthook install).
func Install(ctx context.Context, args lefthook.InstallArgs, verbose bool, hooks []string) error {
	app, err := lefthook.New(lefthook.WithVerbose(verbose))
	if err != nil {
		return err
	}

	return app.Install(ctx, args, hooks...)
}

// Uninstall removes Git hooks (lefthook uninstall).
func Uninstall(ctx context.Context, args lefthook.UninstallArgs, verbose bool) error {
	app, err := lefthook.New(lefthook.WithVerbose(verbose))
	if err != nil {
		return err
	}

	return app.Uninstall(ctx, args)
}

// CheckInstall verifies hooks are installed (lefthook check-install).
func CheckInstall(ctx context.Context, verbose bool) error {
	app, err := lefthook.New(lefthook.WithVerbose(verbose))
	if err != nil {
		return err
	}

	return app.CheckInstall(ctx)
}

// Dump prints merged config (lefthook dump).
func Dump(ctx context.Context, args lefthook.DumpArgs) error {
	app, err := lefthook.New(lefthook.WithColors("no"))
	if err != nil {
		return err
	}

	return app.Dump(ctx, args)
}

// Add creates hook scripts directory (lefthook add).
func Add(ctx context.Context, args lefthook.AddArgs, verbose bool) error {
	app, err := lefthook.New(lefthook.WithVerbose(verbose))
	if err != nil {
		return err
	}

	return app.Add(ctx, args)
}

// Validate validates config (lefthook validate).
func Validate(ctx context.Context, args lefthook.ValidateArgs, verbose bool) error {
	app, err := lefthook.New(lefthook.WithVerbose(verbose))
	if err != nil {
		return err
	}

	return app.Validate(ctx, args)
}

// SelfUpdate updates lefthook binary (lefthook self-update).
func SelfUpdate(ctx context.Context, args lefthook.SelfUpdateArgs) error {
	return lefthook.SelfUpdate(ctx, args)
}
