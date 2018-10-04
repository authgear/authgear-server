package inject

import (
	"net/http"
	"reflect"
)

type Map interface {
	Provide(name string, request *http.Request) interface{}
}

func DefaultInject(
	i interface{},
	dependencyMap Map,
	request *http.Request,
) {
	injectDependency(i, dependencyMap, request)
}

func injectDependency(
	i interface{},
	dependencyMap Map,
	request *http.Request,
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
		dependency := dependencyMap.Provide(dependencyName, request)
		field.Set(reflect.ValueOf(dependency))
	}
}
