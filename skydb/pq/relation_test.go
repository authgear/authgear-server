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

package pq

import (
	"testing"

	"github.com/skygeario/skygear-server/skydb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRelation(t *testing.T) {
	Convey("Conn", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		addUser(t, c, "userid")
		addUser(t, c, "friendid")

		Convey("add relation", func() {
			err := c.AddRelation("userid", "_friend", "friendid")
			So(err, ShouldBeNil)
		})

		Convey("add a user not exist relation", func() {
			err := c.AddRelation("userid", "_friend", "non-exist")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "userID not exist")
		})

		Convey("remove non-exist relation", func() {
			err := c.RemoveRelation("userid", "_friend", "friendid")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual,
				"_friend relation not exist {userid} => {friendid}")
		})

		Convey("remove relation", func() {
			err := c.AddRelation("userid", "_friend", "friendid")
			So(err, ShouldBeNil)
			err = c.RemoveRelation("userid", "_friend", "friendid")
			So(err, ShouldBeNil)
		})
	})

	Convey("Conn Query", t, func() {
		c := getTestConn(t)
		defer cleanupConn(t, c)

		addUser(t, c, "follower")
		addUser(t, c, "followee")
		addUser(t, c, "friend1")
		addUser(t, c, "friend2")
		addUser(t, c, "friend3")
		c.AddRelation("friend1", "_friend", "friend2")
		c.AddRelation("friend1", "_friend", "friend3")
		c.AddRelation("friend2", "_friend", "friend1")
		c.AddRelation("friend3", "_friend", "friend1")
		c.AddRelation("friend1", "_friend", "followee")
		c.AddRelation("follower", "_follow", "followee")

		Convey("query friend relation", func() {
			users := c.QueryRelation("friend1", "_friend", "mutual", skydb.QueryConfig{})
			So(len(users), ShouldEqual, 2)
		})

		Convey("query outward follow relation", func() {
			users := c.QueryRelation("follower", "_follow", "outward", skydb.QueryConfig{})
			So(len(users), ShouldEqual, 1)
		})

		Convey("query inward follow relation", func() {
			users := c.QueryRelation("followee", "_follow", "inward", skydb.QueryConfig{})
			So(len(users), ShouldEqual, 1)
		})

		Convey("query friend relation with pagination", func() {
			users := c.QueryRelation("friend1", "_friend", "mutual", skydb.QueryConfig{
				Limit: 1,
			})
			So(len(users), ShouldEqual, 1)
			So(users[0].ID, ShouldEqual, "friend2")

			users = c.QueryRelation("friend1", "_friend", "mutual", skydb.QueryConfig{
				Limit:  1,
				Offset: 1,
			})
			So(len(users), ShouldEqual, 1)
			So(users[0].ID, ShouldEqual, "friend3")
		})
	})
}
