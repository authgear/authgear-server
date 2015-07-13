package plugin

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
)

// CreateHookFunc returns a hook.HookFunc that run the hook registered by a
// plugin
func CreateHookFunc(p *Plugin, hookInfo pluginHookInfo) hook.Func {
	hookFunc := func(record *oddb.Record) error {
		b, err := json.Marshal(record)
		if err != nil {
			panic("failed to marshal record: " + err.Error())
		}

		out, err := p.transport.RunHook(hookInfo.Type, hookInfo.Trigger, b)
		log.Debugf("Executed a hook with result: %s", out)
		return err
	}
	if hookInfo.Async {
		return func(record *oddb.Record) error {
			go hookFunc(record)
			return nil
		}
	}

	return hookFunc
}
