package tests

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authflowclient "github.com/authgear/authgear-server/e2e/tests/authflow/client"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"gopkg.in/yaml.v2"
)

func LoadAllTestCases(path string) ([]TestCase, error) {
	var testCases []TestCase
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		decoder := yaml.NewDecoder(strings.NewReader(string(data)))
		for {
			var testCase TestCase
			err := decoder.Decode(&testCase)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			testCase.Path = path
			testCases = append(testCases, testCase)
		}

		return nil
	})

	return testCases, err
}

func RunTestCase(t *testing.T, testCase TestCase) {
	t.Logf("running test case: %s\n", testCase.Name)

	ctx := context.Background()

	appID := generateAppID()
	e2eCmd := &End2EndCmd{
		AppID:    appID,
		TestCase: testCase,
	}

	err := e2eCmd.CreateConfigSource()
	if err != nil {
		t.Errorf("failed to create config source: %v", err)
		return
	}

	for _, beforeHook := range testCase.Before {
		switch beforeHook.Type {
		case BeforeHookTypeUserImport:
			err = e2eCmd.ImportUsers(beforeHook.UserImport)
			if err != nil {
				t.Errorf("failed to import users: %v", err)
				return
			}
		case BeforeHookTypeCustomSQL:
			err = e2eCmd.ExecuteCustomSQL(beforeHook.CustomSQL.Path)
			if err != nil {
				t.Errorf("failed to execute custom SQL: %v", err)
				return
			}
		default:
			t.Errorf("unknown before hook type: %s", beforeHook.Type)
		}
	}

	hostName := httputil.HTTPHost(fmt.Sprintf("%s.portal.localhost:4000", appID))
	client := authflowclient.NewClient(ctx, "localhost:4000", hostName)

	var stateToken string

	for i, step := range testCase.Steps {
		var stepName = step.Name
		if stepName == "" {
			stepName = fmt.Sprintf("step %d", i+1)
		}

		var flowResponse *authflowclient.FlowResponse
		var flowErr error

		switch step.Action {
		case StepActionCreate:
			var flowReference authflowclient.FlowReference
			err := json.Unmarshal([]byte(step.Input), &flowReference)
			if err != nil {
				t.Errorf("failed to parse input in '%s': %v\n", stepName, err)
				return
			}

			flowResponse, flowErr = client.Create(flowReference, "")

		case StepActionInput:
			fallthrough

		default:
			var input map[string]interface{}
			err := json.Unmarshal([]byte(step.Input), &input)
			if err != nil {
				t.Errorf("failed to parse JSON input in '%s': %v\n", stepName, err)
				return
			}

			flowResponse, flowErr = client.Input(nil, nil, stateToken, input)
		}

		if flowResponse != nil {
			stateToken = flowResponse.StateToken
		}

		if step.Output != nil {
			errorViolations, resultViolations, err := MatchOutput(*step.Output, flowResponse, flowErr)
			if err != nil {
				t.Errorf("failed to match output in '%s': %v\n", stepName, err)
				t.Errorf("  result: %v\n", flowResponse)
				t.Errorf("  error: %v\n", flowErr)
				return
			}
			if len(errorViolations) > 0 {
				t.Errorf("error output mismatch in '%s': %v\n", stepName, flowErr)
				for _, violation := range errorViolations {
					t.Errorf("  %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
				}
				return
			}
			if len(resultViolations) > 0 {
				t.Errorf("result output mismatch in '%s': %v\n", stepName, flowResponse)
				for _, violation := range resultViolations {
					t.Errorf("  %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
				}
				return
			}
		}
	}
}

func generateAppID() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(id)
}
