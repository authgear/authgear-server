package authenticationflow

import "errors"

// Milestone is a marker.
// The designed use case is to find out whether a particular milestone exists
// in the flow, or any of its subflows.
type Milestone interface {
	Milestone()
}

// FindMilestoneInCurrentFlow find the last milestone in the flow.
// It does not recur into sub flows.
// If the found milestone is a node, then the returned flows is the same as flows.
// If the found milestone is a intent, then the returned flows is Nearest=intent.
func FindMilestoneInCurrentFlow[T Milestone](flows Flows) (T, Flows, bool) {
	newFlows := flows
	w := flows.Nearest
	var t T
	found := false
	for _, node := range w.Nodes {
		n := node
		switch n.Type {
		case NodeTypeSimple:
			if m, ok := n.Simple.(T); ok {
				t = m
				newFlows = flows.Replace(w)
				found = true
			}
		case NodeTypeSubFlow:
			if m, ok := n.SubFlow.Intent.(T); ok {
				t = m
				newFlows = flows.Replace(n.SubFlow)
				found = true
			}
		default:
			panic(errors.New("unreachable"))
		}
	}
	return t, newFlows, found
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
