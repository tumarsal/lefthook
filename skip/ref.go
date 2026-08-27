package skip

import "github.com/gobwas/glob"

type refCondition struct{}

func (refCondition) Match(state func() GitState, item map[string]any) bool {
	ref, ok := item["ref"].(string)
	if !ok {
		return false
	}

	branch := state().Branch
	if ref == branch {
		return true
	}

	g := glob.MustCompile(ref)

	return g.Match(branch)
}
