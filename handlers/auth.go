package handlers

import (
	"encoding/json"
	"net/http"
)

type loginJSON struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *loginJSON) RouteAction() string {
	return "auth:login"
}

type responseJSON struct {
	UserID      string `json:"user_id"`
	AccessToken string `json:"access_token,omitempty"`
}

// LoginHandler is dummy implementation on handling login
// curl -X POST -H "Content-Type: application/json" http://localhost:3000/ -d '{"action":"auth:login"}'
func LoginHandler(response Responser, playload Payloader) {
	var (
		httpStatus = http.StatusOK
		respJSON   responseJSON
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			response.Write([]byte(errString))
		}
	}()

	respJSON.UserID = "rickmak-oursky"
	b, err := json.Marshal(respJSON)
	if err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
	}
	response.Write(b)
}
