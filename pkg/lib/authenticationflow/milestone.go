package authenticationflow

// Milestone is a marker.
// The designed use case is to find out whether a particular milestone exists
// in the flow, or any of its subflows.
type Milestone interface {
	Milestone()
}

func FindMilestone[T Milestone](w *Flow) (T, bool) {
	var t T
	found := false

	err := TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, _ *Flow) error {
			if m, ok := nodeSimple.(T); ok {
				t = m
				found = true
			}
			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if m, ok := intent.(T); ok {
				t = m
				found = true
			}
			return nil
		},
	}, w)
	if err != nil {
		return *new(T), false
	}

	if !found {
		return *new(T), false
	}

	return t, true
}

func FindAllMilestones[T Milestone](w *Flow) []T {
	var ts []T

	err := TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, _ *Flow) error {
			if m, ok := nodeSimple.(T); ok {
				ts = append(ts, m)
			}
			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if m, ok := intent.(T); ok {
				ts = append(ts, m)
			}
			return nil
		},
	}, w)
	if err != nil {
		panic(err)
	}

	return ts
}
