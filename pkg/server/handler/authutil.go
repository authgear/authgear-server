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

package handler

import (
	"context"
	"fmt"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/recordutil"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// UserAuthFetcher provides helper functions to fetch AuthInfo and user Record
// with AuthData in a single structs
type UserAuthFetcher struct {
	DBConn   skydb.Conn
	Database skydb.Database
}

func newUserAuthFetcher(db skydb.Database, conn skydb.Conn) UserAuthFetcher {
	return UserAuthFetcher{
		DBConn:   conn,
		Database: db,
	}
}

func (f *UserAuthFetcher) FetchAuth(authData skydb.AuthData) (authInfo skydb.AuthInfo, user skydb.Record, err error) {
	user, err = f.FetchUser(authData)
	if err != nil {
		return
	}

	err = f.DBConn.GetAuth(user.ID.Key, &authInfo)

	return
}

func (f *UserAuthFetcher) FetchUser(authData skydb.AuthData) (user skydb.Record, err error) {
	query := f.buildAuthDataQuery(authData)

	var results *skydb.Rows
	results, err = f.Database.Query(&query)
	if err != nil {
		return
	}
	defer results.Close()

	records := []skydb.Record{}
	for results.Scan() {
		record := results.Record()
		records = append(records, record)
	}

	if err = results.Err(); err != nil {
		return
	}

	recordsQueried := len(records)
	if recordsQueried == 0 {
		err = skydb.ErrUserNotFound
		return
	} else if recordsQueried > 1 {
		panic(fmt.Errorf("want 1 records queried, got %v", recordsQueried))
	}

	user = records[0]
	getAuthDataFromUser(authData, user)

	return
}

func (f *UserAuthFetcher) buildAuthDataQuery(authData skydb.AuthData) skydb.Query {
	one := uint64(1)
	query := skydb.Query{
		Type:      "user",
		Predicate: authData.MakeEqualPredicate(),
		Limit:     &one,
	}
	return query
}

// createUserWithRecordContext is a context for creating a new user with
// database record
type createUserWithRecordContext struct {
	DBConn       skydb.Conn
	Database     skydb.Database
	AssetStore   asset.Store
	HookRegistry *hook.Registry
	Context      context.Context
}

func (ctx *createUserWithRecordContext) execute(info *skydb.AuthInfo, authData skydb.AuthData, profile skydb.Data) (*skydb.Record, skyerr.Error) {
	db := ctx.Database
	txDB, ok := db.(skydb.Transactional)
	if !ok {
		return nil, skyerr.NewError(skyerr.NotSupported, "database impl does not support transaction")
	}

	userRecord := skydb.Record{
		ID:   skydb.NewRecordID(db.UserRecordType(), info.ID),
		Data: mergeAuthDataWithProfile(authData, profile),
	}

	// derive and extend record schema
	// hotfix (Steven-Chan): moved outside of the transaction to prevent deadlock
	_, err := recordutil.ExtendRecordSchema(db, []*skydb.Record{&userRecord})
	if err != nil {
		log.WithField("err", err).Errorln("failed to migrate record schema")
		if myerr, ok := err.(skyerr.Error); ok {
			return nil, myerr
		}
		return nil, skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate record schema")
	}

	var user *skydb.Record
	txErr := skydb.WithTransaction(txDB, func() error {
		// Check if AuthData duplicated only when it is provided
		// AuthData may be absent, e.g. login with provider
		if !authData.IsEmpty() {
			fetcher := newUserAuthFetcher(db, ctx.DBConn)
			_, _, err := fetcher.FetchAuth(authData)
			if err == nil {
				return errUserDuplicated
			}

			if err != skydb.ErrUserNotFound {
				return skyerr.MakeError(err)
			}
		}

		if err := ctx.DBConn.CreateAuth(info); err != nil {
			if err == skydb.ErrUserDuplicated {
				return errUserDuplicated
			}

			return skyerr.NewResourceSaveFailureErr("auth_data", authData)
		}

		recordReq := recordutil.RecordModifyRequest{
			Db:           db,
			Conn:         ctx.DBConn,
			AssetStore:   ctx.AssetStore,
			HookRegistry: ctx.HookRegistry,
			Atomic:       false,
			Context:      ctx.Context,
			AuthInfo:     info,
			ModifyAt:     timeNow(),
			RecordsToSave: []*skydb.Record{
				&userRecord,
			},
		}

		recordResp := recordutil.RecordModifyResponse{
			ErrMap: map[skydb.RecordID]skyerr.Error{},
		}

		err := recordutil.RecordSaveHandler(&recordReq, &recordResp)
		if err != nil {
			return err
		}

		user = recordResp.SavedRecords[0]
		return nil
	})

	if txErr == nil {
		return user, nil
	}

	if err, ok := txErr.(skyerr.Error); ok {
		return nil, err
	}

	return nil, skyerr.MakeError(txErr)
}

func mergeAuthDataWithProfile(authData skydb.AuthData, profile skydb.Data) skydb.Data {
	if profile == nil {
		profile = skydb.Data{}
	}

	for k, v := range authData {
		profile[k] = v
	}
	return profile
}

// TODO: validate according to settings/options
func validateAuthData(authData skydb.AuthData) error {
	if !authData.IsValid() {
		return fmt.Errorf("Either username and email must be set")
	}

	return nil
}

// getAuthDataFromUser is a AuthData builder function using user Record
func getAuthDataFromUser(authData skydb.AuthData, user skydb.Record) {
	if user.ID.Type != "user" {
		panic("getAuthDataFromUser must be called with user record")
	}

	authData.UpdateFromRecordData(user.Data)
}
