package lefthook

import "github.com/tumarsal/lefthook/v2/internal/command"

// Args types mirror CLI command flags for programmatic use.

type (
	RunArgs       = command.RunArgs
	InstallArgs   = command.InstallArgs
	UninstallArgs = command.UninstallArgs
	DumpArgs      = command.DumpArgs
	AddArgs       = command.AddArgs
	ValidateArgs  = command.ValidateArgs
)

// SelfUpdateArgs configures the self-update command.
type SelfUpdateArgs struct {
	Yes     bool
	Force   bool
	Verbose bool
	// ExePath overrides the binary to update. Defaults to os.Executable().
	ExePath string
}
