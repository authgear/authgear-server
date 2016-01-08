package plugin

import (
	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/skydb"
)

// CreateHookFunc returns a hook.HookFunc that run the hook registered by a
// plugin
func CreateHookFunc(p *Plugin, hookInfo pluginHookInfo) hook.Func {
	hookFunc := func(record *skydb.Record, oldRecord *skydb.Record) error {
		recordout, err := p.transport.RunHook(hookInfo.Type, hookInfo.Trigger, record, oldRecord)
		if err == nil && hookInfo.Trigger == string(hook.BeforeSave) && !hookInfo.Async {
			*record = *recordout
		}
		return err
	}
	if hookInfo.Async {
		return func(record *skydb.Record, oldRecord *skydb.Record) error {
			// TODO(limouren): think of a way to test this go routine
			go hookFunc(record, oldRecord)
			return nil
		}
	}

	return hookFunc
}
