package authenticationflow

import (
	"context"
	"fmt"
)

type EffectGetter interface {
	GetEffects(ctx context.Context, deps *Dependencies, flows Flows) (effs []Effect, err error)
}

type Effect interface {
	doNotCallThisDirectly(ctx context.Context, deps *Dependencies) error
}

type RunEffect func(ctx context.Context, deps *Dependencies) error

func (e RunEffect) doNotCallThisDirectly(ctx context.Context, deps *Dependencies) error {
	return e(ctx, deps)
}

type OnCommitEffect func(ctx context.Context, deps *Dependencies) error

func (e OnCommitEffect) doNotCallThisDirectly(ctx context.Context, deps *Dependencies) error {
	return e(ctx, deps)
}

func ApplyRunEffects(ctx context.Context, deps *Dependencies, flows Flows) error {
	err := TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			effectGetter, ok := nodeSimple.(EffectGetter)
			if !ok {
				return nil
			}
			effs, err := effectGetter.GetEffects(ctx, deps, flows.Replace(w))
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if runEff, ok := eff.(RunEffect); ok {
					err = runEff.doNotCallThisDirectly(ctx, deps)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			effectGetter, ok := intent.(EffectGetter)
			if !ok {
				return nil
			}

			effs, err := effectGetter.GetEffects(ctx, deps, flows.Replace(w))
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if _, ok := eff.(RunEffect); ok {
					// Intent cannot have run-effects.
					panic(fmt.Errorf("%T has RunEffect, which is disallowed", w.Intent))
				}
			}
			return nil
		},
	}, flows.Nearest)
	if err != nil {
		return err
	}

	return nil
}

func ApplyAllEffects(ctx context.Context, deps *Dependencies, flows Flows) error {
	err := ApplyRunEffects(ctx, deps, flows)
	if err != nil {
		return err
	}

	err = TraverseFlow(Traverser{
		NodeSimple: func(nodeSimple NodeSimple, w *Flow) error {
			effectGetter, ok := nodeSimple.(EffectGetter)
			if !ok {
				return nil
			}

			effs, err := effectGetter.GetEffects(ctx, deps, flows.Replace(w))
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if onCommitEff, ok := eff.(OnCommitEffect); ok {
					err = onCommitEff.doNotCallThisDirectly(ctx, deps)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
		Intent: func(intent Intent, w *Flow) error {
			effectGetter, ok := intent.(EffectGetter)
			if !ok {
				return nil
			}

			effs, err := effectGetter.GetEffects(ctx, deps, flows.Replace(w))
			if err != nil {
				return err
			}
			for _, eff := range effs {
				if onCommitEff, ok := eff.(OnCommitEffect); ok {
					err = onCommitEff.doNotCallThisDirectly(ctx, deps)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	}, flows.Nearest)

	if err != nil {
		return err
	}

	return nil
}
