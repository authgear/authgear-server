package plugin

import (
	"errors"
	"testing"

	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
	. "github.com/smartystreets/goconvey/convey"
)

type hookOnlyTransport struct {
	RunHookFunc func(string, string, *oddb.Record) (*oddb.Record, error)
	Transport
}

func (t *hookOnlyTransport) RunHook(recordType string, trigger string, record *oddb.Record) (*oddb.Record, error) {
	return t.RunHookFunc(recordType, trigger, record)
}

func TestCreateHookFunc(t *testing.T) {
	Convey("CreateHookFunc", t, func() {
		transport := &hookOnlyTransport{}
		plugin := Plugin{transport: transport}

		recordin := oddb.Record{
			ID: oddb.NewRecordID("note", "id"),
		}

		Convey("synced before save", func() {
			hookFunc := CreateHookFunc(&plugin, pluginHookInfo{
				Async:   false,
				Trigger: string(hook.BeforeSave),
				Type:    "note",
			})

			called := false
			transport.RunHookFunc = func(recordType string, trigger string, record *oddb.Record) (*oddb.Record, error) {
				called = true
				So(recordType, ShouldEqual, "note")
				So(trigger, ShouldEqual, "beforeSave")
				So(*record, ShouldResemble, oddb.Record{
					ID: oddb.NewRecordID("note", "id"),
				})

				return &oddb.Record{ID: oddb.NewRecordID("note", "modifiedid")}, nil
			}

			err := hookFunc(&recordin)
			So(called, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(recordin, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("note", "modifiedid"),
			})
		})

		Convey("synced before save error result", func() {
			hookFunc := CreateHookFunc(&plugin, pluginHookInfo{
				Async:   false,
				Trigger: string(hook.BeforeSave),
				Type:    "note",
			})

			transport.RunHookFunc = func(recordType string, trigger string, record *oddb.Record) (*oddb.Record, error) {
				return nil, errors.New("exit status 1")
			}

			err := hookFunc(&recordin)
			So(err.Error(), ShouldEqual, "exit status 1")
			So(recordin, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("note", "id"),
			})
		})

		Convey("synced after save", func() {
			hookFunc := CreateHookFunc(&plugin, pluginHookInfo{
				Async:   false,
				Trigger: string(hook.AfterSave),
				Type:    "note",
			})

			called := false
			transport.RunHookFunc = func(recordType string, trigger string, record *oddb.Record) (*oddb.Record, error) {
				called = true
				So(recordType, ShouldEqual, "note")
				So(trigger, ShouldEqual, "afterSave")
				So(*record, ShouldResemble, oddb.Record{
					ID: oddb.NewRecordID("note", "id"),
				})

				return &oddb.Record{ID: oddb.NewRecordID("note", "modifiedid")}, nil
			}

			err := hookFunc(&recordin)
			So(called, ShouldBeTrue)
			So(err, ShouldBeNil)
			So(recordin, ShouldResemble, oddb.Record{
				ID: oddb.NewRecordID("note", "id"),
			})
		})
	})
}
