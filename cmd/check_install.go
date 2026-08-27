package cmd

import (
	"context"
	"errors"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/v2/lefthook"
)

func checkInstall() *cli.Command {
	var verbose bool
	return &cli.Command{
		Name:  "check-install",
		Usage: "check if hooks are installed",
		UsageText: `lefthook check-install – Check if lefthook is installed. Exit codes:
0 – hooks are installed
1 – hooks are not installed or stale`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			err := CheckInstall(ctx, verbose)
			if err == nil {
				os.Exit(0)
			}
			if errors.Is(err, lefthook.ErrNotInstalled) {
				os.Exit(1)
			}
			return err
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			shellCompleteFlags(cmd)
		},
	}
}
