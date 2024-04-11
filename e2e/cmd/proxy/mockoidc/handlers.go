package mockoidc

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	IssuerBase            = ""
	AuthorizationEndpoint = "/authorize"
	TokenEndpoint         = "/token"
	UserinfoEndpoint      = "/userinfo"
	JWKSEndpoint          = "/.well-known/jwks.json"
	DiscoveryEndpoint     = "/.well-known/openid-configuration"

	InvalidRequest       = "invalid_request"
	InvalidClient        = "invalid_client"
	InvalidGrant         = "invalid_grant"
	UnsupportedGrantType = "unsupported_grant_type"
	InvalidScope         = "invalid_scope"
	InternalServerError  = "internal_server_error"

	applicationJSON = "application/json"
	openidScope     = "openid"
)

var (
	GrantTypesSupported = []string{
		"authorization_code",
	}
	ResponseTypesSupported = []string{
		"code",
	}
	SubjectTypesSupported = []string{
		"public",
	}
	IDTokenSigningAlgValuesSupported = []string{
		"RS256",
	}
	TokenEndpointAuthMethodsSupported = []string{
		"client_secret_basic",
		"client_secret_post",
	}
	ClaimsSupported = []string{
		"sub",
		"email",
		"email_verified",
		"preferred_username",
		"phone_number",
		"address",
		"groups",
		"iss",
		"aud",
		"upn",
	}
)

func (m *MockOIDC) Authorize(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	valid := assertPresence(
		[]string{"scope", "client_id" /*, "state"*/, "response_type", "redirect_uri"}, rw, req)
	if !valid {
		return
	}

	if !validateScope(m.Provider.ScopesSupported, rw, req) {
		return
	}

	validClient := assertEqual("client_id", m.ClientID,
		InvalidClient, "Invalid client id", rw, req)
	if !validClient {
		return
	}

	validType := assertEqual("response_type", "code",
		UnsupportedGrantType, "Invalid response type", rw, req)
	if !validType {
		return
	}

	session, err := m.SessionStore.NewSession(
		m.ClientID,
		req.Form.Get("scope"),
		DefaultUser(),
	)
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	redirectURI, err := url.Parse(req.Form.Get("redirect_uri"))
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}
	params, _ := url.ParseQuery(redirectURI.RawQuery)
	params.Set("code", session.SessionID)
	params.Set("state", req.Form.Get("state"))
	redirectURI.RawQuery = params.Encode()

	http.Redirect(rw, req, redirectURI.String(), http.StatusFound)
}

type tokenResponse struct {
	AccessToken string        `json:"access_token,omitempty"`
	IDToken     string        `json:"id_token,omitempty"`
	TokenType   string        `json:"token_type"`
	ExpiresIn   time.Duration `json:"expires_in"`
}

func (m *MockOIDC) Token(rw http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	if !m.validateTokenParams(rw, req) {
		return
	}

	var (
		session *Session
		valid   bool
	)
	grantType := req.Form.Get("grant_type")
	switch grantType {
	case "authorization_code":
		if session, valid = m.validateCodeGrant(rw, req); !valid {
			return
		}
	default:
		errorResponse(rw, InvalidRequest,
			fmt.Sprintf("Invalid grant type: %s", grantType), http.StatusBadRequest)
		return
	}

	tr := &tokenResponse{
		TokenType: "bearer",
		ExpiresIn: m.AccessTTL,
	}
	err = m.setTokens(tr, session, grantType)
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	resp, err := json.Marshal(tr)
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}
	noCache(rw)
	jsonResponse(rw, resp)
}

func (m *MockOIDC) validateTokenParams(rw http.ResponseWriter, req *http.Request) bool {
	if !assertPresence([]string{"client_id", "client_secret"}, rw, req) {
		return false
	}

	validClient := assertEqual("client_id", m.ClientID,
		InvalidClient, "Invalid client id", rw, req)
	if !validClient {
		return false
	}
	validClientSecret := assertEqual("client_secret", m.ClientSecret,
		InvalidClient, "Invalid client secret", rw, req)
	if !validClientSecret {
		return false
	}

	return true
}

func (m *MockOIDC) validateCodeGrant(rw http.ResponseWriter, req *http.Request) (*Session, bool) {
	if !assertPresence([]string{"code"}, rw, req) {
		return nil, false
	}
	equal := assertEqual("grant_type", "authorization_code",
		UnsupportedGrantType, "Invalid grant type", rw, req)
	if !equal {
		return nil, false
	}

	code := req.Form.Get("code")
	session, err := m.SessionStore.GetSessionByID(code)
	if err != nil {
		errorResponse(rw, InvalidGrant, fmt.Sprintf("Invalid code: %s", code),
			http.StatusUnauthorized)
		return nil, false
	}

	return session, true
}

