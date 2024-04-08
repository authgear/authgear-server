package testrunner

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

	authflowclient "github.com/authgear/authgear-server/e2e/pkg/e2eclient"
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

	// Create project per test case
	err := e2eCmd.CreateConfigSource()
	if err != nil {
		t.Errorf("failed to create config source: %v", err)
		return
	}

	// Execute before hooks to prepare fixtures
	ok := executeBeforeAll(t, e2eCmd, testCase)
	if !ok {
		return
	}

	client := authflowclient.NewClient(
		ctx,
		"localhost:4000",
		httputil.HTTPHost(fmt.Sprintf("%s.portal.localhost:4000", appID)),
	)

	var state string

	for i, step := range testCase.Steps {
		if step.Name == "" {
			step.Name = fmt.Sprintf("step %d", i+1)
		}

		state, ok = executeStep(t, e2eCmd, client, step, state)
		if !ok {
			return
		}
	}
}

func executeBeforeAll(t *testing.T, e2eCmd *End2EndCmd, testCase TestCase) (ok bool) {
	for _, beforeHook := range testCase.Before {
		switch beforeHook.Type {
		case BeforeHookTypeUserImport:
			err := e2eCmd.ImportUsers(beforeHook.UserImport)
			if err != nil {
				t.Errorf("failed to import users: %v", err)
				return false
			}
		case BeforeHookTypeCustomSQL:
			err := e2eCmd.ExecuteCustomSQL(beforeHook.CustomSQL.Path)
			if err != nil {
				t.Errorf("failed to execute custom SQL: %v", err)
				return false
			}
		default:
			t.Errorf("unknown before hook type: %s", beforeHook.Type)
			return false
		}
	}

	return true
}

func executeStep(t *testing.T, e2eCmd *End2EndCmd, client *authflowclient.Client, step Step, state string) (nextState string, ok bool) {
	var flowResponse *authflowclient.FlowResponse
	var flowErr error

	nextState = state

	switch step.Action {
	case StepActionCreate:
		var flowReference authflowclient.FlowReference
		err := json.Unmarshal([]byte(step.Input), &flowReference)
		if err != nil {
			t.Errorf("failed to parse input in '%s': %v\n", step.Name, err)
			return
		}

		flowResponse, flowErr = client.Create(flowReference, "")

	case StepActionInput:
		fallthrough
	default:
		var input map[string]interface{}
		err := json.Unmarshal([]byte(step.Input), &input)
		if err != nil {
			t.Errorf("failed to parse JSON input in '%s': %v\n", step.Name, err)
			return
		}

		flowResponse, flowErr = client.Input(nil, nil, state, input)
	}

	if step.Output != nil {
		ok := validateOutput(t, step, flowResponse, flowErr)
		if !ok {
			return "", false
		}
	}

	if flowResponse != nil {
		nextState = flowResponse.StateToken
	}

	return nextState, true
}

func validateOutput(t *testing.T, step Step, flowResponse *authflowclient.FlowResponse, flowErr error) (ok bool) {
	errorViolations, resultViolations, err := MatchOutput(*step.Output, flowResponse, flowErr)
	if err != nil {
		t.Errorf("failed to match output in '%s': %v\n", step.Name, err)
		t.Errorf("  result: %v\n", flowResponse)
		t.Errorf("  error: %v\n", flowErr)
		return false
	}

	if len(errorViolations) > 0 {
		t.Errorf("error output mismatch in '%s': %v\n", step.Name, flowErr)
		for _, violation := range errorViolations {
			t.Errorf("  %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		return false
	}

	if len(resultViolations) > 0 {
		t.Errorf("result output mismatch in '%s': %v\n", step.Name, flowResponse)
		for _, violation := range resultViolations {
			t.Errorf("  %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		return false
	}

	return true
}

func generateAppID() string {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(id)
}
