package uiparam

import (
	"context"
)

type uiParamContextKeyType struct{}

var uiParamContextKey = uiParamContextKeyType{}

type uiParamContext struct {
	State     string
	UILocales []string
}

type UIParam interface {
	GetState() string
	GetUILocales() []string
}

var _ UIParam = &uiParamContext{}

func (u *uiParamContext) GetState() string       { return u.State }
func (u *uiParamContext) GetUILocales() []string { return u.UILocales }

func WithUIParam(ctx context.Context, state string, uiLocales []string) context.Context {
	v, ok := ctx.Value(uiParamContextKey).(*uiParamContext)
	if ok {
		v.State = state
		v.UILocales = uiLocales
		return ctx
	}
	return context.WithValue(ctx, uiParamContextKey, &uiParamContext{
		State:     state,
		UILocales: uiLocales,
	})
}

func GetUIParam(ctx context.Context) UIParam {
	v, ok := ctx.Value(uiParamContextKey).(*uiParamContext)
	if ok {
		return v
	}
	return &uiParamContext{
		State:     "",
		UILocales: []string{},
	}
}
