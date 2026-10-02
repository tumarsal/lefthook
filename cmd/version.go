package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/tumarsal/lefthook/v2/internal/command"
	"github.com/tumarsal/lefthook/v2/internal/logger"
	"github.com/tumarsal/lefthook/v2/lefthook"
)

func version() *cli.Command {
	var verbose bool

	return &cli.Command{
		Name:  "version",
		Usage: "print version",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &verbose,
			},
			&cli.BoolFlag{
				Name:        "full",
				Aliases:     []string{"f"},
				Destination: &verbose,
			},
		},
		Action: func(_ctx context.Context, _ *cli.Command) error {
			logger.New(os.Stdout).Info(PrintVersion(verbose))
			return nil
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			shellCompleteFlags(cmd)
		},
	}
}

// PrintVersion returns the lefthook version string for CLI output.
func PrintVersion(verbose bool) string {
	return lefthook.Version(verbose)
}

// shellCompleteHookNames and shellCompleteFlags delegate to internal/command for CLI completion.
func shellCompleteHookNames() {
	command.ShellCompleteHookNames()
}

func shellCompleteFlags(cmd *cli.Command) {
	command.ShellCompleteFlags(cmd)
}
