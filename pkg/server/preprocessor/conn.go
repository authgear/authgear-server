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

package preprocessor

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

type ConnPreprocessor struct {
	AppName       string
	AccessControl string
	DBOpener      skydb.DBOpener
	DBImpl        string
	Option        string
	DBConfig      skydb.DBConfig
}

func (p ConnPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	log.Debugf("Opening DBConn: {%v %v %v}", p.DBImpl, p.AppName, p.Option)

	dbConfig := p.DBConfig
	if payload.HasMasterKey() {
		dbConfig.CanMigrate = true
	}
	conn, err := p.DBOpener(payload.Context, p.DBImpl, p.AppName, p.AccessControl, p.Option, dbConfig)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedUnableToOpenDatabase, err.Error())
		return http.StatusServiceUnavailable
	}
	payload.DBConn = conn

	log.Debugf("Get DB OK")

	return http.StatusOK
}
