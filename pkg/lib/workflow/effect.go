package workflow

type Effect interface {
	doNotCallThisDirectly(ctx *Context) error
}

type RunEffect func(ctx *Context) error

func (e RunEffect) doNotCallThisDirectly(ctx *Context) error {
	return e(ctx)
}

type OnCommitEffect func(ctx *Context) error

func (e OnCommitEffect) doNotCallThisDirectly(ctx *Context) error {
	return e(ctx)
}

func applyRunEffect(ctx *Context, eff RunEffect) error {
	return eff.doNotCallThisDirectly(ctx)
}

func applyOnCommitEffect(ctx *Context, eff OnCommitEffect) error {
	return eff.doNotCallThisDirectly(ctx)
}
