package inject

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

type RequestDependencyMap interface {
	Provide(name string, request *http.Request) interface{}
}

type DependencyMap interface {
	Provide(name string, ctx context.Context, requestID string, tenantConfig config.TenantConfiguration) interface{}
}

func DefaultRequestInject(
	i interface{},
	dependencyMap DependencyMap,
	request *http.Request,
) (err error) {
	return injectDependency(i, func(name string) interface{} {
		return dependencyMap.Provide(
			name,
			request.Context(),
			request.Header.Get(coreHttp.HeaderRequestID),
			config.GetTenantConfig(request),
		)
	})
}

func DefaultTaskInject( // nolint: golint
	i interface{},
	dependencyMap DependencyMap,
	ctx context.Context,
	taskCtx async.TaskContext,
) (err error) {
	return injectDependency(i, func(name string) interface{} {
		return dependencyMap.Provide(name, ctx, taskCtx.RequestID, taskCtx.TenantConfig)
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
			err = skyerr.NewError(skyerr.InvalidArgument, `Dependency "`+dependencyName+`" is nil, but does not mark as optional`)
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
