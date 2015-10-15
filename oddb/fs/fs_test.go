package fs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oursky/skygear/oddb"
	. "github.com/oursky/skygear/ourtest"
	. "github.com/smartystreets/goconvey/convey"
)

func tempdir() string {
	dir, err := ioutil.TempDir("", "com.oursky.skygear.oddb.fs")
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

func TestSave(t *testing.T) {
	Convey("A Database", t, func() {
		dir, db := getDatabase("fs.save", "someuserid")

		Convey("saves record correctly", func() {
			const expectedFileContent = `{"ID":"note/someid","Data":{"bool":true,"number":1,"string":"string"},"OwnerID":"","ACL":[{"relation":"friend","level":"read"}]}
`
			record := oddb.Record{
				ID:        oddb.NewRecordID("note", "someid"),
				OwnerID:   "owner",
				CreatedAt: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
				CreatorID: "creator",
				UpdatedAt: time.Date(2007, 1, 2, 15, 4, 5, 0, time.UTC),
				UpdaterID: "updater",
				ACL: oddb.NewRecordACL([]oddb.RecordACLEntry{
					oddb.NewRecordACLEntryRelation("friend", oddb.ReadLevel),
				}),
				Data: oddb.Data{
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

	record1 := oddb.Record{
		ID:   oddb.NewRecordID("record", "1"),
		Data: oddb.Data{"string": "A", "int": float64(2)}}
	record2 := oddb.Record{
		ID:   oddb.NewRecordID("record", "2"),
		Data: oddb.Data{"string": "B", "int": float64(0)}}
	record3 := oddb.Record{
		ID:   oddb.NewRecordID("record", "3"),
		Data: oddb.Data{"string": "C", "int": float64(1)}}

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
					{KeyPath: "_id", Order: oddb.Asc},
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
					{KeyPath: "string", Order: oddb.Desc},
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
					{KeyPath: "int", Order: oddb.Asc},
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
