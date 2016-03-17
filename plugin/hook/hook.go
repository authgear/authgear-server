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

package hook

import (
	"fmt"
	"sync"

	"github.com/skygeario/skygear-server/skydb"
	"github.com/skygeario/skygear-server/skyerr"
	"golang.org/x/net/context"
)

// Kind defines when a hook should be executed on mutation of skydb.Record.
type Kind string

// The four kind of hooks provided by Skygear.
const (
	BeforeSave   Kind = "beforeSave"
	AfterSave         = "afterSave"
	BeforeDelete      = "beforeDelete"
	AfterDelete       = "afterDelete"
)

// Func defines the interface of a function that can be hooked.
//
// The supplied record is fully fetched for all four kind of hooks.
type Func func(context.Context, *skydb.Record, *skydb.Record) skyerr.Error

type recordTypeHookMap map[string][]Func

// Registry is a registry of hooks by record type.
//
// It provides method to execute hooks but is not responsible to execute
// registered hook. The responsibility is currently handled by handler.
//
// In future the registry should hook itself into Record lifecycle and manage
// hooks executions itself. Such hooking point does not exist at the moment.
type Registry struct {
	mutex             sync.RWMutex
	beforeSaveHooks   recordTypeHookMap
	afterSaveHooks    recordTypeHookMap
	beforeDeleteHooks recordTypeHookMap
	afterDeleteHooks  recordTypeHookMap
}

// NewRegistry returns a Registry ready for use.
func NewRegistry() *Registry {
	return &Registry{
		sync.RWMutex{},
		recordTypeHookMap{},
		recordTypeHookMap{},
		recordTypeHookMap{},
		recordTypeHookMap{},
	}
}

// Register adds the specific hook for the supplied recordType to be executed
// at the moment provided by kind.
func (r *Registry) Register(kind Kind, recordType string, hook Func) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	recordTypeHookMap, err := r.recordTypeHookMap(kind)
	if err != nil {
		return err
	}

	recordTypeHookMap[recordType] = append(recordTypeHookMap[recordType], hook)
	return nil
}

// ExecuteHooks executes registered hooks for the type of supplied record to
// be executed at the specific kind of moment.
//
// If one of the hooks returns an error, it halts execution of other hooks and
// return sthat error untouched.
func (r *Registry) ExecuteHooks(ctx context.Context, kind Kind, record *skydb.Record, oldRecord *skydb.Record) skyerr.Error {
	hooks, err := r.hooks(kind, record.ID.Type)
	if err != nil {
		return skyerr.NewError(skyerr.UnexpectedError, "Error getting database hooks")
	}

	for _, hook := range hooks {
		if err := hook(ctx, record, oldRecord); err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) hooks(kind Kind, recordType string) (m []Func, err error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	recordTypeHookMap, err := r.recordTypeHookMap(kind)
	if err != nil {
		return nil, err
	}

	hooks := make([]Func, len(recordTypeHookMap[recordType]))
	copy(hooks, recordTypeHookMap[recordType])
	return hooks, nil
}

func (r *Registry) recordTypeHookMap(kind Kind) (m recordTypeHookMap, err error) {
	// Note: do not acquire read lock here
	// acquire lock before calling this function

	switch kind {
	default:
		err = fmt.Errorf("unrecgonized kind of hook = %#v", string(kind))
	case BeforeSave:
		m = r.beforeSaveHooks
	case AfterSave:
		m = r.afterSaveHooks
	case BeforeDelete:
		m = r.beforeDeleteHooks
	case AfterDelete:
		m = r.afterDeleteHooks
	}

	return
}
