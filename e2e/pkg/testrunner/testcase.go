package testrunner

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/google/uuid"

	authflowclient "github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

var _ = TestCaseSchema.Add("TestCase", `
{
	"type": "object",
	"properties": {
		"name": { "type": "string" },
		"focus": { "type": "boolean" },
		"authgear.yaml": { "$ref": "#/$defs/AuthgearYAMLSource" },
		"steps": { "type": "array", "items": { "$ref": "#/$defs/Step" } },
		"before": { "type": "array", "items": { "$ref": "#/$defs/BeforeHook" } }
	},
	"required": ["name", "steps"]
}
`)

type TestCase struct {
	Name string `json:"name"`
	Path string `json:"path"`
	// Applying focus to a test case will make it the only test case to run,
	// mainly used for debugging new test cases.
	Focus              bool               `json:"focus"`
	AuthgearYAMLSource AuthgearYAMLSource `json:"authgear.yaml"`
	Steps              []Step             `json:"steps"`
	Before             []BeforeHook       `json:"before"`
}

func (tc *TestCase) FullName() string {
	return tc.Path + "/" + tc.Name
}

func (tc *TestCase) Run(t *testing.T) {
	ctx := context.Background()

	appID := generateAppID()
	cmd := &End2EndCmd{
		AppID:    appID,
		TestCase: *tc,
		Test:     t,
	}

	// Create project per test case
	err := cmd.CreateConfigSource()
	if err != nil {
		t.Errorf("failed to create config source: %v", err)
		return
	}

	ok := tc.executeBeforeAll(t, cmd)
	if !ok {
		return
	}

	client := authflowclient.NewClient(
		ctx,
		"localhost:4000",
		httputil.HTTPHost(fmt.Sprintf("%s.portal.localhost:4000", appID)),
	)

	var stepResults []StepResult
	var state string

	for i, step := range tc.Steps {
		if step.Name == "" {
			step.Name = fmt.Sprintf("step %d", i+1)
		}

		var result *StepResult
		result, state, ok = tc.executeStep(t, cmd, client, stepResults, state, step)
		if !ok {
			return
		}

		stepResults = append(stepResults, *result)
	}
}

