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

package plugin

import (
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
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

		return skyerr.MakeError(err)
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
