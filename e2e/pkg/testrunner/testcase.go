package testrunner

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
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
	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
)

var _ = TestCaseSchema.Add("TestCase", `
{
	"type": "object",
	"properties": {
		"name": { "type": "string" },
		"focus": { "type": "boolean" },
		"authgear.yaml": { "$ref": "#/$defs/AuthgearYAMLSource" },
		"extra_files_directory": { "type": "string" },
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
	Focus               bool               `json:"focus"`
	AuthgearYAMLSource  AuthgearYAMLSource `json:"authgear.yaml"`
	ExtraFilesDirectory string             `json:"extra_files_directory"`
	Steps               []Step             `json:"steps"`
	Before              []BeforeHook       `json:"before"`
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
		t.Fatalf("%v", err.Error())
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
		case BeforeHookTypeCreateSession:
			err = cmd.ExecuteCreateSession(beforeHook.CreateSession)
			if err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}
		case BeforeHookTypeCreateChallenge:
			err = cmd.ExecuteCreateChallenge(beforeHook.CreateChallenge)
			if err != nil {
				return fmt.Errorf("failed to create challenge: %w", err)
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

	nextState = state

	switch step.Action {
	case StepActionSleep:
		d, err := time.ParseDuration(step.SleepFor)
		if err != nil {
			panic(err)
		}

		time.Sleep(d)
		result = &StepResult{
			Result: nil,
			Error:  nil,
		}
	case StepActionCreate:
		input, ok := prepareInput(t, cmd, prevSteps, step.Input)
		if !ok {
			return nil, state, false
		}

		flowResponse, flowErr = client.CreateFlow(input)

		if step.Output != nil {
			ok := validateAuthflowOutput(t, step, flowResponse, flowErr)
			if !ok {
				return nil, state, false
			}
		}

		if flowResponse != nil {
			nextState = flowResponse.StateToken
		}

		result = &StepResult{
			Result: flowResponse,
			Error:  flowErr,
		}

	case StepActionGenerateTOTPCode:
		var parsedTOTPSecret string
		parsedTOTPSecret, ok = renderTemplateString(t, cmd, prevSteps, step.TOTPSecret)
		if !ok {
			return nil, state, false
		}

		totpCode, err := client.GenerateTOTPCode(parsedTOTPSecret)
		if err != nil {
			t.Errorf("failed to generate TOTP code in '%s': %v\n", step.Name, err)
			return
		}

		result = &StepResult{
			Result: map[string]interface{}{
				"totp_code": totpCode,
			},
			Error: nil,
		}

	case StepActionOAuthRedirect:
		var parsedTo string
		parsedTo, ok = renderTemplateString(t, cmd, prevSteps, step.To)
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

	case StepActionHTTPRequest:
		var outputOk bool = true
		var httpResult interface{} = nil
		requesturl, ok := renderTemplateString(t, cmd, prevSteps, step.HTTPRequestURL)
		if !ok {
			return nil, state, false
		}
		if step.HTTPRequestSessionCookie != nil {
			client.InjectSession(
				step.HTTPRequestSessionCookie.IDPSessionID,
				step.HTTPRequestSessionCookie.IDPSessionToken,
			)
		}
		var headers = step.HTTPRequestHeaders
		if headers == nil {
			headers = map[string]string{}
		}
		body := ""
		if step.HTTPRequestBody != "" {
			body = step.HTTPRequestBody
		} else if step.HTTPRequestFormURLEncodedBody != nil {
			headers["Content-Type"] = "application/x-www-form-urlencoded"
			values := url.Values{}
			for k, v := range step.HTTPRequestFormURLEncodedBody {
				values.Add(k, v)
			}
			body = values.Encode()
		}
		err := client.MakeHTTPRequest(
			step.HTTPRequestMethod,
			requesturl,
			headers,
			body,
			func(r *http.Response) error {
				if r != nil {
					httpResult = NewResultHTTPResponse(r)
				}
				if step.HTTPOutput != nil {
					outputOk = validateHTTPOutput(t, step, step.HTTPOutput, r)
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

		if step.SAMLRequestSessionCookie != nil {
			client.InjectSession(
				step.SAMLRequestSessionCookie.IDPSessionID,
				step.SAMLRequestSessionCookie.IDPSessionToken,
			)
		}

		var relayState string
		if step.SAMLRequestRelayState != "" {
			rs, ok := renderTemplateString(t, cmd, prevSteps, step.SAMLRequestRelayState)
			if !ok {
				return nil, state, false
			}
			relayState = rs
		}

		var samlOutputOk bool = true
		var httpResult interface{} = nil
		err := client.SendSAMLRequest(
			step.SAMLRequestDestination,
			step.SAMLElementName,
			step.SAMLElement,
			step.SAMLRequestBinding,
			relayState,
			func(r *http.Response) error {
				if r != nil {
					httpResult = NewResultHTTPResponse(r)
				}
				if step.SAMLOutput != nil {
					samlOutputOk = validateSAMLOutput(t, step.SAMLOutput, r)
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
	case StepActionOAuthSetup:
		output, err := client.SetupOAuth()
		if err != nil {
			t.Errorf("failed to setup oauth: %v", err)
			return nil, state, false
		}

		result = &StepResult{
			Result: output,
			Error:  nil,
		}
	case StepActionOAuthExchangeCode:
		var codeVerifier string
		codeVerifier, ok = renderTemplateString(t, cmd, prevSteps, step.OAuthExchangeCodeCodeVerifier)
		if !ok {
			return nil, state, false
		}

		var redirectURI string
		redirectURI, ok = renderTemplateString(t, cmd, prevSteps, step.OAuthExchangeCodeRedirectURI)

		output, err := client.OAuthExchangeCode(authflowclient.OAuthExchangeCodeOptions{
			CodeVerifier: codeVerifier,
			RedirectURI:  redirectURI,
		})
		if err != nil {
			t.Errorf("failed to exchange code: %v\n", err)
			return
		}

		if step.Output != nil {
			ok := validateOAuthExchangeCodeOutput(t, step, output)
			if !ok {
				return nil, state, false
			}
		}

		result = &StepResult{
			Result: output,
			Error:  nil,
		}

	case StepActionInput:
		fallthrough
	case "":
		input, ok := prepareInput(t, cmd, prevSteps, step.Input)
		if !ok {
			return nil, state, false
		}

		flowResponse, flowErr = client.InputFlow(nil, nil, state, input)

		if step.Output != nil {
			ok := validateAuthflowOutput(t, step, flowResponse, flowErr)
			if !ok {
				return nil, state, false
			}
		}

		if flowResponse != nil {
			nextState = flowResponse.StateToken
		}

		result = &StepResult{
			Result: flowResponse,
			Error:  flowErr,
		}
	case StepActionAdminAPIQuery:
		if step.AdminAPIRequest == nil {
			t.Errorf("admin_api_request must be provided for admin_api_graphql step")
			return nil, state, false
		}

		var variables map[string]interface{} = map[string]interface{}{}
		if step.AdminAPIRequest.Variables != "" {
			renderedVariables, ok := renderTemplateString(t, cmd, prevSteps, step.AdminAPIRequest.Variables)
			if !ok {
				t.Errorf("failed to render admin_api_request.variables")
			}
			err := json.Unmarshal([]byte(renderedVariables), &variables)
			if err != nil {
				t.Errorf("failed to unmarshal admin_api_request.variables: %v", err)
				return nil, state, false
			}
		}

		renderedQuery, ok := renderTemplateString(t, cmd, prevSteps, step.AdminAPIRequest.Query)
		if !ok {
			t.Errorf("failed to render admin_api_request.query")
		}
		resp, err := cmd.Client.GraphQLAPI(cmd.AppID, authflowclient.GraphQLAPIRequest{
			Query:     renderedQuery,
			Variables: variables,
		})
		if err != nil {
			t.Errorf("failed to make adminapi request: %v", err)
			return nil, state, false
		}

		if step.AdminAPIOutput != nil {
			renderedResult, ok := renderTemplateString(t, cmd, prevSteps, step.AdminAPIOutput.Result)
			if !ok {
				t.Errorf("failed to render adminapi_output.result")
			}
			expectedOutput := *step.AdminAPIOutput
			expectedOutput.Result = renderedResult
			ok = validateAdminAPIOutput(t, &expectedOutput, resp)
			if !ok {
				return nil, state, false
			}
		}

		result = &StepResult{
			Result: resp,
			Error:  err,
		}
	case StepActionAdminAPIUserImportCreate:
		if step.AdminAPIUserImportRequest == nil {
			t.Errorf("admin_api_user_import_request must be provided for admin_api_user_import_create step")
			return nil, state, false
		}

		renderedJSONDocument, ok := renderTemplateString(t, cmd, prevSteps, step.AdminAPIUserImportRequest.JSONDocument)
		if !ok {
			t.Errorf("failed to render admin_api_user_import_request.json_document")
		}

		resp, err := cmd.Client.CreateUserImport(cmd.AppID, authflowclient.UserImportRequest{
			JSONDocument: renderedJSONDocument,
		})

		if step.AdminAPIUserImportOutput != nil {
			ok := validateUserImportOutput(t, &step, resp, err)
			if !ok {
				return nil, state, false
			}
		}

		result = &StepResult{
			Result: resp,
			Error:  err,
		}
	case StepActionAdminAPIUserImportGet:
		renderedID, ok := renderTemplateString(t, cmd, prevSteps, step.AdminAPIUserImportID)
		if !ok {
			t.Errorf("failed to render admin_api_user_import_id")
		}

		resp, err := cmd.Client.GetUserImport(cmd.AppID, renderedID)
		ok = validateUserImportOutput(t, &step, resp, err)
		if !ok {
			return nil, state, false
		}

		result = &StepResult{
			Result: resp,
			Error:  err,
		}
	default:
		t.Errorf("unknown action in '%s': %s", step.Name, step.Action)
		return nil, state, false
	}

	if result != nil {
		result.Step = &step
	}

	return result, nextState, true
}

func prepareInput(t *testing.T, cmd *End2EndCmd, prevSteps []StepResult, input string) (prepared map[string]interface{}, ok bool) {
	renderedString, ok := renderTemplateString(t, cmd, prevSteps, input)
	if !ok {
		return nil, false
	}

	var inputMap map[string]interface{}
	err := json.Unmarshal([]byte(renderedString), &inputMap)
	if err != nil {
		t.Errorf("failed to parse input: %v\n", err)
		return nil, false
	}

	return inputMap, true
}

func renderTemplateString(t *testing.T, cmd *End2EndCmd, prevSteps []StepResult, templateString string) (string, bool) {
	renderedString, err := execTemplate(cmd, prevSteps, templateString)
	if err != nil {
		t.Errorf("failed to render template string: %v", err)
		return "", false
	}

	return renderedString, true
}

func execTemplate(cmd *End2EndCmd, prevSteps []StepResult, content string) (string, error) {
	tmpl := texttemplate.New("")
	tmpl.Funcs(makeTemplateFuncMap(cmd))

	_, err := tmpl.Parse(content)
	if err != nil {
		return "", err
	}

	data := make(map[string]any)
	data["AppID"] = cmd.AppID

	// Add prev result to data
	if len(prevSteps) > 0 {
		lastStep := prevSteps[len(prevSteps)-1]
		data["prev"], err = toMap(lastStep)
		if err != nil {
			return "", err
		}
	}

	// Add named steps to data
	steps := make(map[string]any)
	for _, step := range prevSteps {
		if step.Step.Name != "" {
			_, ok := steps[step.Step.Name]
			if ok {
				return "", fmt.Errorf("step name duplicated: %v", step.Step.Name)
			}
			steps[step.Step.Name], err = toMap(step)
			if err != nil {
				return "", err
			}
		}
	}
	data["steps"] = steps

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
	templateFuncMap["nodeID"] = func(nodeType string, uuid string) string {
		return relay.ToGlobalID(nodeType, uuid)
	}

	return templateFuncMap
}

func validateAuthflowOutput(t *testing.T, step Step, flowResponse *authflowclient.FlowResponse, flowErr error) (ok bool) {
	flowResponseJson, _ := json.MarshalIndent(flowResponse, "", "  ")
	flowErrJson, _ := json.MarshalIndent(flowErr, "", "  ")

	errorViolations, resultViolations, err := MatchAuthflowOutput(*step.Output, flowResponse, flowErr)
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

func validateSAMLElement(t *testing.T, expected *OuputSAMLElement, httpResponse *http.Response) (ok bool) {
	ok = true

	var responseDoc *etree.Document
	switch expected.Binding {
	case authflowclient.SAMLBindingHTTPPost:
		// For post binding, read the SAMLResponse from the response body
		body, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			t.Errorf("failed to read response body: %v", err)
			ok = false
			return
		}
		doc := etree.NewDocument()
		err = doc.ReadFromString(string(body))
		if err != nil {
			ok = false
			t.Errorf("failed to parse response body as html")
			return
		}
		samlResponseEl := doc.FindElement(fmt.Sprintf("./html/body/form/input[@name='%s']", expected.ElementName))
		if samlResponseEl == nil {
			ok = false
			t.Errorf("no %s input found in html", expected.ElementName)
			return
		}
		samlResponseAttr := samlResponseEl.SelectAttr("value")
		if samlResponseAttr == nil {
			ok = false
			t.Errorf("%s input has no value", expected.ElementName)
			return
		}
		decodedXML, err := base64.StdEncoding.DecodeString(samlResponseAttr.Value)
		if err != nil {
			ok = false
			t.Errorf("decode SAML element failed: %v", err)
			return
		}
		responseDoc = etree.NewDocument()
		err = responseDoc.ReadFromString(string(decodedXML))
		if err != nil {
			ok = false
			t.Errorf("failed to parse SAML element as xml")
			return
		}

	case authflowclient.SAMLBindingHTTPRedirect:
		redirectLocation, err := url.Parse(httpResponse.Header.Get("location"))
		if err != nil {
			ok = false
			t.Errorf("invalid redirect location")
			return
		}
		encodedSAMLElement := redirectLocation.Query().Get(expected.ElementName)
		compressedSAMLElement, err := base64.StdEncoding.DecodeString(string(encodedSAMLElement))
		if err != nil {
			ok = false
			t.Errorf("decode SAML element failed: %v", err)
			return
		}
		flateReader := flate.NewReader(bytes.NewBuffer([]byte(compressedSAMLElement)))
		elXML, err := io.ReadAll(flateReader)
		if err != nil {
			ok = false
			t.Errorf("failed to decompress SAML element %v", err)
			return
		}
		responseDoc = etree.NewDocument()
		err = responseDoc.ReadFromString(string(elXML))
		if err != nil {
			ok = false
			t.Errorf("failed to parse SAML element as xml")
			return
		}
	default:
		t.Errorf("not implemented")
		ok = false
		return
	}

	expectedDoc := etree.NewDocument()
	err := expectedDoc.ReadFromString(string(expected.Match))
	if err != nil {
		ok = false
		t.Errorf("failed to parse match as xml")
		return
	}

	var assertElements func(parentPath string, expectedEls []*etree.Element)

	assertElements = func(parentPath string, expectedEls []*etree.Element) {
		for idx, el := range expectedEls {
			expectedEl := el
			// Do not use GetPath to get the path because it does not handle duplicated elements of the same Tag
			path := parentPath + fmt.Sprintf("/*[%d]", idx+1)
			// Check existence
			actualEl := responseDoc.FindElement(path)
			if actualEl == nil {
				ok = false
				t.Errorf("element not found in path: %v", path)
				continue
			}

			// Check Tag
			if expectedEl.Tag != "any" && expectedEl.Tag != actualEl.Tag {
				ok = false
				t.Errorf("element %v has unmatched tag: expected: %v, actual: %v",
					path,
					expectedEl.Tag,
					actualEl.Tag,
				)
				continue
			}

			// Check Space
			if expectedEl.Space != "" && expectedEl.Space != actualEl.Space {
				ok = false
				t.Errorf("element %v has unmatched space: expected: %v, actual: %v",
					path,
					expectedEl.Space,
					actualEl.Space,
				)
				continue
			}

			// Check attributes
			for _, expectedAttr := range expectedEl.Attr {
				actualValue := actualEl.SelectAttrValue(expectedAttr.Key, "")
				if actualValue != expectedAttr.Value {
					ok = false
					t.Errorf("element %v has unmatched attribute. key: %v, expected: %v, actual: %v",
						path,
						expectedAttr.Key,
						expectedAttr.Value,
						actualValue,
					)
				}
			}
			// Check text
			expectedText := strings.Trim(expectedEl.Text(), "\n ")
			actualText := strings.Trim(actualEl.Text(), "\n ")
			if expectedText != "" && expectedText != actualText {
				ok = false
				t.Errorf("element %v has unmatched text. expected: %v, actual: %v",
					path,
					expectedEl.Text(),
					actualEl.Text(),
				)
			}
			// Check children
			if len(expectedEl.ChildElements()) > 0 {
				assertElements(path, expectedEl.ChildElements())
			}
		}
	}

	assertElements("", expectedDoc.ChildElements())
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

func validateHTTPOutput(t *testing.T, step Step, httpOutput *HTTPOutput, response *http.Response) (ok bool) {
	ok = true
	if response == nil {
		t.Errorf("expected http response but got nil")
		ok = false
		return
	}
	if httpOutput.HTTPStatus != nil {
		if !validateHTTPResponseStatus(t, int(*httpOutput.HTTPStatus), response) {
			ok = false
		}
	}
	if httpOutput.RedirectPath != nil {
		if !validateRedirectLocation(t,
			*httpOutput.RedirectPath,
			response,
		) {
			ok = false
		}
	}
	if httpOutput.SAMLElement != nil {
		statusOk := validateSAMLElement(t,
			httpOutput.SAMLElement,
			response,
		)
		if !statusOk {
			ok = false
		}
	}
	if httpOutput.JSONBody != nil {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			t.Errorf("failed to read response body: %v", err)
			ok = false
			return
		}
		var bodyIntf interface{}
		err = json.Unmarshal(body, &bodyIntf)
		if err != nil {
			t.Errorf("failed to parse response body as json: %v", err)
			t.Errorf("  result: %s\n", string(body))
			ok = false
			return
		}
		bodyJson, _ := json.MarshalIndent(bodyIntf, "", "  ")
		violations, err := MatchJSON(string(body), *httpOutput.JSONBody)
		if err != nil {
			t.Errorf("failed to match output in '%s': %v\n", step.Name, err)
			t.Errorf("  result: %s\n", bodyJson)
			ok = false
		}
		if len(violations) > 0 {
			t.Errorf("result output mismatch in '%s':\n", step.Name)
			for _, violation := range violations {
				t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
			}
			t.Errorf("  result: %s\n", bodyJson)
			ok = false
		}
	}
	return ok
}

func validateSAMLOutput(t *testing.T, samlOutput *SAMLOutput, response *http.Response) (ok bool) {
	ok = true
	if response == nil {
		t.Errorf("expected http response but got nil")
		ok = false
		return
	}
	if samlOutput.HTTPStatus != nil {
		if !validateHTTPResponseStatus(t, int(*samlOutput.HTTPStatus), response) {
			ok = false
		}
	}
	if samlOutput.RedirectPath != nil {
		if !validateRedirectLocation(t,
			*samlOutput.RedirectPath,
			response,
		) {
			ok = false
		}
	}
	if samlOutput.SAMLElement != nil {
		statusOk := validateSAMLElement(t,
			samlOutput.SAMLElement,
			response,
		)
		if !statusOk {
			ok = false
		}
	}
	return ok
}

func validateOAuthExchangeCodeOutput(t *testing.T, step Step, output *authflowclient.OAuthExchangeCodeResult) (ok bool) {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")

	violations, err := MatchJSON(string(outputJSON), step.Output.Result)
	if err != nil {
		t.Errorf("failed to match output in '%s': %v\n", step.Name, err)
		t.Errorf("  result: %v\n", string(outputJSON))
		return false
	}

	if len(violations) > 0 {
		t.Errorf("result output mismatch in '%v':\n", step.Name)
		for _, violation := range violations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  result: %v\n", string(outputJSON))
		return false
	}

	return true
}

func validateAdminAPIOutput(t *testing.T, expected *AdminAPIOutput, resp *authflowclient.GraphQLResponse) (ok bool) {
	respJSON, _ := json.MarshalIndent(resp, "", "  ")

	violations, err := MatchAdminAPIOutput(*expected, resp)
	if err != nil {
		t.Errorf("failed to match admin_api_output: %v\n", err)
		t.Errorf("  result: %s\n", respJSON)
		return false
	}

	if len(violations) > 0 {
		t.Errorf("admin_api_output mismatch:\n")
		for _, violation := range violations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  result: %s\n", respJSON)
		return false
	}

	return true
}

func validateUserImportOutput(t *testing.T, step *Step, userImportResult *authflowclient.UserImportResponseResult, userImportError error) (ok bool) {
	userImportResultJSON, _ := json.MarshalIndent(userImportResult, "", "  ")
	userImportErrorJSON, _ := json.MarshalIndent(userImportError, "", "  ")

	errorViolations, resultViolations, err := MatchUserImportOutput(*step.AdminAPIUserImportOutput, userImportResult, userImportError)
	if err != nil {
		t.Errorf("failed to match output in '%s': %v\n", step.Name, err)
		t.Errorf("  result: %s\n", userImportResultJSON)
		t.Errorf("  error: %s\n", userImportErrorJSON)
		return false
	}

	if len(errorViolations) > 0 {
		t.Errorf("error output mismatch in '%s':\n", step.Name)
		for _, violation := range errorViolations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  error: %s\n", userImportErrorJSON)
		return false
	}

	if len(resultViolations) > 0 {
		t.Errorf("result output mismatch in '%s':\n", step.Name)
		for _, violation := range resultViolations {
			t.Errorf("  | %s: %s. Expected %s, got %s", violation.Path, violation.Message, violation.Expected, violation.Actual)
		}
		t.Errorf("  result: %s\n", userImportResultJSON)
		return false
	}

	return true
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
