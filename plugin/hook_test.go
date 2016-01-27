package plugin

import (
	"errors"
	"testing"

	"github.com/oursky/skygear/plugin/hook"
	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
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
			So(err.Error(), ShouldEqual, "exit status 1")
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
