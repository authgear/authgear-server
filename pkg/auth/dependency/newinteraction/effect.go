package newinteraction

type Effect interface {
	apply(ctx *Context) error
}

type EffectRun func(ctx *Context) error

func (e EffectRun) apply(ctx *Context) error {
	return e(ctx)
}

type EffectOnCommit func(ctx *Context) error

func (e EffectOnCommit) apply(ctx *Context) error {
	if ctx.IsDryRun {
		return nil
	}

	return e(ctx)
}
