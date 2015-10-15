package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/oursky/skygear/ourtest"
	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func tempdir() string {
	dir, err := ioutil.TempDir("", "com.oursky.skygear.skydb.fs")
	if err != nil {
		panic(err)
	}

	return dir
}

func getDatabase(name string, userID string) (dir string, db *fileDatabase) {
	if name == "" {
		name = "fs-test"
	}
	dir = tempDir()
	db = newDatabase(nil, dir, name, userID)
	return
}

func transformRows(rows *skydb.Rows, err error) ([]skydb.Record, error) {
	if err != nil {
		return nil, err
	}

	records := []skydb.Record{}
	for rows.Scan() {
		records = append(records, rows.Record())
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return records, nil
}

func TestSave(t *testing.T) {
	Convey("A Database", t, func() {
		dir, db := getDatabase("fs.save", "someuserid")

		Convey("saves record correctly", func() {
			const expectedFileContent = `{"ID":"note/someid","Data":{"bool":true,"number":1,"string":"string"},"OwnerID":"","ACL":[{"relation":"friend","level":"read"}]}
`
			record := skydb.Record{
				ID:        skydb.NewRecordID("note", "someid"),
				OwnerID:   "owner",
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				CreatorID: "creator",
				UpdatedAt: time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
				UpdaterID: "updater",
				ACL: skydb.NewRecordACL([]skydb.RecordACLEntry{
					skydb.NewRecordACLEntryRelation("friend", skydb.ReadLevel),
				}),
				Data: skydb.Data{
					"string": "string",
					"number": float64(1),
					"bool":   true,
				},
			}
			err := db.Save(&record)
			So(err, ShouldBeNil)
			So(record.DatabaseID, ShouldEqual, "someuserid")

			contentBytes, err := ioutil.ReadFile(filepath.Join(dir, "note", "someid"))
			So(err, ShouldBeNil)

			content := string(contentBytes)
			So(content, ShouldEqualJSON, `{
				"ID": "note/someid",
				"OwnerID": "owner",
				"CreatedAt": "2006-01-02T15:04:05Z",
				"CreatorID": "creator",
				"UpdatedAt": "2007-01-02T15:04:05Z",
				"UpdaterID": "updater",
				"ACL": [
					{
						"relation": "friend",
						"level": "read"
					}
				],
				"Data": {
					"bool": true,
					"number": 1,
					"string": "string"
				}
			}`)
		})

		Reset(func() {
			os.RemoveAll(dir)
		})
	})
}

func TestQuerySort(t *testing.T) {
	dir, db := getDatabase("fs.query.sort", "")
	defer os.RemoveAll(dir)

	record1 := skydb.Record{
		ID:   skydb.NewRecordID("record", "1"),
		Data: skydb.Data{"string": "A", "int": float64(2)}}
	record2 := skydb.Record{
		ID:   skydb.NewRecordID("record", "2"),
		Data: skydb.Data{"string": "B", "int": float64(0)}}
	record3 := skydb.Record{
		ID:   skydb.NewRecordID("record", "3"),
		Data: skydb.Data{"string": "C", "int": float64(1)}}

	for _, record := range []skydb.Record{record1, record2, record3} {
		if err := db.Save(&record); err != nil {
			panic(err)
		}
	}

	Convey("Given a Database", t, func() {
		Convey("it sorts by record ID", func() {
			query := skydb.Query{
				Type: "record",
				Sorts: []skydb.Sort{
					{KeyPath: "_id", Order: skydb.Asc},
				},
			}

			records, err := transformRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{
				record1,
				record2,
				record3,
			})
		})

		Convey("it sorts by string value", func() {
			query := skydb.Query{
				Type: "record",
				Sorts: []skydb.Sort{
					{KeyPath: "string", Order: skydb.Desc},
				},
			}

			records, err := transformRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{
				record3,
				record2,
				record1,
			})
		})
		Convey("it sorts by integer value", func() {
			query := skydb.Query{
				Type: "record",
				Sorts: []skydb.Sort{
					{KeyPath: "int", Order: skydb.Asc},
				},
			}

			records, err := transformRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{
				record2,
				record3,
				record1,
			})
		})
	})
}
