package hook

import (
	"fmt"

	"github.com/oursky/skygear/oddb"
)

// Kind defines when a hook should be executed on mutation of oddb.Record.
type Kind string

// The four kind of hooks provided by Ourd.
const (
	BeforeSave   Kind = "beforeSave"
	AfterSave         = "afterSave"
	BeforeDelete      = "beforeDelete"
	AfterDelete       = "afterDelete"
)

// Func defines the interface of a function that can be hooked.
//
// The supplied record is fully fetched for all four kind of hooks.
type Func func(*oddb.Record, *oddb.Record) error

type recordTypeHookMap map[string][]Func

// Registry is a registry of hooks by record type.
//
// It provides method to execute hooks but is not responsible to execute
// registered hook. The responsibility is currently handled by handler.
//
// In future the registry should hook itself into Record lifecycle and manage
// hooks executions itself. Such hooking point does not exist at the moment.
type Registry struct {
	beforeSaveHooks   recordTypeHookMap
	afterSaveHooks    recordTypeHookMap
	beforeDeleteHooks recordTypeHookMap
	afterDeleteHooks  recordTypeHookMap
}

// NewRegistry returns a Registry ready for use.
func NewRegistry() *Registry {
	return &Registry{
		recordTypeHookMap{},
		recordTypeHookMap{},
		recordTypeHookMap{},
		recordTypeHookMap{},
	}
}

// Register adds the specific hook for the supplied recordType to be executed
// at the moment provided by kind.
func (r *Registry) Register(kind Kind, recordType string, hook Func) error {
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
func (r *Registry) ExecuteHooks(kind Kind, record *oddb.Record, oldRecord *oddb.Record) error {
	recordTypeHookMap, err := r.recordTypeHookMap(kind)
	if err != nil {
		return err
	}

	for _, hook := range recordTypeHookMap[record.ID.Type] {
		if err := hook(record, oldRecord); err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) recordTypeHookMap(kind Kind) (m recordTypeHookMap, err error) {
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
