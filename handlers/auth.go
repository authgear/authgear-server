package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type loginJSON struct {
	Email    interface{} `json:"email"`
	Password interface{} `json:"password"`
}

type responseJSON struct {
	UserID      interface{} `json:"user_id"`
	AccessToken interface{} `json:"access_token,omitempty"`
}


// LoginHandler is dummy implementation on handling login
// curl -X POST -H "Content-Type: application/json" http://localhost:3000/ -d '{"action":"auth:login"}'
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var (
		httpStatus = http.StatusOK
		reqJSON    loginJSON
		respJSON   responseJSON
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			w.Write([]byte(errString))
		}
	}()
	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, &reqJSON); err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
		return
	}

	respJSON.UserID = "rickmak-oursky"
	b, err := json.Marshal(respJSON)
	if err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
	}
	w.Write(b)
}
