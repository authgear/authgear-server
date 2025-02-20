package config

import (
	"context"

	"github.com/authgear/authgear-server/pkg/util/resource"
)

type AppDomains []string

type AppContext struct {
	AppFs     resource.Fs
	PlanFs    resource.Fs
	Resources *resource.Manager
	Config    *Config
	PlanName  string
	Domains   AppDomains
}

type contextKeyType struct{}

var contextKey = contextKeyType{}

type contextValue struct {
	AppContext *AppContext
}

func getContext(ctx context.Context) *contextValue {
	actx, _ := ctx.Value(contextKey).(*contextValue)
	return actx
}

func WithAppContext(ctx context.Context, appCtx *AppContext) context.Context {
	actx := getContext(ctx)
	if actx == nil {
		actx = &contextValue{}
	}
	actx.AppContext = appCtx
	return context.WithValue(ctx, contextKey, actx)
}

func GetAppContext(ctx context.Context) (*AppContext, bool) {
	actx := getContext(ctx)
	if actx == nil || actx.AppContext == nil {
		return nil, false
	}
	return actx.AppContext, true
}
