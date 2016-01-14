package router

import (
	"fmt"
	"reflect"

	"github.com/facebookgo/inject"
	"github.com/facebookgo/structtag"
)

// HandlerInjector is standard way to inject Services and Preprocessors
// specified by the Handler struct.
type HandlerInjector struct {
	ServiceGraph    *inject.Graph
	PreprocessorMap *PreprocessorRegistry
}

func (i *HandlerInjector) InjectServices(h Handler) Handler {
	err := i.ServiceGraph.Provide(&inject.Object{Value: h})
	if err != nil {
		panic(fmt.Sprintf("Unable to set up handler: %v", err))
	}

	err = i.ServiceGraph.Populate()
	if err != nil {
		panic(fmt.Sprintf("Unable to set up handler: %v", err))
	}
	return h
}

func (i *HandlerInjector) InjectProcessors(h Handler) Handler {
	t := reflect.TypeOf(h)
	reflectValue := reflect.ValueOf(h)
	for c := 0; c < t.Elem().NumField(); c++ {
		fieldTag := t.Elem().Field(c).Tag
		found, value, err := structtag.Extract("preprocessor", string(fieldTag))
		if err != nil {
			panic(fmt.Sprintf("Unable to set up handler: %v", err))
		}
		if found {
			processorField := reflectValue.Elem().Field(c)
			processor := reflect.ValueOf((*i.PreprocessorMap)[value])
			processorField.Set(processor)
		}
	}
	h.Setup()
	return h
}

func (i *HandlerInjector) Inject(h Handler) Handler {
	i.InjectServices(h)
	i.InjectProcessors(h)
	return h
}
