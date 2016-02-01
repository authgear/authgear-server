package plugin

import (
	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	"golang.org/x/net/context"
)

// CreateHookFunc returns a hook.HookFunc that run the hook registered by a
// plugin
func CreateHookFunc(p *Plugin, hookInfo pluginHookInfo) hook.Func {
	hookFunc := func(ctx context.Context, record *skydb.Record, oldRecord *skydb.Record) skyerr.Error {
		recordout, err := p.transport.RunHook(ctx, hookInfo.Name, record, oldRecord)
		if err == nil && hookInfo.Trigger == string(hook.BeforeSave) && !hookInfo.Async {
			*record = *recordout
		}

		if err == nil {
			return nil
		}

		if pluginError, ok := err.(skyerr.Error); ok {
			return pluginError
		}

		return skyerr.NewUnknownErr(err)
	}
	if hookInfo.Async {
		return func(ctx context.Context, record *skydb.Record, oldRecord *skydb.Record) skyerr.Error {
			// TODO(limouren): think of a way to test this go routine
			go hookFunc(ctx, record, oldRecord)
			return nil
		}
	}

	return hookFunc
}
