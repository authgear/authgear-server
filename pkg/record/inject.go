package record

import (
	"net/http"
)

type DependencyMap struct{}

func NewDependencyMap() DependencyMap {
	return DependencyMap{}
}

func (m DependencyMap) Provide(dependencyName string, r *http.Request) interface{} {
	return nil
}