func (m *MockOIDC) setTokens(tr *tokenResponse, s *Session, grantType string) error {
	var err error
	tr.AccessToken, err = s.AccessToken(m.Config(), m.Keypair, m.Clock.NowUTC())
	if err != nil {
		return err
	}
	if len(s.Scopes) > 0 && s.Scopes[0] == openidScope {
		tr.IDToken, err = s.IDToken(m.Config(), m.Keypair, m.Clock.NowUTC())
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MockOIDC) Userinfo(rw http.ResponseWriter, req *http.Request) {
	token, authorized := m.authorizeBearer(rw, req)
	if !authorized {
		return
	}

	session, err := m.SessionStore.GetSessionByToken(token)
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	resp, err := session.User.Userinfo(session.Scopes)
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}
	jsonResponse(rw, resp)
}

type discoveryResponse struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSUri               string `json:"jwks_uri"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`

	GrantTypesSupported               []string `json:"grant_types_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
}

func (m *MockOIDC) Discovery(rw http.ResponseWriter, _ *http.Request) {
	discovery := &discoveryResponse{
		Issuer:                m.Issuer(),
		AuthorizationEndpoint: m.AuthorizationEndpoint(),
		TokenEndpoint:         m.TokenEndpoint(),
		JWKSUri:               m.JWKSEndpoint(),
		UserinfoEndpoint:      m.UserinfoEndpoint(),

		GrantTypesSupported:               GrantTypesSupported,
		ResponseTypesSupported:            ResponseTypesSupported,
		SubjectTypesSupported:             SubjectTypesSupported,
		IDTokenSigningAlgValuesSupported:  IDTokenSigningAlgValuesSupported,
		ScopesSupported:                   m.Provider.ScopesSupported,
		TokenEndpointAuthMethodsSupported: TokenEndpointAuthMethodsSupported,
		ClaimsSupported:                   ClaimsSupported,
		CodeChallengeMethodsSupported:     m.CodeChallengeMethodsSupported,
	}

	resp, err := json.Marshal(discovery)
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}
	jsonResponse(rw, resp)
}

func (m *MockOIDC) JWKS(rw http.ResponseWriter, _ *http.Request) {
	jwks, err := m.Keypair.JWKS()
	if err != nil {
		internalServerError(rw, err.Error())
		return
	}

	jsonResponse(rw, jwks)
}

func (m *MockOIDC) authorizeBearer(rw http.ResponseWriter, req *http.Request) (*jwt.Token, bool) {
	header := req.Header.Get("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) < 2 || parts[0] != "Bearer" {
		errorResponse(rw, InvalidRequest, "Invalid authorization header",
			http.StatusUnauthorized)
		return nil, false
	}

	return m.authorizeToken(parts[1], rw)
}

func (m *MockOIDC) authorizeToken(t string, rw http.ResponseWriter) (*jwt.Token, bool) {
	token, err := m.Keypair.VerifyJWT(t, m.Clock.NowUTC)
	if err != nil {
		errorResponse(rw, InvalidRequest, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		internalServerError(rw, "Unable to extract token claims")
		return nil, false
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		internalServerError(rw, "Unable to extract token expiration")
		return nil, false
	}
	if m.Clock.NowUTC().Unix() > int64(exp) {
		errorResponse(rw, InvalidRequest, "The token is expired", http.StatusUnauthorized)
		return nil, false
	}
	return token, true
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

func assertEqual(param, value, errorType, errorMsg string, rw http.ResponseWriter, req *http.Request) bool {
	formValue := req.Form.Get(param)
	if subtle.ConstantTimeCompare([]byte(value), []byte(formValue)) == 0 {
		errorResponse(rw, errorType, fmt.Sprintf("%s: %s", errorMsg, formValue),
			http.StatusUnauthorized)
		return false
	}
	return true
}

func validateScope(supported []string, rw http.ResponseWriter, req *http.Request) bool {
	allowed := make(map[string]struct{})
	for _, scope := range supported {
		allowed[scope] = struct{}{}
	}

	scopes := strings.Split(req.Form.Get("scope"), " ")
	for _, scope := range scopes {
		if _, ok := allowed[scope]; !ok {
			errorResponse(rw, InvalidScope, fmt.Sprintf("Unsupported scope: %s", scope),
				http.StatusBadRequest)
			return false
		}
	}
	return true
}

func validateCodeChallengeMethodSupported(rw http.ResponseWriter, method string, supportedMethods []string) bool {
	if method != "" && !contains(method, supportedMethods) {
		errorResponse(rw, InvalidRequest, "Invalid code challenge method", http.StatusBadRequest)
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

func contains(value string, list []string) bool {
	for _, element := range list {
		if element == value {
			return true
		}
	}
	return false
}
