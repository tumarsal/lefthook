// Package lefthook provides a public API for embedding lefthook in Go programs.
//
// Example:
//
//	checker := skip.NewSkipChecker(
//	    skip.WithCommand(systemCmd),
//	    skip.WithCondition(skip.Env()),
//	    skip.WithCondition(skip.Bin()),
//	)
//
//	app, err := lefthook.New(
//	    lefthook.WithSkipChecker(checker),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := app.Run(context.Background(), "pre-commit"); err != nil {
//	    log.Fatal(err)
//	}
package lefthook

import (
	"context"

	"github.com/tumarsal/lefthook/v2/internal/command"
	"github.com/tumarsal/lefthook/v2/skip"
)

// Option configures App.
type Option func(*options)

type options struct {
	verbose     bool
	colors      string
	skipChecker skip.Checker
	hookCommand []string
}

// WithSkipChecker sets a custom skip/only condition checker for hook execution.
func WithSkipChecker(checker skip.Checker) Option {
	return func(o *options) {
		o.skipChecker = checker
	}
}

// WithHookCommand sets the argv written into Git hooks (e.g. "tira", "my", "git", "lefthook").
// After Install, hooks call: <cmd...> run <hook-name> ...
// Overrides config lefthook:; InstallArgs.HookCommand overrides this option.
func WithHookCommand(cmd ...string) Option {
	return func(o *options) {
		o.hookCommand = append([]string(nil), cmd...)
	}
}

// WithVerbose enables debug logging.
func WithVerbose(verbose bool) Option {
	return func(o *options) {
		o.verbose = verbose
	}
}

// WithColors sets color output mode: on, off, or auto.
func WithColors(colors string) Option {
	return func(o *options) {
		o.colors = colors
	}
}

// App runs lefthook hooks programmatically.
type App struct {
	inner *command.Lefthook
}

// New creates an App for running lefthook hooks.
func New(opts ...Option) (*App, error) {
	cfg := &options{colors: "auto"}
	for _, opt := range opts {
		opt(cfg)
	}

	var lhOpts []command.LefthookOption
	if cfg.skipChecker != nil {
		lhOpts = append(lhOpts, command.WithSkipChecker(cfg.skipChecker))
	}
	if len(cfg.hookCommand) > 0 {
		lhOpts = append(lhOpts, command.WithHookCommand(cfg.hookCommand...))
	}

	inner, err := command.NewLefthook(cfg.verbose, cfg.colors, lhOpts...)
	if err != nil {
		return nil, err
	}

	return &App{inner: inner}, nil
}

// Run executes a hook by name with optional git hook arguments.
// For full CLI parity use RunWithArgs.
func (a *App) Run(ctx context.Context, hook string, gitArgs ...string) error {
	return a.RunWithArgs(ctx, RunArgs{
		Hook:    hook,
		GitArgs: gitArgs,
	})
}
