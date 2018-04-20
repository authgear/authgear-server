// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	// NOTE: inject need a unique name for each handler especially when
	// handlers are of the same type. This can occur for plugin.Handler, which
	// is shared among all plugin handlers.
	err := i.ServiceGraph.Provide(&inject.Object{
		Value: h,                          // handler to inject services to
		Name:  fmt.Sprintf("%T@%p", h, h), // give unique name to each handler (e.g. plugin.AcmeHandler@0x1040a0d0)
	})
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
			if !processor.IsValid() {
				panic(fmt.Sprintf(`Preprocessor "%s" does not exist`, value))
			}
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
