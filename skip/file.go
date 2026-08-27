package skip

import (
	"path/filepath"

	"github.com/spf13/afero"
)

type fileCondition struct {
	fs   afero.Fs
	root string
}

// File returns a Condition that matches file: keys in skip/only items.
//
// Relative paths are resolved from root. Absolute paths are used as-is.
func File(fs afero.Fs, root string) Condition {
	return fileCondition{fs: fs, root: root}
}

func (f fileCondition) Match(_ func() GitState, item map[string]any) bool {
	path, ok := item["file"].(string)
	if !ok || path == "" {
		return false
	}

	if f.fs == nil {
		return false
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(f.root, path)
	}

	exists, err := afero.Exists(f.fs, path)
	return err == nil && exists
}
