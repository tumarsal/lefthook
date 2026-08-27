package skip

import (
	"os"
	"strings"
)

type envCondition struct{}

// Env returns a Condition that matches env: keys in skip/only items.
//
// env: NAME matches when the variable is set and non-empty.
// env: NAME=value matches when the variable equals value exactly.
func Env() Condition {
	return envCondition{}
}

func (envCondition) Match(_ func() GitState, item map[string]any) bool {
	spec, ok := item["env"].(string)
	if !ok {
		return false
	}

	name, value, hasValue := strings.Cut(spec, "=")
	if name == "" {
		return false
	}

	actual, ok := os.LookupEnv(name)
	if !hasValue {
		return ok && actual != ""
	}

	return ok && actual == value
}
