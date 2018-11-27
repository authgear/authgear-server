package inject

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type Map interface {
	Provide(name string, request *http.Request) interface{}
}

func DefaultInject(
	i interface{},
	dependencyMap Map,
	request *http.Request,
) (err error) {
	return injectDependency(i, dependencyMap, request)
}

func injectDependency(
	i interface{},
	dependencyMap Map,
	request *http.Request,
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
		dependency := dependencyMap.Provide(dependencyName, request)

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
