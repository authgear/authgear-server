package plugin

import (
	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/skydb"
	"golang.org/x/net/context"
)

// CreateHookFunc returns a hook.HookFunc that run the hook registered by a
// plugin
func CreateHookFunc(p *Plugin, hookInfo pluginHookInfo) hook.Func {
	hookFunc := func(ctx context.Context, record *skydb.Record, oldRecord *skydb.Record) error {
		recordout, err := p.transport.RunHook(ctx, hookInfo.Name, record, oldRecord)
		if err == nil && hookInfo.Trigger == string(hook.BeforeSave) && !hookInfo.Async {
			*record = *recordout
		}
		return err
	}
	if hookInfo.Async {
		return func(ctx context.Context, record *skydb.Record, oldRecord *skydb.Record) error {
			// TODO(limouren): think of a way to test this go routine
			go hookFunc(ctx, record, oldRecord)
			return nil
		}
	}

	return hookFunc
}
