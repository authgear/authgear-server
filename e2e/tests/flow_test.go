package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Name    string `yaml:"name"`
	Project string `yaml:"project"`
	Steps   []Step `yaml:"steps"`
}

type StepAction string

var (
	StepActionCreate StepAction = "create"
	StepActionInput  StepAction = "input"
)

type Step struct {
	Action StepAction `yaml:"action"`
	Input  string     `yaml:"input"`
	Assert []Assert   `yaml:"assert"`
}

type AssertField string

var (
	AssertFieldType       AssertField = "type"
	AssertFieldActionType AssertField = "action.type"
	AssertFieldStateToken AssertField = "state_token"
)

type AssertOp string

var (
	AssertOpEq AssertOp = "eq"
)

type Assert struct {
	Field AssertField `yaml:"field"`
	Op    AssertOp    `yaml:"op"`
	Value string      `yaml:"value"`
}

func TestCases(t *testing.T) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".yaml" {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}

			var testCase TestCase
			err = yaml.Unmarshal(data, &testCase)
			if err != nil {
				t.Fatal(err)
			}

			t.Logf("Running test case: %s", testCase.Name)

			client := authflowclient.NewClient(context.Background(), "localhost:4000", httputil.HTTPHost(fmt.Sprintf("%s.portal.localhost:4000", testCase.Project)))

			var stateToken string

			for i, step := range testCase.Steps {
				stepYAML, err := yaml.Marshal(step)
				if err != nil {
					t.Fatalf("Failed to marshal step %d to YAML: %v", i+1, err)
				}

				var flowResponse *authflowclient.FlowResponse

				if step.Action == "create" {
					var flowReference authflowclient.FlowReference
					err := json.Unmarshal([]byte(step.Input), &flowReference)
					if err != nil {
						t.Fatalf("Failed to parse input in step %d: %v\n%s", i+1, err, stepYAML)
					}

					flowResponse, err = client.Create(flowReference, "")
					if err != nil {
						t.Fatalf("Failed to create flow in step %d: %v\n%s", i+1, err, stepYAML)
					}

					stateToken = flowResponse.StateToken
				} else if step.Action == "input" || step.Action == "" {
					var input map[string]interface{}
					err := json.Unmarshal([]byte(step.Input), &input)
					if err != nil {
						t.Fatalf("Failed to parse JSON input in step %d: %v\n%s", i+1, err, stepYAML)
					}

					flowResponse, err = client.Input(nil, nil, stateToken, input)
					if err != nil {
						t.Fatalf("Failed to input in step %d: %v\n%s", i+1, err, stepYAML)
					}

					stateToken = flowResponse.StateToken
				} else {
					t.Fatalf("Unknown action in step %d: %s\n%s", i+1, step.Action, stepYAML)
				}

				for _, assertion := range step.Assert {
					value, ok := GetFlowResponseValue(flowResponse, assertion.Field)
					if !ok {
						t.Errorf("Field '%s' not found in step %d\n%s", assertion.Field, i+1, stepYAML)
						continue
					}

					valueStr, ok := value.(string)
					if !ok {
						t.Errorf("Field '%s' in step %d is not a string\n%s", assertion.Field, i+1, stepYAML)
						continue
					}

					if assertion.Op == "eq" && valueStr != assertion.Value {
						t.Errorf("Assertion failed in step %d: Expected '%s' to be '%s'\n%s", i+1, valueStr, assertion.Value, stepYAML)
					}
				}
			}
		})

		return nil
	})

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func GetFlowResponseValue(flowResponse *authflowclient.FlowResponse, field AssertField) (interface{}, bool) {
	switch field {
	case AssertFieldType:
		return string(flowResponse.Type), true
	case AssertFieldActionType:
		return string(flowResponse.Action.Type), true
	case AssertFieldStateToken:
		return string(flowResponse.StateToken), true
	default:
		return nil, false
	}
}
