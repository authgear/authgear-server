package adminapi

import (
	"encoding/json"
	"fmt"
	"testing"

	relay "github.com/authgear/graphql-go-relay"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func TestCreateIdentity_CreateOOBOTPEmailIfEnabled(t *testing.T) {
	Convey("CreateIdentity_CreateOOBOTPEmailIfEnabled", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/create_identity_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/create_identity_user.sql")
		So(err, ShouldBeNil)

		userID, err := createIdentity_getUserID(cmd)
		So(err, ShouldBeNil)

		nodeUserID := relay.ToGlobalID("User", userID)

		rawResponse, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
mutation createIdentityMutation(
	$userID: ID!
	$identityDefinition: IdentityDefinitionLoginID!
	$password: String
) {
	createIdentity(
		input: {
			userID: $userID
			definition: { loginID: $identityDefinition }
			password: $password
		}
	) {
		user {
			standardAttributes
			identities {
				edges {
					node {
						type
						claims
					}
				}
			}
			authenticators {
				edges {
					node {
						kind
						type
						claims
					}
				}
			}
		}
	}
}
			`,
			Variables: map[string]interface{}{
				"userID": nodeUserID,
				"identityDefinition": map[string]interface{}{
					"key":   "email",
					"value": "user@example.com",
				},
			},
		})
		So(err, ShouldBeNil)
		So(createIdentity_removeUpdatedAtInStandardAttributes(rawResponse), ShouldEqualJSON, `
{
    "data": {
        "createIdentity": {
            "user": {
                "authenticators": {
                    "edges": [
                        {
                            "node": {
                                "claims": {
                                    "https://authgear.com/claims/oob_otp/email": "user@example.com"
                                },
                                "kind": "PRIMARY",
                                "type": "OOB_OTP_EMAIL"
                            }
                        }
                    ]
                },
                "identities": {
                    "edges": [
                        {
                            "node": {
                                "claims": {
                                    "email": "user@example.com",
                                    "https://authgear.com/claims/login_id/key": "email",
                                    "https://authgear.com/claims/login_id/original_value": "user@example.com",
                                    "https://authgear.com/claims/login_id/type": "email",
                                    "https://authgear.com/claims/login_id/value": "user@example.com"
                                },
                                "type": "LOGIN_ID"
                            }
                        }
                    ]
                },
                "standardAttributes": {
                    "email": "user@example.com",
                    "email_verified": false
                }
            }
        }
    }
}
`)
	})
}

func TestCreateIdentity_CreateOOBOTPSMSIfEnabled(t *testing.T) {
	Convey("CreateIdentity_CreateOOBOTPSMSIfEnabled", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/create_identity_test.go",
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/create_identity_user.sql")
		So(err, ShouldBeNil)

		userID, err := createIdentity_getUserID(cmd)
		So(err, ShouldBeNil)

		nodeUserID := relay.ToGlobalID("User", userID)

		rawResponse, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
mutation createIdentityMutation(
	$userID: ID!
	$identityDefinition: IdentityDefinitionLoginID!
	$password: String
) {
	createIdentity(
		input: {
			userID: $userID
			definition: { loginID: $identityDefinition }
			password: $password
		}
	) {
		user {
			standardAttributes
			identities {
				edges {
					node {
						type
						claims
					}
				}
			}
			authenticators {
				edges {
					node {
						kind
						type
						claims
					}
				}
			}
		}
	}
}
			`,
			Variables: map[string]interface{}{
				"userID": nodeUserID,
				"identityDefinition": map[string]interface{}{
					"key":   "phone",
					"value": "+85298765432",
				},
			},
		})
		So(err, ShouldBeNil)
		So(createIdentity_removeUpdatedAtInStandardAttributes(rawResponse), ShouldEqualJSON, `
{
    "data": {
        "createIdentity": {
            "user": {
                "authenticators": {
                    "edges": [
                        {
                            "node": {
                                "claims": {
                                    "https://authgear.com/claims/oob_otp/phone": "+85298765432"
                                },
                                "kind": "PRIMARY",
                                "type": "OOB_OTP_SMS"
                            }
                        }
                    ]
                },
                "identities": {
                    "edges": [
                        {
                            "node": {
                                "claims": {
                                    "https://authgear.com/claims/login_id/key": "phone",
                                    "https://authgear.com/claims/login_id/original_value": "+85298765432",
                                    "https://authgear.com/claims/login_id/type": "phone",
                                    "https://authgear.com/claims/login_id/value": "+85298765432",
                                    "phone_number": "+85298765432"
                                },
                                "type": "LOGIN_ID"
                            }
                        }
                    ]
                },
                "standardAttributes": {
                    "phone_number": "+85298765432",
                    "phone_number_verified": false
                }
            }
        }
    }
}
`)
	})
}

