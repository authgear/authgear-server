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

package audit

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type PwHousekeeper struct {
	AppName       string
	AccessControl string
	DBOpener      skydb.DBOpener
	DBImpl        string
	Option        string
	DBConfig      skydb.DBConfig

	PwHistorySize          int
	PwHistoryDays          int
	PasswordHistoryEnabled bool
}

func (p *PwHousekeeper) doHousekeep(authID string) {
	ctx := context.Background()
	logger := logging.CreateLogger(ctx, "audit")

	if !p.enabled() {
		return
	}

	conn, err := p.DBOpener(ctx, p.DBImpl, p.AppName, p.AccessControl, p.Option, p.DBConfig)
	if err != nil {
		logger.Warnf(`Unable to housekeep password history`)
		return
	}
	defer conn.Close()

	err = conn.RemovePasswordHistory(authID, p.PwHistorySize, p.PwHistoryDays)
	if err != nil {
		logger.Warnf(`Unable to housekeep password history`)
	}
}

func (p *PwHousekeeper) enabled() bool {
	return p.PasswordHistoryEnabled
}

func (p *PwHousekeeper) Housekeep(authID string) {
	go p.doHousekeep(authID)
}
