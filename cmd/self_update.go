package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/tumarsal/lefthook/v2/lefthook"
)

func selfUpdate() *cli.Command {
	var args lefthook.SelfUpdateArgs

	return &cli.Command{
		Name:  "self-update",
		Usage: "update lefthook executable",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "yes",
				Aliases:     []string{"y"},
				Usage:       "do not prompt y/n",
				Destination: &args.Yes,
			},
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "force reinstall",
				Destination: &args.Force,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &args.Verbose,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return SelfUpdate(ctx, args)
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			shellCompleteFlags(cmd)
		},
	}
}
