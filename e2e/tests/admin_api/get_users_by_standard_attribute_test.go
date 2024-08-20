package adminapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func TestGetUsersByStandardAttributeByEmail(t *testing.T) {
	Convey("GetUsersByStandardAttributeByEmai", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_users_by_standard_attribute_test.go",
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
					getUsersByStandardAttribute(attributeName:"email", attributeValue: "e2e_admin_api_get_users@example.com") {
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
					"getUsersByStandardAttribute": [
						{
							"createdAt": "2024-08-27T07:51:42.040683Z",
							"lastLoginAt": null,
							"standardAttributes": {
								"email": "e2e_admin_api_get_users@example.com",
								"email_verified": false,
								"updated_at": 1724745102
							}
						}
					]
				}
			}
		`)
	})
}
func TestGetUsersByStandardAttributeByPhone(t *testing.T) {
	Convey("GetUsersByStandardAttributeByPhone", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_users_by_standard_attribute_test.go",
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
					getUsersByStandardAttribute(attributeName:"phone_number", attributeValue: "+85261236544") {
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
					"getUsersByStandardAttribute": [
						{
							"createdAt": "2024-08-27T07:51:42.040683Z",
							"lastLoginAt": null,
							"standardAttributes": {
								"phone_number": "+85261236544",
								"phone_number_verified": false,
								"updated_at": 1724745102
							}
						}
					]
				}
			}
		`)
	})
}

func TestGetUsersByStandardAttributeByUsername(t *testing.T) {
	Convey("GetUsersByStandardAttributeByUsername", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_users_by_standard_attribute_test.go",
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
					getUsersByStandardAttribute(attributeName:"preferred_username", attributeValue: "e2e_admin_api_get_users_username") {
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
					"getUsersByStandardAttribute": [
						{
							"createdAt": "2024-08-27T07:51:42.040683Z",
							"lastLoginAt": null,
							"standardAttributes": {
								"preferred_username": "e2e_admin_api_get_users_username",
								"updated_at": 1724745102
							}
						}
					]
				}
			}
		`)
	})
}

func TestGetUsersByStandardAttributeNotFound(t *testing.T) {
	Convey("GetUsersByStandardAttributeNotFound", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_users_by_standard_attribute_test.go",
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
					getUsersByStandardAttribute(attributeName:"email", attributeValue: "e2e_admin_api_get_users_not_found@example.com") {
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
					"getUsersByStandardAttribute": []
				}
			}
		`)
	})
}

func TestGetUsersByStandardAttributeInvalidAttributeValue(t *testing.T) {
	Convey("GetUsersByStandardAttributeInvalidAttributeValue", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_users_by_standard_attribute_test.go",
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
					getUsersByStandardAttribute(attributeName:"phone_number", attributeValue: "e2e_admin_api_get_users@example.com") {
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
				"data": null,
				"errors": [
					{
						"extensions": {
							"errorName": "Invalid",
							"reason": "GetUsersInvalidArgument"
						},
						"locations": [{ "column": 6, "line": 3 }],
						"message": "invalid attributeValue",
						"path": ["getUsersByStandardAttribute"]
					}
				]
			}
		`)
	})
}

func TestGetUsersByStandardAttributeInvalidAttributeName(t *testing.T) {
	Convey("GetUsersByStandardAttributeInvalidAttributeName", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_users_by_standard_attribute_test.go",
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
					getUsersByStandardAttribute(attributeName:"custom_name", attributeValue: "e2e_admin_api_get_users@example.com") {
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
				"data": null,
				"errors": [
					{
						"extensions": {
							"errorName": "Invalid",
							"reason": "GetUsersInvalidArgument"
						},
						"locations": [{ "column": 6, "line": 3 }],
						"message": "attributeName must be email, phone_number or preferred_username",
						"path": ["getUsersByStandardAttribute"]
					}
				]
			}
		`)
	})
}
