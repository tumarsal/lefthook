package cmd

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"

	"github.com/evilmartians/lefthook/v2/lefthook"
)

var errInvalidFormat = errors.New("invalid 'format' value, supported: 'toml', 'yaml', 'json'")

func dump() *cli.Command {
	args := lefthook.DumpArgs{
		Format: "yaml",
	}

	return &cli.Command{
		Name:  "dump",
		Usage: "print config merged from all extensions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "format",
				Usage:       "'yaml', 'toml', or 'json' (default: 'yaml')",
				Aliases:     []string{"f"},
				Destination: &args.Format,
				Validator: func(format string) error {
					switch format {
					case "":
					case "yaml":
					case "toml":
					case "json":
					default:
						return errInvalidFormat
					}
					return nil
				},
			},
		},
		Action: func(ctx context.Context, _ *cli.Command) error {
			return Dump(ctx, args)
		},
		ShellComplete: func(ctx context.Context, cmd *cli.Command) {
			shellCompleteFlags(cmd)
		},
	}
}
