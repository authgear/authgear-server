package uiparam

import (
	"context"
)

type uiParamContextKeyType struct{}

var uiParamContextKey = uiParamContextKeyType{}

type T struct {
	ClientID  string
	Prompt    []string
	State     string
	XState    string
	UILocales string
}

func WithUIParam(ctx context.Context, uiParam *T) context.Context {
	v, ok := ctx.Value(uiParamContextKey).(*T)
	if ok {
		v.ClientID = uiParam.ClientID
		v.Prompt = uiParam.Prompt
		v.State = uiParam.State
		v.XState = uiParam.XState
		v.UILocales = uiParam.UILocales
		return ctx
	}
	return context.WithValue(ctx, uiParamContextKey, uiParam)
}

func GetUIParam(ctx context.Context) *T {
	v, ok := ctx.Value(uiParamContextKey).(*T)
	if ok {
		return v
	}
	return &T{
		ClientID:  "",
		Prompt:    nil,
		State:     "",
		XState:    "",
		UILocales: "",
	}
}
