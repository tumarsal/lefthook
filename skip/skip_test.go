package skip

import (
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type mockCmd struct{}

func (mockCmd) Run(cmd []string, _root string, _in io.Reader, _out io.Writer, _errOut io.Writer) error {
	if len(cmd) == 3 && cmd[2] == "success" {
		return nil
	}

	return errors.New("failure")
}

func TestChecker_Check(t *testing.T) {
	checker := NewSkipChecker(WithCommand(mockCmd{}))

	for _, tt := range [...]struct {
		name       string
		state      func() GitState
		skip, only any
		skipped    bool
	}{
		{
			name:    "when true",
			state:   func() GitState { return GitState{} },
			skip:    true,
			skipped: true,
		},
		{
			name:    "when false",
			state:   func() GitState { return GitState{} },
			skip:    false,
			skipped: false,
		},
		{
			name:    "when merge",
			state:   func() GitState { return GitState{State: "merge"} },
			skip:    "merge",
			skipped: true,
		},
		{
			name:    "when merge-commit",
			state:   func() GitState { return GitState{State: "merge-commit"} },
			skip:    "merge-commit",
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			state:   func() GitState { return GitState{State: "rebase"} },
			skip:    "merge",
			skipped: false,
		},
		{
			name:    "when rebase",
			state:   func() GitState { return GitState{State: "rebase"} },
			skip:    []any{"rebase"},
			skipped: true,
		},
		{
			name:    "when rebase (but want merge)",
			state:   func() GitState { return GitState{State: "rebase"} },
			skip:    []any{"merge"},
			skipped: false,
		},
		{
			name:    "when branch",
			state:   func() GitState { return GitState{Branch: "feat/skipme"} },
			skip:    []any{map[string]any{"ref": "feat/skipme"}},
			skipped: true,
		},
		{
			name:    "when branch doesn't match",
			state:   func() GitState { return GitState{Branch: "feat/important"} },
			skip:    []any{map[string]any{"ref": "feat/skipme"}},
			skipped: false,
		},
		{
			name:    "when branch glob",
			state:   func() GitState { return GitState{Branch: "feat/important"} },
			skip:    []any{map[string]any{"ref": "feat/*"}},
			skipped: true,
		},
		{
			name:    "when branch glob doesn't match",
			state:   func() GitState { return GitState{Branch: "feat"} },
			skip:    []any{map[string]any{"ref": "feat/*"}},
			skipped: false,
		},
		{
			name:    "when only specified",
			state:   func() GitState { return GitState{Branch: "feat"} },
			only:    []any{map[string]any{"ref": "feat"}},
			skipped: false,
		},
		{
			name:    "when only branch doesn't match",
			state:   func() GitState { return GitState{Branch: "dev"} },
			only:    []any{map[string]any{"ref": "feat"}},
			skipped: true,
		},
		{
			name:    "when only branch with glob",
			state:   func() GitState { return GitState{Branch: "feat/important"} },
			only:    []any{map[string]any{"ref": "feat/*"}},
			skipped: false,
		},
		{
			name:    "when only merge",
			state:   func() GitState { return GitState{State: "merge"} },
			only:    []any{"merge"},
			skipped: false,
		},
		{
			name:    "when only and skip",
			state:   func() GitState { return GitState{State: "rebase"} },
			skip:    []any{map[string]any{"ref": "feat/*"}},
			only:    "rebase",
			skipped: false,
		},
		{
			name:    "when only and skip applies skip",
			state:   func() GitState { return GitState{State: "rebase"} },
			skip:    []any{"rebase"},
			only:    "rebase",
			skipped: true,
		},
		{
			name:    "when skip with run command",
			state:   func() GitState { return GitState{} },
			skip:    []any{map[string]any{"run": "success"}},
			skipped: true,
		},
		{
			name:    "when skip with multi-run command",
			state:   func() GitState { return GitState{Branch: "feat"} },
			skip:    []any{map[string]any{"run": "success", "ref": "feat"}},
			skipped: true,
		},
		{
			name:    "when only with run command",
			state:   func() GitState { return GitState{} },
			only:    []any{map[string]any{"run": "fail"}},
			skipped: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.skipped, checker.Check(tt.state, tt.skip, tt.only))
		})
	}
}

type customKeyCondition struct {
	key string
}

func (c customKeyCondition) Match(_ func() GitState, item map[string]any) bool {
	value, ok := item[c.key].(string)
	return ok && value == "yes"
}

func TestChecker_WithCondition(t *testing.T) {
	checker := NewSkipChecker(WithCondition(customKeyCondition{key: "custom"}))

	state := func() GitState { return GitState{} }

	assert.True(t, checker.Check(state, []any{map[string]any{"custom": "yes"}}, nil))
	assert.False(t, checker.Check(state, []any{map[string]any{"custom": "no"}}, nil))
	assert.False(t, checker.Check(state, []any{map[string]any{"other": "yes"}}, nil))
}

func TestEnvCondition(t *testing.T) {
	checker := NewSkipChecker(WithCondition(Env()))
	state := func() GitState { return GitState{} }

	t.Setenv("LEFTHOOK_TEST_VAR", "1")
	assert.True(t, checker.Check(state, []any{map[string]any{"env": "LEFTHOOK_TEST_VAR"}}, nil))
	assert.True(t, checker.Check(state, []any{map[string]any{"env": "LEFTHOOK_TEST_VAR=1"}}, nil))
	assert.False(t, checker.Check(state, []any{map[string]any{"env": "LEFTHOOK_TEST_VAR=0"}}, nil))

	t.Setenv("LEFTHOOK_TEST_VAR", "")
	assert.False(t, checker.Check(state, []any{map[string]any{"env": "LEFTHOOK_TEST_VAR"}}, nil))
}

func TestFileCondition(t *testing.T) {
	fs := afero.NewMemMapFs()
	root := t.TempDir()
	path := filepath.Join(root, "lefthook.yml")
	assert.NoError(t, afero.WriteFile(fs, path, []byte("x"), 0o644))

	checker := NewSkipChecker(WithCondition(File(fs, root)))
	state := func() GitState { return GitState{} }

	assert.True(t, checker.Check(state, []any{map[string]any{"file": "lefthook.yml"}}, nil))
	assert.False(t, checker.Check(state, []any{map[string]any{"file": "missing.yml"}}, nil))
}

func TestBinCondition(t *testing.T) {
	checker := NewSkipChecker(WithCondition(Bin()))
	state := func() GitState { return GitState{} }

	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go executable not found in PATH")
	}

	assert.True(t, checker.Check(state, []any{map[string]any{"bin": "go"}}, nil))
	assert.False(t, checker.Check(state, []any{map[string]any{"bin": "definitely-not-a-real-binary-name"}}, nil))
}
