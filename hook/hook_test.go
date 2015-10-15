package hook

import (
	"testing"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

type stackingHook struct {
	records         []*skydb.Record
	originalRecords []*skydb.Record
}

func (p *stackingHook) Func(record *skydb.Record, originalRecord *skydb.Record) error {
	p.records = append(p.records, record)
	p.originalRecords = append(p.originalRecords, originalRecord)
	return nil
}

func TestHookRegistry(t *testing.T) {
	Convey("Registry", t, func() {

		beforeSave := stackingHook{}
		afterSave := stackingHook{}
		beforeDelete := stackingHook{}
		afterDelete := stackingHook{}

		registry := NewRegistry()

		Convey("executes hooks once", func() {
			registry.Register(BeforeSave, "record", beforeSave.Func)
			registry.Register(AfterSave, "record", afterSave.Func)
			registry.Register(BeforeDelete, "record", beforeDelete.Func)
			registry.Register(AfterDelete, "record", afterDelete.Func)

			record := &skydb.Record{
				ID: skydb.NewRecordID("record", "id"),
			}

			originalRecord := &skydb.Record{
				ID: record.ID,
				Data: skydb.Data{
					"value": "old",
				},
			}

			Convey("for beforeSave", func() {
				registry.ExecuteHooks(BeforeSave, record, originalRecord)
				So(beforeSave.records, ShouldResemble, []*skydb.Record{record})
				So(beforeSave.originalRecords, ShouldResemble, []*skydb.Record{originalRecord})
				So(afterSave.records, ShouldBeEmpty)
				So(afterSave.originalRecords, ShouldBeEmpty)
				So(beforeDelete.records, ShouldBeEmpty)
				So(afterDelete.records, ShouldBeEmpty)
			})

			Convey("for afterSave", func() {
				registry.ExecuteHooks(AfterSave, record, originalRecord)
				So(beforeSave.records, ShouldBeEmpty)
				So(beforeSave.originalRecords, ShouldBeEmpty)
				So(afterSave.records, ShouldResemble, []*skydb.Record{record})
				So(afterSave.originalRecords, ShouldResemble, []*skydb.Record{originalRecord})
				So(beforeDelete.records, ShouldBeEmpty)
				So(afterDelete.records, ShouldBeEmpty)
			})

			Convey("for beforeDelete", func() {
				registry.ExecuteHooks(BeforeDelete, record, originalRecord)
				So(beforeSave.records, ShouldBeEmpty)
				So(beforeSave.originalRecords, ShouldBeEmpty)
				So(afterSave.records, ShouldBeEmpty)
				So(afterSave.originalRecords, ShouldBeEmpty)
				So(beforeDelete.records, ShouldResemble, []*skydb.Record{record})
				So(afterDelete.records, ShouldBeEmpty)
			})

			Convey("for afterDelete", func() {
				registry.ExecuteHooks(AfterDelete, record, originalRecord)
				So(beforeSave.records, ShouldBeEmpty)
				So(beforeSave.originalRecords, ShouldBeEmpty)
				So(afterSave.records, ShouldBeEmpty)
				So(afterSave.originalRecords, ShouldBeEmpty)
				So(beforeDelete.records, ShouldBeEmpty)
				So(afterDelete.records, ShouldResemble, []*skydb.Record{record})
			})
		})

		Convey("executes multiple hooks", func() {
			hook1 := stackingHook{}
			hook2 := stackingHook{}
			registry.Register(AfterSave, "note", hook1.Func)
			registry.Register(AfterSave, "note", hook2.Func)

			record := &skydb.Record{
				ID: skydb.NewRecordID("note", "id"),
			}
			originalRecord := &skydb.Record{
				ID: record.ID,
				Data: skydb.Data{
					"value": "old",
				},
			}
			registry.ExecuteHooks(AfterSave, record, originalRecord)

			So(hook1.records, ShouldResemble, []*skydb.Record{record})
			So(hook2.records, ShouldResemble, []*skydb.Record{record})
			So(hook1.originalRecords, ShouldResemble, []*skydb.Record{originalRecord})
			So(hook2.originalRecords, ShouldResemble, []*skydb.Record{originalRecord})
		})

		Convey("executes no hooks", func() {
			record := &skydb.Record{
				ID: skydb.NewRecordID("record", "id"),
			}
			So(func() {
				registry.ExecuteHooks(BeforeDelete, record, nil)
			}, ShouldNotPanic)
		})

		Convey("panics executing nil record", func() {
			So(func() {
				registry.ExecuteHooks(AfterDelete, nil, nil)
			}, ShouldPanic)
		})
	})
}
