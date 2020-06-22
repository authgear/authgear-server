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

package password

import (
	"github.com/skygeario/skygear-server/pkg/auth/config"
	"github.com/skygeario/skygear-server/pkg/log"
)

type HousekeeperLogger struct {
	*log.Logger
}

func NewHousekeeperLogger(lf *log.Factory) HousekeeperLogger {
	return HousekeeperLogger{lf.New("password-housekeeper")}
}

type Housekeeper struct {
	Store  *HistoryStore
	Logger HousekeeperLogger
	Config *config.PasswordPolicyConfig
}

func (p *Housekeeper) Housekeep(authID string) (err error) {
	if !p.Config.IsEnabled() {
		return
	}

	p.Logger.Debug("remove password history")
	err = p.Store.RemovePasswordHistory(authID, p.Config.HistorySize, p.Config.HistoryDays)
	if err != nil {
		p.Logger.WithError(err).Error("unable to housekeep password history")
	}

	return
}
