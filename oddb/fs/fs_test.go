package fs

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"io/ioutil"
	"os"

	"github.com/oursky/ourd/oddb"
)

func tempdir() string {
	dir, err := ioutil.TempDir("", "com.oursky.ourd.oddb.fs")
	if err != nil {
		panic(err)
	}

	return dir
}

func transformRows(rows *oddb.Rows, err error) ([]oddb.Record, error) {
	if err != nil {
		return nil, err
	}

	records := []oddb.Record{}
	for rows.Scan() {
		records = append(records, rows.Record())
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return records, nil
}

func TestQuerySort(t *testing.T) {
	dir := tempdir()
	defer os.RemoveAll(dir)
	db := newDatabase(nil, dir, "query.sort")

	record1 := oddb.Record{Type: "record", Key: "1", Data: oddb.Data{"string": "A", "int": float64(2)}}
	record2 := oddb.Record{Type: "record", Key: "2", Data: oddb.Data{"string": "B", "int": float64(0)}}
	record3 := oddb.Record{Type: "record", Key: "3", Data: oddb.Data{"string": "C", "int": float64(1)}}

	for _, record := range []oddb.Record{record1, record2, record3} {
		if err := db.Save(&record); err != nil {
			panic(err)
		}
	}

	Convey("Given a Database", t, func() {
		Convey("it sorts by record ID", func() {
			query := oddb.Query{
				Type: "record",
				Sorts: []oddb.Sort{
					{"_id", oddb.Asc},
				},
			}

			records, err := transformRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []oddb.Record{
				record1,
				record2,
				record3,
			})
		})

		Convey("it sorts by string value", func() {
			query := oddb.Query{
				Type: "record",
				Sorts: []oddb.Sort{
					{"string", oddb.Desc},
				},
			}

			records, err := transformRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []oddb.Record{
				record3,
				record2,
				record1,
			})
		})
		Convey("it sorts by integer value", func() {
			query := oddb.Query{
				Type: "record",
				Sorts: []oddb.Sort{
					{"int", oddb.Asc},
				},
			}

			records, err := transformRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []oddb.Record{
				record2,
				record3,
				record1,
			})
		})
	})
}
