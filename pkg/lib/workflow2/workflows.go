package workflow2

type Workflows struct {
	Root    *Workflow
	Nearest *Workflow
}

func NewWorkflows(root *Workflow) Workflows {
	return Workflows{
		Root:    root,
		Nearest: root,
	}
}

func (w Workflows) Replace(nearest *Workflow) Workflows {
	w.Nearest = nearest
	return w
}