// Execute before hooks to prepare fixtures
func (tc *TestCase) executeBeforeAll(t *testing.T, cmd *End2EndCmd) (ok bool) {
	for _, beforeHook := range tc.Before {
		switch beforeHook.Type {
		case BeforeHookTypeUserImport:
			err := cmd.ImportUsers(beforeHook.UserImport)
			if err != nil {
				t.Errorf("failed to import users: %v", err)
				return false
			}
		case BeforeHookTypeCustomSQL:
			err := cmd.ExecuteCustomSQL(beforeHook.CustomSQL.Path)
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

func (tc *TestCase) executeStep(
	t *testing.T,
	cmd *End2EndCmd,
	client *authflowclient.Client,
	prevSteps []StepResult,
	state string,
	step Step,
) (result *StepResult, nextState string, ok bool) {
	var flowResponse *authflowclient.FlowResponse
	var flowErr error

	switch step.Action {
	case StepActionCreate:
		var flowReference authflowclient.FlowReference
		err := json.Unmarshal([]byte(step.Input), &flowReference)
		if err != nil {
			t.Errorf("failed to parse input in '%s': %v\n", step.Name, err)
			return
		}

		flowResponse, flowErr = client.Create(flowReference, "")

		if step.Output != nil {
			ok := validateOutput(t, step, flowResponse, flowErr)
			if !ok {
				return nil, state, false
			}
		}

		nextState = state
		if flowResponse != nil {
			nextState = flowResponse.StateToken
		}

		result = &StepResult{
			Result: flowResponse,
			Error:  flowErr,
		}

	case StepActionGenerateTOTPCode:
		var lastStep *StepResult
		if len(prevSteps) != 0 {
			lastStep = &prevSteps[len(prevSteps)-1]
		}

		var parsedTOTPSecret string
		parsedTOTPSecret, ok = prepareTOTPSecret(t, cmd, lastStep, step.TOTPSecret)
		if !ok {
			return nil, state, false
		}

		totpCode, err := client.GenerateTOTPCode(parsedTOTPSecret)
		if err != nil {
			t.Errorf("failed to generate TOTP code in '%s': %v\n", step.Name, err)
			return
		}
		nextState = state

		result = &StepResult{
			Result: map[string]interface{}{
				"totp_code": totpCode,
			},
			Error: nil,
		}

	case StepActionOAuthRedirect:
		var lastStep *StepResult

		if len(prevSteps) != 0 {
			lastStep = &prevSteps[len(prevSteps)-1]
		}

		var parsedTo string
		parsedTo, ok = prepareTo(t, cmd, lastStep, step.To)
		if !ok {
			return nil, state, false
		}

		finalURL, err := client.OAuthRedirect(parsedTo, step.RedirectURI)
		if err != nil {
			t.Errorf("failed to follow OAuth redirect in '%s': %v\n", step.Name, err)
			return
		}

		finalURLParsed, err := url.Parse(finalURL)
		if err != nil {
			t.Errorf("failed to parse final URL in '%s': %v\n", step.Name, err)
			return
		}

		nextState = state

		result = &StepResult{
			Result: map[string]interface{}{
				"query": finalURLParsed.RawQuery,
			},
			Error: nil,
		}

	case StepActionInput:
		fallthrough
	case "":
		if len(prevSteps) == 0 {
			t.Errorf("no previous step result in '%s'", step.Name)
			return
		}

		lastStep := prevSteps[len(prevSteps)-1]
		input, ok := prepareInput(t, cmd, &lastStep, step.Input)
		if !ok {
			return nil, state, false
		}

		flowResponse, flowErr = client.Input(nil, nil, state, input)

		if step.Output != nil {
			ok := validateOutput(t, step, flowResponse, flowErr)
			if !ok {
				return nil, state, false
			}
		}

		nextState = state
		if flowResponse != nil {
			nextState = flowResponse.StateToken
		}

		result = &StepResult{
			Result: flowResponse,
			Error:  flowErr,
		}
	default:
		t.Errorf("unknown action in '%s': %s", step.Name, step.Action)
		return nil, state, false
	}

	return result, nextState, true
}

func prepareInput(t *testing.T, cmd *End2EndCmd, prev *StepResult, input string) (prepared map[string]interface{}, ok bool) {
	parsedInput, err := execTemplate(cmd, prev, input)
	if err != nil {
		t.Errorf("failed to parse input: %v\n", err)
		return nil, false
	}

	var inputMap map[string]interface{}
	err = json.Unmarshal([]byte(parsedInput), &inputMap)
	if err != nil {
		t.Errorf("failed to parse input: %v\n", err)
		return nil, false
	}

	return inputMap, true
}

func prepareTOTPSecret(t *testing.T, cmd *End2EndCmd, prev *StepResult, totpSecret string) (prepared string, ok bool) {
	parsedTOTPSecret, err := execTemplate(cmd, prev, totpSecret)
	if err != nil {
		t.Errorf("failed to parse totp_secret: %v\n", err)
		return "", false
	}

	return parsedTOTPSecret, true
}

func prepareTo(t *testing.T, cmd *End2EndCmd, prev *StepResult, to string) (prepared string, ok bool) {
	parsedTo, err := execTemplate(cmd, prev, to)
	if err != nil {
		t.Errorf("failed to parse to: %v\n", err)
		return "", false
	}

	return parsedTo, true
}

func execTemplate(cmd *End2EndCmd, prev *StepResult, content string) (string, error) {
	tmpl := texttemplate.New("")
	tmpl.Funcs(makeTemplateFuncMap(cmd))

	_, err := tmpl.Parse(content)
	if err != nil {
		return "", err
	}

	data := make(map[string]interface{})

	// Add prev result to data
	data["prev"], err = toMap(prev)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func makeTemplateFuncMap(cmd *End2EndCmd) texttemplate.FuncMap {
	templateFuncMap := sprig.HermeticHtmlFuncMap()
	templateFuncMap["linkOTPCode"] = func(claimName string, claimValue string) string {
		otpCode, err := cmd.GetLinkOTPCodeByClaim(claimName, claimValue)
		if err != nil {
			panic(err)
		}
		return otpCode
	}
	templateFuncMap["generateTOTPCode"] = func(secret string) string {
		totp, err := secretcode.NewTOTPFromSecret(secret)
		if err != nil {
			panic(err)
		}

		code, err := totp.GenerateCode(time.Now().UTC())
		if err != nil {
			panic(err)
		}
		return code
	}
	templateFuncMap["generateIDToken"] = func(userID string) string {
		idToken, err := cmd.GenerateIDToken(userID)
		if err != nil {
			panic(err)
		}
		return idToken
	}
	templateFuncMap["generateUUID"] = func() string {
		return uuid.Must(uuid.NewRandom()).String()
	}

	return templateFuncMap
}

func validateOutput(t *testing.T, step Step, flowResponse *authflowclient.FlowResponse, flowErr error) (ok bool) {
	flowResponseJson, _ := json.MarshalIndent(flowResponse, "", "  ")
	flowErrJson, _ := json.MarshalIndent(flowErr, "", "  ")

	errorViolations, resultViolations, err := MatchOutput(*step.Output, flowResponse, flowErr)
	if err != nil {
		t.Errorf("failed to match output in '%s': %v\n", step.Name, err)
		t.Errorf("  result: %s\n", flowResponseJson)
		t.Errorf("  error: %s\n", flowErrJson)
		return false
	}

	if len(errorViolations) > 0 {
		t.Errorf("error output mismatch in '%s':\n", step.Name)
		for _, violation := range errorViolations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  error: %s\n", flowErrJson)
		return false
	}

	if len(resultViolations) > 0 {
		t.Errorf("result output mismatch in '%s':\n", step.Name)
		for _, violation := range resultViolations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  result: %s\n", flowResponseJson)
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

func toMap(data interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var mapData map[string]interface{}
	err = json.Unmarshal(jsonData, &mapData)
	if err != nil {
		panic(err)
	}

	return mapData, nil
}
