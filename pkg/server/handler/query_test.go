// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"testing"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestQueryFromRaw(t *testing.T) {
	Convey("QueryParser", t, func() {
		parser := &QueryParser{
			UserID: "USER_ID",
		}

		Convey("should parse simple predicate with multi-components keypath", func() {
			query := skydb.Query{}
			err := parser.queryFromRaw(map[string]interface{}{
				"record_type": "note",
				"predicate": []interface{}{
					"eq",
					map[string]interface{}{"$type": "keypath", "$val": "category.name"},
					"Interesting",
				},
			}, &query)
			So(err, ShouldBeNil)
			So(query, ShouldResemble, skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					skydb.Equal,
					[]interface{}{
						skydb.Expression{
							Type:  skydb.KeyPath,
							Value: "category.name",
						},
						skydb.Expression{
							Type:  skydb.Literal,
							Value: "Interesting",
						},
					},
				},
			})
		})

		Convey("functional predicate with user relation", func() {
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

		Convey("functional predicate with user discover", func() {
			query := skydb.Query{}
			err := parser.queryFromRaw(map[string]interface{}{
				"record_type": "note",
				"predicate": []interface{}{
					"func",
					"userDiscover",
					map[string]interface{}{
						"usernames": []string{
							"john.doe",
							"jane.doe",
						},
						"emails": []string{
							"john.doe@example.com",
							"jane.doe@example.com",
						},
					},
				},
			}, &query)
			So(err, ShouldBeNil)
			So(query, ShouldResemble, skydb.Query{
				Type: "note",
				Predicate: skydb.Predicate{
					skydb.Functional,
					[]interface{}{
						skydb.Expression{
							Type: skydb.Function,
							Value: skydb.UserDiscoverFunc{
								Usernames: []string{
									"john.doe",
									"jane.doe",
								},
								Emails: []string{
									"john.doe@example.com",
									"jane.doe@example.com",
								},
							},
						},
					},
				},
			})
		})
	})

}
