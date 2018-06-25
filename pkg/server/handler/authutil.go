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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/skygeario/skygear-server/pkg/server/asset"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/recordutil"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

const (
	// UserRecordLastLoginAtKey is the key for the time when the user
	// last logged in.
	UserRecordLastLoginAtKey = "last_login_at"
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

// FetchUser fetches user record by AuthData for auth, e.g. login
// The query should not be affected by the user record acl
// so this function bypasses access control
func (f *UserAuthFetcher) FetchUser(authData skydb.AuthData) (user skydb.Record, err error) {
	query := f.buildAuthDataQuery(authData)

	var results *skydb.Rows
	results, err = f.Database.Query(&query, &skydb.AccessControlOptions{
		BypassAccessControl: true,
	})
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
//
// DEPRECATED: Do not use. Use authUserRecordContext instead.
type createUserWithRecordContext struct {
	DBConn         skydb.Conn
	Database       skydb.Database
	AssetStore     asset.Store
	HookRegistry   *hook.Registry
	AuthRecordKeys [][]string
	Context        context.Context
}

func (ctx *createUserWithRecordContext) execute(info *skydb.AuthInfo, authData skydb.AuthData, profile skydb.Data) (*skydb.Record, skyerr.Error) {
	newCtx := authUserRecordContext{
		DBConn:         ctx.DBConn,
		Database:       ctx.Database,
		AssetStore:     ctx.AssetStore,
		HookRegistry:   ctx.HookRegistry,
		AuthRecordKeys: ctx.AuthRecordKeys,
		Context:        ctx.Context,

		BeforeSaveFunc: func(conn skydb.Conn, info *skydb.AuthInfo) error {
			if err := conn.CreateAuth(info); err != nil {
				if err == skydb.ErrUserDuplicated {
					return errUserDuplicated
				}

				return skyerr.NewResourceSaveFailureErr("auth_data", authData)
			}
			return nil
		},
	}
	return newCtx.execute(info, authData, profile)
}

// authUserRecordContext is a context for manipulating a user with
// database record
type authUserRecordContext struct {
	DBConn         skydb.Conn
	Database       skydb.Database
	AssetStore     asset.Store
	HookRegistry   *hook.Registry
	AuthRecordKeys [][]string
	Context        context.Context

	// BeforeSaveFunc is a function that is called before saving the user
	// record to the database. This function is called in the same transaction
	// of saving the user record.
	//
	// The before save func is useful for saving the pass AuthInfo.
	BeforeSaveFunc func(conn skydb.Conn, info *skydb.AuthInfo) error

	// BeforeCommitFunc is a function that is called after saving the user
	// record to the database, but before the transaction is committed.
	//
	// The before commit func is useful for saving other data in the
	// same transaction.
	BeforeCommitFunc func(conn skydb.Conn, user *skydb.Record) error
}

func (ctx *authUserRecordContext) execute(info *skydb.AuthInfo, authData skydb.AuthData, profile skydb.Data) (*skydb.Record, skyerr.Error) {
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
	_, err := recordutil.ExtendRecordSchema(ctx.Context, db, []*skydb.Record{&userRecord})
	if err != nil {
		logger := logging.CreateLogger(ctx.Context, "handler")
		logger.WithError(err).Errorln("failed to migrate record schema")
		if myerr, ok := err.(skyerr.Error); ok {
			return nil, myerr
		}
		return nil, skyerr.NewError(skyerr.IncompatibleSchema, "failed to migrate record schema")
	}

	var user *skydb.Record
	txErr := skydb.WithTransaction(txDB, func() error {
		if ctx.BeforeSaveFunc != nil {
			if err := ctx.BeforeSaveFunc(ctx.DBConn, info); err != nil {
				return err
			}
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

		if err := recordResp.ErrMap[userRecord.ID]; err != nil {
			if skyErr, ok := err.(skyerr.Error); ok && skyErr.Code() == skyerr.Duplicated {
				return errUserDuplicated
			}

			return err
		}

		user = recordResp.SavedRecords[0]

		if ctx.BeforeCommitFunc != nil {
			if err := ctx.BeforeCommitFunc(ctx.DBConn, user); err != nil {
				return err
			}
		}
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

	for k, v := range authData.GetData() {
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

// checkUserIsNotDisabled is used by login handlers to check if the user is
// not disabled.
func checkUserIsNotDisabled(authInfo *skydb.AuthInfo) skyerr.Error {
	if authInfo.IsDisabled() {
		logrus.WithFields(logrus.Fields{
			"auth_id": authInfo.ID,
		}).Info("User is disabled")
		info := map[string]interface{}{}
		if authInfo.DisabledExpiry != nil {
			info["expiry"] = authInfo.DisabledExpiry.Format(time.RFC3339)
		}
		if authInfo.DisabledMessage != "" {
			info["message"] = authInfo.DisabledMessage
		}
		return skyerr.NewErrorWithInfo(skyerr.UserDisabled, "user is disabled", info)
	}

	authInfo.RefreshDisabledStatus()
	return nil
}
