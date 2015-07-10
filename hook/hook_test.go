package hook

import (
	"testing"

	"github.com/oursky/ourd/oddb"
	. "github.com/smartystreets/goconvey/convey"
)

type stackingHook struct {
	records []*oddb.Record
}

func (p *stackingHook) Func(record *oddb.Record) error {
	p.records = append(p.records, record)
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

			record := &oddb.Record{
				ID: oddb.NewRecordID("record", "id"),
			}

			Convey("for beforeSave", func() {
				registry.ExecuteHooks(BeforeSave, record)
				So(beforeSave.records, ShouldResemble, []*oddb.Record{record})
				So(afterSave.records, ShouldBeEmpty)
				So(beforeDelete.records, ShouldBeEmpty)
				So(afterDelete.records, ShouldBeEmpty)
			})

			Convey("for afterSave", func() {
				registry.ExecuteHooks(AfterSave, record)
				So(beforeSave.records, ShouldBeEmpty)
				So(afterSave.records, ShouldResemble, []*oddb.Record{record})
				So(beforeDelete.records, ShouldBeEmpty)
				So(afterDelete.records, ShouldBeEmpty)
			})

			Convey("for beforeDelete", func() {
				registry.ExecuteHooks(BeforeDelete, record)
				So(beforeSave.records, ShouldBeEmpty)
				So(afterSave.records, ShouldBeEmpty)
				So(beforeDelete.records, ShouldResemble, []*oddb.Record{record})
				So(afterDelete.records, ShouldBeEmpty)
			})

			Convey("for afterDelete", func() {
				registry.ExecuteHooks(AfterDelete, record)
				So(beforeSave.records, ShouldBeEmpty)
				So(afterSave.records, ShouldBeEmpty)
				So(beforeDelete.records, ShouldBeEmpty)
				So(afterDelete.records, ShouldResemble, []*oddb.Record{record})
			})
		})

		Convey("executes multiple hooks", func() {
			hook1 := stackingHook{}
			hook2 := stackingHook{}
			registry.Register(AfterSave, "note", hook1.Func)
			registry.Register(AfterSave, "note", hook2.Func)

			record := &oddb.Record{
				ID: oddb.NewRecordID("note", "id"),
			}
			registry.ExecuteHooks(AfterSave, record)

			So(hook1.records, ShouldResemble, []*oddb.Record{record})
			So(hook2.records, ShouldResemble, []*oddb.Record{record})
		})

		Convey("executes no hooks", func() {
			record := &oddb.Record{
				ID: oddb.NewRecordID("record", "id"),
			}
			So(func() {
				registry.ExecuteHooks(BeforeDelete, record)
			}, ShouldNotPanic)
		})

		Convey("panics executing nil record", func() {
			So(func() {
				registry.ExecuteHooks(AfterDelete, nil)
			}, ShouldPanic)
		})
	})
}
