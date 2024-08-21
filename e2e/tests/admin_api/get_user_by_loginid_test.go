package adminapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func TestGetUserByLoginIDByEmail(t *testing.T) {
	Convey("GetUserByLoginIDByEmail", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_loginid_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		// Prepare testing user
		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/get_users_admin_api_user.sql")
		So(err, ShouldBeNil)

		rawResp, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
				query {
					getUserByLoginID(loginIDKey:"email", loginIDValue: "e2e_admin_api_get_users@example.com") {
							createdAt
							standardAttributes
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByLoginID": {
						"createdAt": "2024-08-27T07:51:42.040683Z",
						"lastLoginAt": null,
						"standardAttributes": {
							"email": "e2e_admin_api_get_users@example.com",
							"email_verified": false,
							"updated_at": 1724745102
						}
					}
				}
			}
		`)
	})
}
func TestGetUserByLoginIDByPhone(t *testing.T) {
	Convey("GetUserByLoginIDByPhone", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_loginid_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		// Prepare testing user
		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/get_users_admin_api_user.sql")
		So(err, ShouldBeNil)

		rawResp, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
				query {
					getUserByLoginID(loginIDKey:"phone", loginIDValue: "+85261236544") {
							createdAt
							standardAttributes
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByLoginID": {
						"createdAt": "2024-08-27T07:51:42.040683Z",
						"lastLoginAt": null,
						"standardAttributes": {
							"phone_number": "+85261236544",
							"phone_number_verified": false,
							"updated_at": 1724745102
						}
					}
				}
			}
		`)
	})
}

func TestGetUserByLoginIDByUsername(t *testing.T) {
	Convey("GetUserByLoginIDByUsername", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_loginid_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		// Prepare testing user
		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/get_users_admin_api_user.sql")
		So(err, ShouldBeNil)

		rawResp, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
				query {
					getUserByLoginID(loginIDKey:"username", loginIDValue: "e2e_admin_api_get_users_username") {
							createdAt
							standardAttributes
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByLoginID": {
						"createdAt": "2024-08-27T07:51:42.040683Z",
						"lastLoginAt": null,
						"standardAttributes": {
							"preferred_username": "e2e_admin_api_get_users_username",
							"updated_at": 1724745102
						}
					}
				}
			}
		`)
	})
}

func TestGetUserByLoginIDNotFound(t *testing.T) {
	Convey("GetUserByLoginIDNotFound", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_loginid_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		// Prepare testing user
		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/get_users_admin_api_user.sql")
		So(err, ShouldBeNil)

		rawResp, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
				query {
					getUserByLoginID(loginIDKey:"email", loginIDValue: "e2e_admin_api_get_users_not_found@example.com") {
							createdAt
							standardAttributes
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByLoginID": null
				}
			}
		`)
	})
}

func TestGetUserByLoginIDInvalidLoginIDValue(t *testing.T) {
	Convey("GetUserByLoginIDInvalidLoginIDValue", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_loginid_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		// Prepare testing user
		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/get_users_admin_api_user.sql")
		So(err, ShouldBeNil)

		rawResp, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
				query {
					getUserByLoginID(loginIDKey:"email", loginIDValue: "bad_email_address") {
							id
							createdAt
							standardAttributes
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByLoginID": null
				},
				"errors": [
					{
						"message": "invalid Login ID key",
						"locations": [
							{
								"line": 3,
								"column": 6
							}
						],
						"path": [
							"getUserByLoginID"
						],
						"extensions": {
							"errorName": "Invalid",
							"reason": "GetUsersInvalidArgument"
						}
					}
				]
			}
		`)
	})
}

func TestGetUserByLoginIDInvalidLoginIDKey(t *testing.T) {
	Convey("GetUserByLoginIDInvalidLoginIDKey", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_loginid_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		// Prepare testing user
		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/get_users_admin_api_user.sql")
		So(err, ShouldBeNil)

		rawResp, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
				query {
					getUserByLoginID(loginIDKey:"email_invalid", loginIDValue: "e2e_admin_api_get_users_not_found@example.com") {
							id
							createdAt
							standardAttributes
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByLoginID": null
				},
				"errors": [
					{
						"message": "invalid Login ID key",
						"locations": [
							{
								"line": 3,
								"column": 6
							}
						],
						"path": [
							"getUserByLoginID"
						],
						"extensions": {
							"errorName": "Invalid",
							"reason": "GetUsersInvalidArgument"
						}
					}
				]
			}
		`)
	})
}
