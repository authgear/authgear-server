package handler

import (
	"testing"

	"github.com/oursky/skygear/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQueryFromRaw(t *testing.T) {
	Convey("functional predicate", t, func() {
		Convey("functional predicate with user relation", func() {
			parser := &QueryParser{
				UserID: "USER_ID",
			}
			query := skydb.Query{}
			err := parser.queryFromRaw(map[string]interface{}{
				"record_type": "note",
				"predicate": []interface{}{
					"func",
					"userRelation",
					map[string]interface{}{"$type": "keypath", "$val": "assignee"},
					map[string]interface{}{"$type": "relation", "$name": "_follow", "$direction": "outward"},
				},
			}, &query)
			So(err, ShouldBeNil)
			So(query, ShouldResemble, skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					skydb.Functional,
					[]interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"assignee", "_follow", "outward", "USER_ID"},
						},
					},
				},
			})
		})

		Convey("functional predicate with user friend relation", func() {
			parser := &QueryParser{
				UserID: "USER_ID",
			}
			query := skydb.Query{}
			err := parser.queryFromRaw(map[string]interface{}{
				"record_type": "note",
				"predicate": []interface{}{
					"func",
					"userRelation",
					map[string]interface{}{"$type": "keypath", "$val": "_owner"},
					map[string]interface{}{"$type": "relation", "$name": "_friend", "$direction": "mutual"},
				},
			}, &query)
			So(err, ShouldBeNil)
			So(query, ShouldResemble, skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					skydb.Functional,
					[]interface{}{
						skydb.Expression{
							Type:  skydb.Function,
							Value: skydb.UserRelationFunc{"_owner", "_friend", "mutual", "USER_ID"},
						},
					},
				},
			})
		})
	})

}
