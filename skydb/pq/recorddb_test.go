package pq

import (
	"testing"

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
		c.AddRelation("user1", "follow", "user2")
		c.AddRelation("user1", "follow", "user3")
		c.AddRelation("user3", "follow", "user1")
		c.AddRelation("user1", "friend", "user4")
		c.AddRelation("user4", "friend", "user1")
		c.AddRelation("user4", "friend", "user5")
		c.AddRelation("user5", "friend", "user4")
		c.AddRelation("user4", "follow", "user5")

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
				Predicate: &skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: &skydb.UserRelationFunc{"_owner", "_follow", "outward", "user1"},
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
				Predicate: &skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: &skydb.UserRelationFunc{"_owner", "_follow", "inward", "user2"},
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
				Predicate: &skydb.Predicate{
					Operator: skydb.Or,
					Children: []interface{}{
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: &skydb.UserRelationFunc{"_owner", "_follow", "outward", "user1"},
								},
							},
						},
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: &skydb.UserRelationFunc{"_owner", "_follow", "inward", "user2"},
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
				Predicate: &skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: &skydb.UserRelationFunc{"_owner", "_follow", "mutual", "user1"},
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
				Predicate: &skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: &skydb.UserRelationFunc{"_owner", "_follow", "mutual", "user2"},
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
				Predicate: &skydb.Predicate{
					Operator: skydb.Functional,
					Children: []interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: &skydb.UserRelationFunc{"_owner", "_friend", "mutual", "user1"},
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
				Predicate: &skydb.Predicate{
					Operator: skydb.Or,
					Children: []interface{}{
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: &skydb.UserRelationFunc{"_owner", "_follow", "outward", "user4"},
								},
							},
						},
						skydb.Predicate{
							Operator: skydb.Functional,
							Children: []interface{}{
								skydb.Expression{
									Type:  skydb.Function,
									Value: &skydb.UserRelationFunc{"_owner", "_friend", "mutual", "user4"},
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
