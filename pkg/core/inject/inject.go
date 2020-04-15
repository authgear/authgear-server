package inject

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type DependencyMap interface {
	Provide(name string, request *http.Request, ctx context.Context, requestID string, tenantConfig config.TenantConfiguration) interface{}
}

func DefaultRequestInject(
	i interface{},
	dependencyMap DependencyMap,
	request *http.Request,
) (err error) {
	ctx := WithInject(request.Context())
	return injectDependency(i, func(name string) interface{} {
		return dependencyMap.Provide(
			name,
			request,
			ctx,
			request.Header.Get(coreHttp.HeaderRequestID),
			*config.GetTenantConfig(request.Context()),
		)
	})
}

func injectDependency(
	i interface{},
	injectFunc func(name string) interface{},
) (err error) {
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
		dependency := injectFunc(dependencyName)

		if optionSet["optional"] == nil && dependency == nil {
			err = errors.New(`dependency "` + dependencyName + `" is nil, but does not mark as optional`)
			return
		}

		if dependency != nil {
			field.Set(reflect.ValueOf(dependency))
		}
	}
	return
}

func arrayToSet(array []string) (set map[string]interface{}) {
	set = make(map[string]interface{})
	for _, value := range array {
		set[value] = struct{}{}
	}

	return
}
