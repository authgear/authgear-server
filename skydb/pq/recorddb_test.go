package pq

import (
	"testing"
	"time"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUserRelationQuery(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		addUser(t, c, "user1")
		addUser(t, c, "user2") // followed by user1
		addUser(t, c, "user3") // mutual follower of user1
		addUser(t, c, "user4") // friend of user1
		addUser(t, c, "user5") // friend of user4 and followed by user4
		c.AddRelation("user1", "_follow", "user2")
		c.AddRelation("user1", "_follow", "user3")
		c.AddRelation("user3", "_follow", "user1")
		c.AddRelation("user1", "_friend", "user4")
		c.AddRelation("user4", "_friend", "user1")
		c.AddRelation("user4", "_friend", "user5")
		c.AddRelation("user5", "_friend", "user4")
		c.AddRelation("user4", "_follow", "user5")

		record0 := skydb.Record{
			ID:      skydb.NewRecordID("record", "0"),
			OwnerID: "user1",
			Data:    skydb.Data{},
		}
		record1 := skydb.Record{
			ID:      skydb.NewRecordID("record", "1"),
			OwnerID: "user2",
			Data:    skydb.Data{},
		}
		record2 := skydb.Record{
			ID:      skydb.NewRecordID("record", "2"),
			OwnerID: "user3",
			Data:    skydb.Data{},
		}
		record3 := skydb.Record{
			ID:      skydb.NewRecordID("record", "3"),
			OwnerID: "user4",
			Data:    skydb.Data{},
		}
		record4 := skydb.Record{
			ID:      skydb.NewRecordID("record", "4"),
			OwnerID: "user5",
			Data:    skydb.Data{},
		}

		db := c.PublicDB()
		So(db.Extend("record", nil), ShouldBeNil)
		So(db.Save(&record0), ShouldBeNil)
		So(db.Save(&record1), ShouldBeNil)
		So(db.Save(&record2), ShouldBeNil)
		So(db.Save(&record3), ShouldBeNil)
		So(db.Save(&record4), ShouldBeNil)

		sortsByID := []skydb.Sort{
			skydb.Sort{
				KeyPath: "_id",
				Order:   skydb.Ascending,
			},
		}

		Convey("query follow outward", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"_owner", "_follow", "outward", "user1"},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record1, record2})
		})

		Convey("query follow inward", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"_owner", "_follow", "inward", "user2"},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0})
		})

		Convey("query follow outward OR inward", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Or,
					Children: []interface{}{
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: skydb.UserRelationFunc{"_owner", "_follow", "outward", "user1"},
								},
							},
						},
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: skydb.UserRelationFunc{"_owner", "_follow", "inward", "user2"},
								},
							},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0, record1, record2})
		})

		Convey("query follow mutual for user1", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"_owner", "_follow", "mutual", "user1"},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record2})
		})

		Convey("query follow mutual for user2", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"_owner", "_follow", "mutual", "user2"},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(len(records), ShouldEqual, 0)
		})

		Convey("query friend mutual", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"_owner", "_friend", "mutual", "user1"},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))

			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record3})
		})

		Convey("distinct record satisfying both relations", func() {
			query := skydb.Query{
				Type: "record",
				Predicate: skydb.Predicate{
					Operator: skydb.Or,
					Children: []interface{}{
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: skydb.UserRelationFunc{"_owner", "_follow", "outward", "user4"},
								},
							},
						},
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: skydb.UserRelationFunc{"_owner", "_friend", "mutual", "user4"},
								},
							},
						},
					},
				},
				Sorts: sortsByID,
			}
			records, err := exhaustRows(db.Query(&query))
			So(err, ShouldBeNil)
			So(records, ShouldResemble, []skydb.Record{record0, record4})

			count, err := db.QueryCount(&query)
			So(err, ShouldBeNil)
			So(count, ShouldEqual, 2)
		})
	})
}

