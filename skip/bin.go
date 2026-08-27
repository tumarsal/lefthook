package skip

import "os/exec"

type binCondition struct {
	lookPath func(string) (string, error)
}

// Bin returns a Condition that matches bin: keys in skip/only items.
//
// A bin: name matches when the executable is found in PATH.
func Bin() Condition {
	return binCondition{lookPath: exec.LookPath}
}

func (b binCondition) Match(_ func() GitState, item map[string]any) bool {
	name, ok := item["bin"].(string)
	if !ok || name == "" {
		return false
	}

	lookPath := b.lookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	_, err := lookPath(name)
	return err == nil
}
