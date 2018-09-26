package inject

import (
	"context"
	"reflect"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type Map interface {
	Provide(name string, ctx context.Context, configuration config.TenantConfiguration) interface{}
}

func DefaultInject(
	i interface{},
	dependencyMap Map,
	ctx context.Context,
	configuration config.TenantConfiguration,
) {
	injectDependency(i, dependencyMap, ctx, configuration)
}

func injectDependency(
	i interface{},
	dependencyMap Map,
	ctx context.Context,
	configuration config.TenantConfiguration,
) {
	t := reflect.TypeOf(i).Elem()
	v := reflect.ValueOf(i).Elem()

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		dependencyName := t.Field(i).Tag.Get("dependency")
		if dependencyName == "" {
			continue
		}

		field := v.Field(i)
		dependency := dependencyMap.Provide(dependencyName, ctx, configuration)
		field.Set(reflect.ValueOf(dependency))
	}
}
