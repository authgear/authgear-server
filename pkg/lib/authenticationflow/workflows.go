package authenticationflow

type Flows struct {
	Root    *Flow
	Nearest *Flow
}

func NewFlows(root *Flow) Flows {
	return Flows{
		Root:    root,
		Nearest: root,
	}
}

func (w Flows) Replace(nearest *Flow) Flows {
	w.Nearest = nearest
	return w
}
