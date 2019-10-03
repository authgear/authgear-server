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
	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/passwordhistory"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

func NewPwHousekeeper(
	passwordHistoryStore passwordhistory.Store,
	loggerFactory logging.Factory,
	pwHistorySize int,
	pwHistoryDays int,
	passwordHistoryEnabled bool,
) *PwHousekeeper {
	return &PwHousekeeper{
		passwordHistoryStore:   passwordHistoryStore,
		logger:                 loggerFactory.NewLogger("password-housekeeper"),
		pwHistorySize:          pwHistorySize,
		pwHistoryDays:          pwHistoryDays,
		passwordHistoryEnabled: passwordHistoryEnabled,
	}
}

type PwHousekeeper struct {
	passwordHistoryStore   passwordhistory.Store
	logger                 *logrus.Entry
	pwHistorySize          int
	pwHistoryDays          int
	passwordHistoryEnabled bool
}

func (p *PwHousekeeper) Housekeep(authID string) (err error) {
	if !p.enabled() {
		return
	}

	p.logger.Debug("Remove password history")
	err = p.passwordHistoryStore.RemovePasswordHistory(authID, p.pwHistorySize, p.pwHistoryDays)
	if err != nil {
		p.logger.WithError(err).Error("Unable to housekeep password history")
	}

	return
}

func (p *PwHousekeeper) enabled() bool {
	return p.passwordHistoryEnabled
}
