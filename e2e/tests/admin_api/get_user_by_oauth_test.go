package adminapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func TestGetUserByOAuthFound(t *testing.T) {
	Convey("GetUserByOAuthFound", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_oauth_test.go",
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
					getUserByOAuth(oauthProviderAlias:"azureadv2", oauthProviderUserID: "e2e_admin_api_get_user_oauth_id") {
							createdAt
							formattedName
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByOAuth": {
						"createdAt": "2024-08-27T07:51:42.040683Z",
						"formattedName": "e2e OAuth user",
						"lastLoginAt": null
					}
				}
			}
		`)
	})
}

func TestGetUserByOAuthNotFound(t *testing.T) {
	Convey("GetUserByOAuthNotFound", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_oauth_test.go",
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
					getUserByOAuth(oauthProviderAlias:"azureadv2", oauthProviderUserID: "e2e_admin_api_get_user_oauth_not_found") {
							createdAt
							formattedName
							lastLoginAt
					}
				}
	
			`,
		})
		So(err, ShouldBeNil)
		So(rawResp, ShouldEqualJSON, `
			{
				"data": {
					"getUserByOAuth": null
				}
			}
		`)
	})
}

func TestGetUserByOAuthNotInvalidOAuthAlias(t *testing.T) {
	Convey("GetUserByOAuthNotInvalidOAuthAlias", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/get_user_by_oauth_test.go",
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
					getUserByOAuth(oauthProviderAlias:"unknown_alias", oauthProviderUserID: "e2e_admin_api_get_user_oauth_id") {
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
					"getUserByOAuth": null
				},
				"errors": [
					{
						"message": "invalid OAuth provider alias",
						"locations": [
							{
								"line": 3,
								"column": 6
							}
						],
						"path": [
							"getUserByOAuth"
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
