package interaction

type Effect interface {
	apply(ctx *Context) error
}

type EffectRun func(ctx *Context) error

func (e EffectRun) apply(ctx *Context) error {
	if ctx.IsCommitting {
		return nil
	}
	return e(ctx)
}

type EffectOnCommit func(ctx *Context) error

func (e EffectOnCommit) apply(ctx *Context) error {
	if ctx.IsDryRun || !ctx.IsCommitting {
		return nil
	}

	return e(ctx)
}
