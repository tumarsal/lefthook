package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHook_WithLefthookPath(t *testing.T) {
	content := string(Hook("pre-commit", Args{
		LefthookPath: "tira my git lefthook",
	}))

	assert.Contains(t, content, "tira my git lefthook \"$@\"")
	assert.Contains(t, content, `call_lefthook run "pre-commit"`)
	assert.NotContains(t, content, "LEFTHOOK_BIN")
	assert.NotContains(t, content, `test -n "tira my git lefthook"`)
}

func TestHook_WithoutLefthookPath(t *testing.T) {
	content := string(Hook("pre-commit", Args{}))

	require.Contains(t, content, `if test -n "$LEFTHOOK_BIN"`)
	assert.NotContains(t, content, "tira my git lefthook")
}