func TestGetByIDs(t *testing.T) {
	Convey("Database", t, func() {
		c := getTestConn(t)
		defer cleanupDB(t, c.Db)

		db := c.PrivateDB("getuser")
		So(db.Extend("record", skydb.RecordSchema{
			"string": skydb.FieldType{Type: skydb.TypeString},
		}), ShouldBeNil)

		insertRow(t, c.Db, `INSERT INTO app_com_oursky_skygear."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "string") `+
			`VALUES ('getuser', 'id0', 'getuser', '1988-02-06', 'getuser', '1988-02-06', 'getuser', 'string')`)
		insertRow(t, c.Db, `INSERT INTO app_com_oursky_skygear."record" `+
			`(_database_id, _id, _owner_id, _created_at, _created_by, _updated_at, _updated_by, "string") `+
			`VALUES ('getuser', 'id1', 'getuser', '1988-02-06', 'getuser', '1988-02-06', 'getuser', 'string')`)

		Convey("get one record", func() {
			scanner, err := db.GetByIDs([]skydb.RecordID{skydb.NewRecordID("record", "id1")})
			So(err, ShouldBeNil)

			scanner.Scan()
			record := scanner.Record()

			So(record.ID, ShouldResemble, skydb.NewRecordID("record", "id1"))
			So(record.DatabaseID, ShouldResemble, "getuser")
			So(record.OwnerID, ShouldResemble, "getuser")
			So(record.CreatorID, ShouldResemble, "getuser")
			So(record.UpdaterID, ShouldResemble, "getuser")
			So(record.Data["string"], ShouldEqual, "string")

			So(record.CreatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
			So(record.UpdatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
			noMore := scanner.Scan()
			So(noMore, ShouldEqual, false)
		})

		Convey("get one record with duplicated record ID", func() {
			scanner, err := db.GetByIDs([]skydb.RecordID{
				skydb.NewRecordID("record", "id1"),
				skydb.NewRecordID("record", "id1"),
			})
			So(err, ShouldBeNil)

			scanner.Scan()
			record := scanner.Record()

			So(record.ID, ShouldResemble, skydb.NewRecordID("record", "id1"))
			So(record.DatabaseID, ShouldResemble, "getuser")
			So(record.OwnerID, ShouldResemble, "getuser")
			So(record.CreatorID, ShouldResemble, "getuser")
			So(record.UpdaterID, ShouldResemble, "getuser")
			So(record.Data["string"], ShouldEqual, "string")

			So(record.CreatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
			So(record.UpdatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))

			noMore := scanner.Scan()
			So(noMore, ShouldEqual, false)
		})

		Convey("get record with one of them is placeholder", func() {
			scanner, err := db.GetByIDs([]skydb.RecordID{
				skydb.RecordID{},
				skydb.NewRecordID("record", "id1"),
			})
			So(err, ShouldBeNil)

			scanner.Scan()
			record := scanner.Record()

			So(record.ID, ShouldResemble, skydb.NewRecordID("record", "id1"))
			So(record.DatabaseID, ShouldResemble, "getuser")
			So(record.OwnerID, ShouldResemble, "getuser")
			So(record.CreatorID, ShouldResemble, "getuser")
			So(record.UpdaterID, ShouldResemble, "getuser")
			So(record.Data["string"], ShouldEqual, "string")

			So(record.CreatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))
			So(record.UpdatedAt, ShouldResemble, time.Date(1988, 2, 6, 0, 0, 0, 0, time.UTC))

			noMore := scanner.Scan()
			So(noMore, ShouldEqual, false)
		})

		Convey("get multiple record", func() {
			scanner, err := db.GetByIDs([]skydb.RecordID{
				skydb.NewRecordID("record", "id0"),
				skydb.NewRecordID("record", "id1"),
			})
			So(err, ShouldBeNil)

			scanner.Scan()
			record := scanner.Record()

			scanner.Scan()
			record2 := scanner.Record()
			So([]skydb.RecordID{
				record.ID,
				record2.ID,
			}, ShouldResemble, []skydb.RecordID{
				skydb.NewRecordID("record", "id0"),
				skydb.NewRecordID("record", "id1"),
			})

			noMore := scanner.Scan()
			So(noMore, ShouldEqual, false)
		})

	})
}
