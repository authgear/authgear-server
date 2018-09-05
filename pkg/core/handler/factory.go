package handler

import (
	"context"
	"reflect"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Factory interface {
	NewHandler(context.Context, config.TenantConfiguration) Handler
}

type ProviderGraph interface {
	Provide(name string, ctx context.Context, configuration config.TenantConfiguration) interface{}
}

func DefaultInject(
	h Handler,
	dependencyGraph ProviderGraph,
	ctx context.Context,
	configuration config.TenantConfiguration,
) {
	t := reflect.TypeOf(h).Elem()
	v := reflect.ValueOf(h).Elem()

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		dependencyName := t.Field(i).Tag.Get("dependency")
		field := v.Field(i)
		dependency := dependencyGraph.Provide(dependencyName, ctx, configuration)
		field.Set(reflect.ValueOf(dependency))
	}
}
