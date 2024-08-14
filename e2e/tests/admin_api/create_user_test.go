package adminapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/e2e/pkg/testrunner"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestCreateUserWithPassword(t *testing.T) {
	ctx := context.Background()

	cmd := testrunner.End2EndCmd{
		TestCase: testrunner.TestCase{
			Path: "admin_api/create_user_test.go",
		},
		Test: t,
	}
	appID, err := cmd.CreateConfigSource()
	if err != nil {
		t.Fatalf(err.Error())
	}

	cmd.AppID = appID

	userEmail := fmt.Sprintf("%s@example.com", appID)

	client := e2eclient.NewClient(
		ctx,
		"localhost:4000",
		"localhost:4002",
		httputil.HTTPHost(fmt.Sprintf("%s.portal.localhost:4000", appID)),
	)
	_, err = client.GraphQLAPI(nil, nil, appID, e2eclient.GraphQLAPIRequest{
		Query: `
			mutation createUserMutation(
				$identityDefinition: IdentityDefinitionLoginID!
				$password: String
				$sendPassword: Boolean
				$setPasswordExpired: Boolean
			) {
				createUser(
					input: {
						definition: { loginID: $identityDefinition },
						password: $password,
						sendPassword: $sendPassword,
						setPasswordExpired: $setPasswordExpired
					}
				) {
					user {
						id
					}
				}
			}
		`,
		Variables: map[string]interface{}{
			"identityDefinition": map[string]interface{}{
				"key":   "email",
				"value": userEmail,
			},
			"password":           "password",
			"sendPassword":       true,
			"setPasswordExpired": true,
		},
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Verify password created with expireAfter
	passwordCreated, err := verifyPasswordCreated(cmd, userEmail)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !passwordCreated {
		t.Fatalf("Password not created with expireAfter")
	}

	// Verify email sent
	emailSent, err := verifyEmailInLog("Get Started With Authgear", userEmail)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if !emailSent {
		t.Fatalf("Create user email with recipient '%s' not found.", userEmail)
	}
}

func verifyPasswordCreated(cmd testrunner.End2EndCmd, userEmail string) (bool, error) {
	rawResult, err := cmd.QuerySQLSelectRaw(fmt.Sprintf(`
		SELECT expire_after
		FROM _auth_authenticator
		JOIN _auth_authenticator_password ON _auth_authenticator_password.id = _auth_authenticator.id
		JOIN _auth_user ON _auth_authenticator.user_id = _auth_user.id
		WHERE _auth_authenticator.app_id = '%s'
		AND standard_attributes ->> 'email' = '%s';
	`, cmd.AppID, userEmail))
	if err != nil {
		return false, err
	}

	var rows []interface{}
	err = json.Unmarshal([]byte(rawResult), &rows)
	if err != nil {
		return false, err
	}

	if len(rows) == 0 {
		return false, nil
	}

	row := rows[0].(map[string]interface{})
	expireAfter := row["expire_after"]
	if expireAfter == nil {
		return false, fmt.Errorf("Password not created with expireAfter")
	}

	return true, nil
}

func verifyEmailInLog(subject string, recipient string) (bool, error) {
	file, err := os.Open("../../logs/e2e-smtp.log")
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Subject:["+subject) &&
			strings.Contains(line, "To:["+recipient) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}
