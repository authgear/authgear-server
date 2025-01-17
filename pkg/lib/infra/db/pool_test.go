package db

import (
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPool(t *testing.T) {
	// dummyPoolOpener returns a new pointer to an zero value of sql.DB.
	dummyPoolOpener := func(info ConnectionInfo, opts ConnectionOptions) (*sql.DB, error) {
		return &sql.DB{}, nil
	}

	Convey("Pool", t, func() {
		original := actualPoolOpener
		defer func() {
			actualPoolOpener = original
		}()

		actualPoolOpener = dummyPoolOpener

		pool := NewPool()
		opts := ConnectionOptions{}

		global := ConnectionInfo{
			Purpose:     ConnectionPurposeGlobal,
			DatabaseURL: "postgresql://localhost:5432/db1",
		}
		app := ConnectionInfo{
			Purpose:     ConnectionPurposeApp,
			DatabaseURL: "postgresql://localhost:5432/db1",
		}
		auditRead := ConnectionInfo{
			Purpose:     ConnectionPurposeAuditReadOnly,
			DatabaseURL: "postgresql://localhost:5432/db1",
		}
		auditWrite := ConnectionInfo{
			Purpose:     ConnectionPurposeAuditReadWrite,
			DatabaseURL: "postgresql://localhost:5432/db1",
		}
		search := ConnectionInfo{
			Purpose:     ConnectionPurposeSearch,
			DatabaseURL: "postgresql://localhost:5432/db1",
		}

		globaldb, err := pool.Open(global, opts)
		So(err, ShouldBeNil)
		So(globaldb, ShouldNotBeNil)

		appdb1, err := pool.Open(app, opts)
		So(err, ShouldBeNil)
		So(appdb1, ShouldNotBeNil)

		appdb2, err := pool.Open(app, opts)
		So(err, ShouldBeNil)
		So(appdb2, ShouldNotBeNil)

		auditdb_Read, err := pool.Open(auditRead, opts)
		So(err, ShouldBeNil)
		So(auditdb_Read, ShouldNotBeNil)

		auditdb_Write, err := pool.Open(auditWrite, opts)
		So(err, ShouldBeNil)
		So(auditdb_Write, ShouldNotBeNil)

		searchdb, err := pool.Open(search, opts)
		So(err, ShouldBeNil)
		So(searchdb, ShouldNotBeNil)

		appdb3, err := pool.Open(ConnectionInfo{
			Purpose:     ConnectionPurposeApp,
			DatabaseURL: "postgresql://localhost:5432/db1",
		}, opts)
		So(err, ShouldBeNil)
		So(appdb3, ShouldNotBeNil)

		// ---

		So(appdb1 != globaldb, ShouldBeTrue)
		So(appdb1 == appdb2, ShouldBeTrue)

		So(auditdb_Read != globaldb, ShouldBeTrue)
		So(auditdb_Read != appdb1, ShouldBeTrue)
		So(auditdb_Read != appdb2, ShouldBeTrue)

		So(auditdb_Write != globaldb, ShouldBeTrue)
		So(auditdb_Write != appdb1, ShouldBeTrue)
		So(auditdb_Write != appdb2, ShouldBeTrue)
		So(auditdb_Write != auditdb_Read, ShouldBeTrue)

		So(searchdb != globaldb, ShouldBeTrue)
		So(searchdb != appdb1, ShouldBeTrue)
		So(searchdb != appdb2, ShouldBeTrue)
		So(searchdb != auditdb_Read, ShouldBeTrue)
		So(searchdb != auditdb_Write, ShouldBeTrue)

		So(appdb3 != globaldb, ShouldBeTrue)
		So(appdb3 == appdb1, ShouldBeTrue)
		So(appdb3 == appdb2, ShouldBeTrue)
		So(appdb3 != auditdb_Read, ShouldBeTrue)
		So(appdb3 != auditdb_Write, ShouldBeTrue)
		So(appdb3 != searchdb, ShouldBeTrue)
	})
}
