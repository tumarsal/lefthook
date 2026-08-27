// Package skip provides extensible skip/only condition evaluation for lefthook hooks.
//
// Use NewSkipChecker to create a Checker with built-in ref and run conditions.
// Register additional Condition implementations via WithCondition to support custom
// YAML keys in skip/only arrays.
package skip

import "io"

// GitState describes the current git branch and repository state.
type GitState struct {
	Branch string
	State  string // merge | rebase | merge-commit | ""
}

// Checker evaluates skip/only settings from hook configuration.
type Checker interface {
	// Check returns true when execution should be skipped.
	Check(state func() GitState, skip, only any) bool
}

// Condition matches a single map element from a skip/only array.
//
// Return true when the condition key is present in item and the check succeeds.
// Return false when the key is absent or the check fails.
type Condition interface {
	Match(state func() GitState, item map[string]any) bool
}

// Command runs shell commands for run: skip conditions.
type Command interface {
	Run(cmd []string, root string, in io.Reader, out, errOut io.Writer) error
}

// Logger receives diagnostic messages from skip evaluation.
type Logger interface {
	Errorf(format string, args ...any)
}

// Option configures NewSkipChecker.
type Option func(*config)

type config struct {
	cmd        Command
	logger     Logger
	conditions []Condition
}

// NewSkipChecker returns a Checker with built-in ref and run conditions.
// Additional Condition values can be registered via WithCondition.
func NewSkipChecker(opts ...Option) Checker {
	cfg := &config{}

	for _, opt := range opts {
		opt(cfg)
	}

	const builtinConditions = 2

	conditions := make([]Condition, 0, builtinConditions+len(cfg.conditions))
	conditions = append(conditions,
		refCondition{},
		runCondition{cmd: cfg.cmd, logger: cfg.logger},
	)
	conditions = append(conditions, cfg.conditions...)

	return &checker{conditions: conditions}
}

// WithCommand sets the command runner used for run: conditions.
func WithCommand(cmd Command) Option {
	return func(c *config) {
		c.cmd = cmd
	}
}

// WithLogger sets the logger used for run: condition diagnostics.
func WithLogger(l Logger) Option {
	return func(c *config) {
		c.logger = l
	}
}

// WithCondition appends a custom skip/only condition.
func WithCondition(c Condition) Option {
	return func(cfg *config) {
		cfg.conditions = append(cfg.conditions, c)
	}
}

// WithConditions appends multiple custom skip/only conditions.
func WithConditions(cs ...Condition) Option {
	return func(cfg *config) {
		cfg.conditions = append(cfg.conditions, cs...)
	}
}
