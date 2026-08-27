package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatHookCommand(t *testing.T) {
	assert.Equal(t, "", formatHookCommand(nil))
	assert.Equal(t, "", formatHookCommand([]string{}))
	assert.Equal(t, "", formatHookCommand([]string{"", ""}))
	assert.Equal(t, "tira my git lefthook", formatHookCommand([]string{"tira", "my", "git", "lefthook"}))
	assert.Equal(t, "'path with spaces' lefthook", formatHookCommand([]string{"path with spaces", "lefthook"}))
}

func TestResolveLefthookPath(t *testing.T) {
	l := &Lefthook{hookCommand: []string{"from", "app"}}

	assert.Equal(t, "from args", l.resolveLefthookPath("from-yaml", []string{"from", "args"}))
	assert.Equal(t, "from app", l.resolveLefthookPath("from-yaml", nil))
	assert.Equal(t, "from-yaml", (&Lefthook{}).resolveLefthookPath("from-yaml", nil))
}
