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
	"context"
	"errors"
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/plugin/hook"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

type hookOnlyTransport struct {
	RunHookFunc func(context.Context, string, *skydb.Record, *skydb.Record) (*skydb.Record, error)
	Transport
}

func (t *hookOnlyTransport) RunHook(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
	return t.RunHookFunc(ctx, hookName, record, originalRecord)
}

func TestCreateHookFunc(t *testing.T) {
	Convey("CreateHookFunc", t, func() {
		transport := &hookOnlyTransport{}
		plugin := Plugin{transport: transport}

		recordin := skydb.Record{
			ID: skydb.NewRecordID("note", "id"),
		}
		originalRecord := skydb.Record{
			ID: recordin.ID,
		}

		Convey("synced before save", func() {
			hookFunc := CreateHookFunc(&plugin, pluginHookInfo{
				Async:   false,
				Trigger: string(hook.BeforeSave),
				Type:    "note",
				Name:    "note_beforeSave",
			})

			called := false
			transport.RunHookFunc = func(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
				called = true
				So(hookName, ShouldEqual, "note_beforeSave")
				So(*record, ShouldResemble, skydb.Record{
					ID: skydb.NewRecordID("note", "id"),
				})

				return &skydb.Record{ID: skydb.NewRecordID("note", "modifiedid")}, nil
			}

			err := hookFunc(nil, &recordin, &originalRecord)
			So(called, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(recordin, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "modifiedid"),
			})
		})

		Convey("synced before save error result", func() {
			hookFunc := CreateHookFunc(&plugin, pluginHookInfo{
				Async:   false,
				Trigger: string(hook.BeforeSave),
				Type:    "note",
				Name:    "note_beforeSave",
			})

			transport.RunHookFunc = func(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
				return nil, errors.New("exit status 1")
			}

			err := hookFunc(nil, &recordin, &originalRecord)
			So(err.Error(), ShouldEqual, "UnexpectedError: exit status 1")
			So(recordin, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
			})
		})

		Convey("synced after save", func() {
			hookFunc := CreateHookFunc(&plugin, pluginHookInfo{
				Async:   false,
				Trigger: string(hook.AfterSave),
				Type:    "note",
				Name:    "note_afterSave",
			})

			called := false
			transport.RunHookFunc = func(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
				called = true
				So(hookName, ShouldEqual, "note_afterSave")
				So(*record, ShouldResemble, skydb.Record{
					ID: skydb.NewRecordID("note", "id"),
				})

				return &skydb.Record{ID: skydb.NewRecordID("note", "modifiedid")}, nil
			}

			err := hookFunc(nil, &recordin, &originalRecord)
			So(called, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(recordin, ShouldResemble, skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
			})
		})
	})
}
