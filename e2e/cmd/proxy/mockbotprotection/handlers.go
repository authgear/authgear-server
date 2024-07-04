package mockbotprotection

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

const (
	InvalidRequest      = "invalid_request"
	InternalServerError = "internal_server_error"
	VerifyEndpoint      = "/verify"

	applicationJSON = "application/json"
)

func (m *MockBotProtection) Verify(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	switch m.Provider.Type {
	case config.BotProtectionProviderTypeCloudflare:
		verifyCloudflare(rw, req)
	case config.BotProtectionProviderTypeRecaptchaV2:
		verifyRecaptchav2(rw, req)
	}
}

func verifyCloudflare(rw http.ResponseWriter, req *http.Request) {
	requiredFields := []string{"response", "secret"}
	valid := assertPresence(requiredFields, rw, req)
	if !valid {
		return
	}

	response := req.Form.Get("response")

	switch response {
	case "pass":
		jsonResponse(rw, []byte(`{"success": true}`))
	case "service_unavailable":
		jsonResponse(rw, []byte(`{"success": false, "error-codes": ["internal-error"]}`))
	case "failed":
		fallthrough
	default:
		jsonResponse(rw, []byte(`{"success": false, "error-codes": ["invalid-input-response"]}`))
	}
}

func verifyRecaptchav2(rw http.ResponseWriter, req *http.Request) {
	requiredFields := []string{"response", "secret"}
	valid := assertPresence(requiredFields, rw, req)
	if !valid {
		return
	}

	response := req.Form.Get("response")

	switch response {
	case "pass":
		jsonResponse(rw, []byte(`{"success": true}`))
	// Note that recaptcha v2 does not support service_unavailable
	case "failed":
		fallthrough
	default:
		jsonResponse(rw, []byte(`{"success": false, "error-codes": ["invalid-input-response"]}`))
	}
}

func assertPresence(params []string, rw http.ResponseWriter, req *http.Request) bool {
	for _, param := range params {
		if req.Form.Get(param) != "" {
			continue
		}
		errorResponse(
			rw,
			InvalidRequest,
			fmt.Sprintf("The request is missing the required parameter: %s", param),
			http.StatusBadRequest,
		)
		return false
	}
	return true
}

func errorResponse(rw http.ResponseWriter, error, description string, statusCode int) {
	errJSON := map[string]string{
		"error":             error,
		"error_description": description,
	}
	resp, err := json.Marshal(errJSON)
	if err != nil {
		http.Error(rw, error, http.StatusInternalServerError)
	}

	noCache(rw)
	rw.Header().Set("Content-Type", applicationJSON)
	rw.WriteHeader(statusCode)

	_, err = rw.Write(resp)
	if err != nil {
		panic(err)
	}
}

func internalServerError(rw http.ResponseWriter, errorMsg string) {
	errorResponse(rw, InternalServerError, errorMsg, http.StatusInternalServerError)
}

func jsonResponse(rw http.ResponseWriter, data []byte) {
	noCache(rw)
	rw.Header().Set("Content-Type", applicationJSON)
	rw.WriteHeader(http.StatusOK)

	_, err := rw.Write(data)
	if err != nil {
		panic(err)
	}
}

func noCache(rw http.ResponseWriter) {
	rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	rw.Header().Set("Pragma", "no-cache")
}
