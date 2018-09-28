package inject

import (
	"context"
	"reflect"
	"strings"

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
		dependencyTag := t.Field(i).Tag.Get("dependency")
		if dependencyTag == "" {
			continue
		}

		dependencyArgs := strings.Split(dependencyTag, ",")
		dependencyName := dependencyArgs[0]
		options := dependencyArgs[1:]
		optionSet := arrayToSet(options)

		field := v.Field(i)
		dependency := dependencyMap.Provide(dependencyName, ctx, configuration)

		if optionSet["optional"] == nil && dependency == nil {
			panic(`Dependency "` + dependencyName + `" is nil, but does not mark as optional`)
		}

		if dependency != nil {
			field.Set(reflect.ValueOf(dependency))
		}
	}
}

func arrayToSet(array []string) (set map[string]interface{}) {
	set = make(map[string]interface{})
	for _, value := range array {
		set[value] = struct{}{}
	}

	return
}
