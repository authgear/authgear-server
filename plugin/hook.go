package plugin

import (
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
)

// CreateHookFunc returns a hook.HookFunc that run the hook registered by a
// plugin
func CreateHookFunc(p *Plugin, hookInfo pluginHookInfo) hook.Func {
	hookFunc := func(record *oddb.Record) error {
		recordout, err := p.transport.RunHook(hookInfo.Type, hookInfo.Trigger, record)
		if err == nil && hookInfo.Trigger == string(hook.BeforeSave) && !hookInfo.Async {
			*record = *recordout
		}
		return err
	}
	if hookInfo.Async {
		return func(record *oddb.Record) error {
			// TODO(limouren): think of a way to test this go routine
			go hookFunc(record)
			return nil
		}
	}

	return hookFunc
}