func TestCreateIdentity_CreatePasswordIfGiven(t *testing.T) {
	Convey("CreateIdentity_CreatePasswordIfGiven", t, func() {
		cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
			TestCase: &testrunner.TestCase{
				Path: "admin_api/create_identity_test.go",
				AuthgearYAMLSource: testrunner.AuthgearYAMLSource{
					Override: `

authentication:
  primary_authenticators:
  - password
`,
				},
			},
			Test: t,
		})
		So(err, ShouldBeNil)

		err = cmd.ExecuteSQLInsertUpdateFile("../../tests/admin_api/create_identity_user.sql")
		So(err, ShouldBeNil)

		userID, err := createIdentity_getUserID(cmd)
		So(err, ShouldBeNil)

		nodeUserID := relay.ToGlobalID("User", userID)

		rawResponse, err := cmd.Client.GraphQLAPIRaw(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
			Query: `
mutation createIdentityMutation(
	$userID: ID!
	$identityDefinition: IdentityDefinitionLoginID!
	$password: String
) {
	createIdentity(
		input: {
			userID: $userID
			definition: { loginID: $identityDefinition }
			password: $password
		}
	) {
		user {
			standardAttributes
			identities {
				edges {
					node {
						type
						claims
					}
				}
			}
			authenticators {
				edges {
					node {
						kind
						type
						claims
					}
				}
			}
		}
	}
}
			`,
			Variables: map[string]interface{}{
				"userID": nodeUserID,
				"identityDefinition": map[string]interface{}{
					"key":   "email",
					"value": "user@example.com",
				},
				"password": "password",
			},
		})
		So(err, ShouldBeNil)
		So(createIdentity_removeUpdatedAtInStandardAttributes(rawResponse), ShouldEqualJSON, `
{
    "data": {
        "createIdentity": {
            "user": {
                "authenticators": {
                    "edges": [
                        {
                            "node": {
                                "claims": {},
                                "kind": "PRIMARY",
                                "type": "PASSWORD"
                            }
                        }
                    ]
                },
                "identities": {
                    "edges": [
                        {
                            "node": {
                                "claims": {
                                    "email": "user@example.com",
                                    "https://authgear.com/claims/login_id/key": "email",
                                    "https://authgear.com/claims/login_id/original_value": "user@example.com",
                                    "https://authgear.com/claims/login_id/type": "email",
                                    "https://authgear.com/claims/login_id/value": "user@example.com"
                                },
                                "type": "LOGIN_ID"
                            }
                        }
                    ]
                },
                "standardAttributes": {
                    "email": "user@example.com",
                    "email_verified": false
                }
            }
        }
    }
}
`)
	})
}

func createIdentity_getUserID(cmd *testrunner.End2EndCmd) (userID string, err error) {
	rawResult, err := cmd.QuerySQLSelectRaw(fmt.Sprintf(`
		SELECT id
		FROM _auth_user
		WHERE app_id = '%s';
	`, cmd.AppID))
	if err != nil {
		return "", err
	}

	var rows []interface{}
	err = json.Unmarshal([]byte(rawResult), &rows)
	if err != nil {
		return "", err
	}

	if len(rows) == 0 {
		return "", fmt.Errorf("user not found")
	}

	row := rows[0].(map[string]interface{})
	return row["id"].(string), nil
}

func createIdentity_removeUpdatedAtInStandardAttributes(rawJSON string) (out string) {
	var o map[string]interface{}
	err := json.Unmarshal([]byte(rawJSON), &o)
	if err != nil {
		panic(err)
	}

	stdAttrs := o["data"].(map[string]interface{})["createIdentity"].(map[string]interface{})["user"].(map[string]interface{})["standardAttributes"].(map[string]interface{})
	delete(stdAttrs, "updated_at")

	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}

	out = string(b)
	return
}
