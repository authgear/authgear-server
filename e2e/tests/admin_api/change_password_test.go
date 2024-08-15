package adminapi

import (
	"encoding/json"
	"fmt"
	"testing"

	relay "github.com/authgear/graphql-go-relay"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
)

func TestChangePassword(t *testing.T) {
	cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
		TestCase: &testrunner.TestCase{
			Path: "admin_api/change_password_test.go",
		},
		Test: t,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	username := "change_password_test"

	// Create user
	if err := cmd.ImportUsers("../../tests/admin_api/users.json"); err != nil {
		t.Fatalf(err.Error())
	}

	userID, originalPasswordHash, err := getPasswordHash(cmd, username)
	if err != nil {
		t.Fatalf(err.Error())
	}

	resp, err := cmd.Client.GraphQLAPI(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
		Query: `
			mutation resetPasswordMutation(
				$userID: ID!,
				$password: String!,
				$sendPassword: Boolean,
				$setPasswordExpired: Boolean
			) {
				resetPassword(input: {
					userID: $userID,
					password: $password,
					sendPassword: $sendPassword,
					setPasswordExpired: $setPasswordExpired
				}) {
					user {
						id
					}
				}
			}

		`,
		Variables: map[string]interface{}{
			"userID":             relay.ToGlobalID("User", userID),
			"password":           "new_password",
			"sendPassword":       true,
			"setPasswordExpired": true,
		},
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
	if resp.Errors != nil {
		t.Fatalf("error: %v", resp.Errors)
	}

	// Verify password updated
	_, newPasswordHash, err := getPasswordHash(cmd, username)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if originalPasswordHash == newPasswordHash {
		t.Fatalf("Password not updated")
	}

	// Verify email sent
	emailSent, err := verifyEmailInLog("[Authgear] Your password has been changed", username)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !emailSent {
		t.Fatalf("Reset password email with recipient '%s' not found.", username)
	}
}

func TestChangePasswordNoTarget(t *testing.T) {
	cmd, err := testrunner.NewEnd2EndCmd(testrunner.NewEnd2EndCmdOptions{
		TestCase: &testrunner.TestCase{
			Path: "admin_api/change_password_test.go",
		},
		Test: t,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	username := "change_password_test_no_target"

	// Create user
	if err := cmd.ImportUsers("../../tests/admin_api/users.json"); err != nil {
		t.Fatalf(err.Error())
	}

	userID, _, err := getPasswordHash(cmd, username)
	if err != nil {
		t.Fatalf(err.Error())
	}

	resp, err := cmd.Client.GraphQLAPI(nil, nil, cmd.AppID, e2eclient.GraphQLAPIRequest{
		Query: `
			mutation resetPasswordMutation(
				$userID: ID!,
				$password: String!,
				$sendPassword: Boolean,
				$setPasswordExpired: Boolean
			) {
				resetPassword(input: {
					userID: $userID,
					password: $password,
					sendPassword: $sendPassword,
					setPasswordExpired: $setPasswordExpired
				}) {
					user {
						id
					}
				}
			}

		`,
		Variables: map[string]interface{}{
			"userID":             relay.ToGlobalID("User", userID),
			"password":           "new_password",
			"sendPassword":       true,
			"setPasswordExpired": true,
		},
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	if resp.Errors == nil {
		t.Fatalf("expect error, got nil")
	}

	for _, err := range resp.Errors {
		if err.Extensions.Reason == "SendPasswordNoTarget" {
			return
		}
	}

	t.Fatalf("expect error no target to send the password, got %v", resp.Errors)
}

func getPasswordHash(cmd *testrunner.End2EndCmd, username string) (userID string, hash string, err error) {
	rawResult, err := cmd.QuerySQLSelectRaw(fmt.Sprintf(`
		SELECT expire_after, user_id, password_hash
		FROM _auth_authenticator
		JOIN _auth_authenticator_password ON _auth_authenticator_password.id = _auth_authenticator.id
		JOIN _auth_user ON _auth_authenticator.user_id = _auth_user.id
		WHERE _auth_authenticator.app_id = '%s'
		AND standard_attributes ->> 'preferred_username' = '%s';
	`, cmd.AppID, username))
	if err != nil {
		return "", "", err
	}

	var rows []interface{}
	err = json.Unmarshal([]byte(rawResult), &rows)
	if err != nil {
		return "", "", err
	}

	if len(rows) == 0 {
		return "", "", fmt.Errorf("password not found")
	}

	row := rows[0].(map[string]interface{})
	return row["user_id"].(string), row["password_hash"].(string), nil
}
