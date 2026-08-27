package skip

type checker struct {
	conditions []Condition
}

// Check returns the result of applying a skip/only setting which can be a branch, git state, shell command, etc.
func (c *checker) Check(state func() GitState, skip any, only any) bool {
	if skip == nil && only == nil {
		return false
	}

	if skip != nil {
		if c.matches(state, skip) {
			return true
		}
	}

	if only != nil {
		return !c.matches(state, only)
	}

	return false
}

func (c *checker) matches(state func() GitState, value any) bool {
	switch typedValue := value.(type) {
	case bool:
		return typedValue
	case string:
		return typedValue == state().State
	case []any:
		return c.matchesSlices(state, typedValue)
	}

	return false
}

func (c *checker) matchesSlices(gitState func() GitState, slice []any) bool {
	for _, item := range slice {
		switch typedItem := item.(type) {
		case string:
			if typedItem == gitState().State {
				return true
			}
		case map[string]any:
			if c.matchesMap(gitState, typedItem) {
				return true
			}
		}
	}

	return false
}

func (c *checker) matchesMap(gitState func() GitState, item map[string]any) bool {
	for _, condition := range c.conditions {
		if condition.Match(gitState, item) {
			return true
		}
	}

	return false
}
