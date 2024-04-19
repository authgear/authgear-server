package authenticationflow

import "errors"

// Milestone is a marker.
// The designed use case is to find out whether a particular milestone exists
// in the flow, or any of its subflows.
type Milestone interface {
	Milestone()
}

func FindFirstMilestone[T Milestone](w *Flow) (T, bool) {
	return findMilestone[T](w, true)
}

func FindMilestone[T Milestone](w *Flow) (T, bool) {
	return findMilestone[T](w, false)
}

// This function only find milestones in the provided flow,
// and will not nest into subflows
func FindMilestoneInCurrentFlow[T Milestone](w *Flow) (T, bool) {
	var t T
	found := false
	for _, node := range w.Nodes {
		n := node
		switch n.Type {
		case NodeTypeSimple:
			if m, ok := n.Simple.(T); ok {
				t = m
				found = true
			}
		case NodeTypeSubFlow:
			if m, ok := n.SubFlow.Intent.(T); ok {
				t = m
				found = true
			}
		default:
			panic(errors.New("unreachable"))
		}
	}
	return t, found
}

func findMilestone[T Milestone](w *Flow, stopOnFirst bool) (T, bool) {
	var t T
	found := false

	err := TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, _ *Flow) error {
			if m, ok := nodeSimple.(T); ok && (!found || !stopOnFirst) {
				t = m
				found = true
			}
			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			if m, ok := intent.(T); ok && (!found || !stopOnFirst) {
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
