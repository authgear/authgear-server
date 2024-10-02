package testrunner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	texttemplate "text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/beevik/etree"

	authflowclient "github.com/authgear/authgear-server/e2e/pkg/e2eclient"
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
	// Create project per test case
	cmd, err := NewEnd2EndCmd(NewEnd2EndCmdOptions{
		TestCase: tc,
		Test:     t,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = tc.executeBeforeAll(cmd)
	if err != nil {
		t.Errorf("failed to execute before hooks: %v", err)
		return
	}

	var stepResults []StepResult
	var state string
	var ok bool

	for i, step := range tc.Steps {
		if step.Name == "" {
			step.Name = fmt.Sprintf("step %d", i+1)
		}

		var result *StepResult
		result, state, ok = tc.executeStep(t, cmd, cmd.Client, stepResults, state, step)
		if !ok {
			return
		}

		stepResults = append(stepResults, *result)
	}
}

// Execute before hooks to prepare fixtures
func (tc *TestCase) executeBeforeAll(cmd *End2EndCmd) (err error) {
	for _, beforeHook := range tc.Before {
		switch beforeHook.Type {
		case BeforeHookTypeUserImport:
			err = cmd.ImportUsers(beforeHook.UserImport)
			if err != nil {
				return fmt.Errorf("failed to import users: %w", err)
			}
		case BeforeHookTypeCustomSQL:
			err = cmd.ExecuteSQLInsertUpdateFile(beforeHook.CustomSQL.Path)
			if err != nil {
				return fmt.Errorf("failed to execute custom SQL: %w", err)
			}
		default:
			errStr := fmt.Sprintf("unknown before hook type: %s", beforeHook.Type)
			return errors.New(errStr)
		}
	}

	return nil
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

		flowResponse, flowErr = client.CreateFlow(flowReference, "")

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

	case StepActionQuery:
		jsonArrString, err := cmd.QuerySQLSelectRaw(step.Query)
		if err != nil {
			t.Errorf("failed to execute SQL Select query: %v", err)
			return
		}

		rowsResult := map[string]interface{}{}
		var rows []interface{}
		err = json.Unmarshal([]byte(jsonArrString), &rows)
		if err != nil {
			t.Errorf("failed to unmarshal json rows: %v", err)
			return
		}
		rowsResult["rows"] = rows
		result = &StepResult{
			Result: rowsResult,
			Error:  nil,
		}

		if step.QueryOutput != nil {
			ok := validateQueryResult(t, step, rows)
			if !ok {
				return nil, state, false
			}
		}

		nextState = state
	case StepActionHTTPRequest:
		var lastStep *StepResult
		if len(prevSteps) != 0 {
			lastStep = &prevSteps[len(prevSteps)-1]
		}

		var outputOk bool = true
		var httpResult interface{} = nil
		url, ok := prepareHTTPRequestURL(t, cmd, lastStep, step.HTTPRequestURL)
		if !ok {
			return nil, state, false
		}
		err := client.MakeHTTPRequest(step.HTTPRequestMethod, url, func(r *http.Response) error {
			if r != nil {
				httpResult = NewResultHTTPResponse(r)
			}
			if step.SAMLOutput != nil {
				outputOk = validateHTTPResponse(t, step, r)
			}
			return nil
		})
		if err != nil {
			t.Errorf("failed to send http request: %v", err)
			return nil, state, false
		}
		if !outputOk {
			return nil, state, false
		}
		result = &StepResult{
			Result: httpResult,
			Error:  nil,
		}

	case StepActionSAMLRequest:
		var samlOutputOk bool = true
		var httpResult interface{} = nil
		err := client.SendSAMLRequest(
			step.SAMLRequestDestination,
			step.SAMLRequest,
			step.SAMLRequestBinding, func(r *http.Response) error {
				if r != nil {
					httpResult = NewResultHTTPResponse(r)
				}
				if step.SAMLOutput != nil {
					samlOutputOk = validateSAMLResponse(t, step, r)
				}
				return nil
			})
		if err != nil {
			t.Errorf("failed to send saml request: %v", err)
			return nil, state, false
		}
		if !samlOutputOk {
			return nil, state, false
		}
		result = &StepResult{
			Result: httpResult,
			Error:  nil,
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

		flowResponse, flowErr = client.InputFlow(nil, nil, state, input)

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

func prepareHTTPRequestURL(t *testing.T, cmd *End2EndCmd, prev *StepResult, url string) (prepared string, ok bool) {
	url, err := execTemplate(cmd, prev, url)
	if err != nil {
		t.Errorf("failed to parse http_request_url: %v\n", err)
		return "", false
	}

	return url, true
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

func validateQueryResult(t *testing.T, step Step, rows []interface{}) (ok bool) {
	rowsJSON, _ := json.MarshalIndent(rows, "", "  ")
	resultViolations, err := MatchOutputQueryResult(*step.QueryOutput, rows)
	if err != nil {
		t.Errorf("failed to match output in '%s': %v\n", step.Name, err)
		t.Errorf("  result: %s\n", rowsJSON)
		return false
	}

	if len(resultViolations) > 0 {
		t.Errorf("result output mismatch in '%s':\n", step.Name)
		for _, violation := range resultViolations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  result: %s\n", rowsJSON)
		return false
	}

	return true

}

func validateRedirectLocation(t *testing.T, expectedPath string, response *http.Response) (ok bool) {
	ok = true
	actualLocationURL, err := url.Parse(response.Header.Get("Location"))
	if err != nil {
		ok = false
		t.Errorf("Location header is not a valid url")
	}
	if !ok {
		return ok
	}
	// We only compare the url without query parameters
	if expectedPath != actualLocationURL.EscapedPath() {
		ok = false
		t.Errorf("redirect path unmatch. expected: %s, actual: %s",
			expectedPath,
			actualLocationURL.EscapedPath(),
		)
	}
	return ok
}

func validateSAMLStatus(t *testing.T, expectedStatus string, responseBody []byte) (ok bool) {
	ok = true
	doc := etree.NewDocument()
	err := doc.ReadFromString(string(responseBody))
	if err != nil {
		ok = false
		t.Errorf("failed to read parse body as xml")
		return
	}
	statusCodeEl := doc.FindElement("./Status/StatusCode")
	if statusCodeEl == nil {
		ok = false
		t.Errorf("no StatusCode element found")
		return
	}
	statusCodeValue := statusCodeEl.SelectAttr("Value")
	if statusCodeValue == nil {
		ok = false
		t.Errorf("no Value in StatusCode")
		return
	}
	if statusCodeValue.Value != expectedStatus {
		ok = false
		t.Errorf("unexpected SAML status. expected: %s, actual: %s",
			expectedStatus,
			statusCodeValue.Value,
		)
		return
	}
	return ok
}

func validateHTTPResponseStatus(t *testing.T, expectedStatus int, response *http.Response) (ok bool) {
	if response.StatusCode != expectedStatus {
		t.Errorf("http response status code unmatch. expected: %d, actual: %d",
			expectedStatus,
			response.StatusCode,
		)
		return false
	}
	return true
}

func validateHTTPResponse(t *testing.T, step Step, response *http.Response) (ok bool) {
	ok = true
	if response == nil {
		t.Errorf("expected http response but got nil")
		ok = false
		return
	}
	if step.HTTPOutput.HTTPStatus != nil {
		if !validateHTTPResponseStatus(t, int(*step.HTTPOutput.HTTPStatus), response) {
			ok = false
		}
	}
	if step.HTTPOutput.RedirectPath != nil {
		if !validateRedirectLocation(t,
			*step.HTTPOutput.RedirectPath,
			response,
		) {
			ok = false
		}
	}
	return ok
}

func validateSAMLResponse(t *testing.T, step Step, response *http.Response) (ok bool) {
	ok = true
	if response == nil {
		t.Errorf("expected http response but got nil")
		ok = false
		return
	}
	if step.SAMLOutput.HTTPStatus != nil {
		if !validateHTTPResponseStatus(t, int(*step.SAMLOutput.HTTPStatus), response) {
			ok = false
		}
	}
	if step.SAMLOutput.RedirectPath != nil {
		if !validateRedirectLocation(t,
			*step.SAMLOutput.RedirectPath,
			response,
		) {
			ok = false
		}
	}
	if step.SAMLOutput.SAMLStatus != nil {
		responseData, err := io.ReadAll(response.Body)
		if err != nil {
			ok = false
			t.Errorf("failed to read response body")
		}
		statusOk := validateSAMLStatus(t,
			*step.SAMLOutput.SAMLStatus,
			responseData,
		)
		if !statusOk {
			ok = false
		}
	}
	return ok
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
