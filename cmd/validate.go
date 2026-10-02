package cmd

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/tumarsal/lefthook/v2/lefthook"
)

func validate() *cli.Command {
	var args lefthook.ValidateArgs
	var verbose bool

	return &cli.Command{
		Name:  "validate",
		Usage: "validate lefthook config",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return Validate(ctx, args, verbose)
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			shellCompleteFlags(cmd)
		},
	}
}
