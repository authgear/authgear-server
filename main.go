package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	r := NewRouter()
	r.HandleFunc("", HomeHandler)
	r.HandleFunc("auth:login", LoginHandler)
	log.Println("Listening...")
	http.ListenAndServe(":3000", r)
}

// HomeHandler temp landing. FIXME
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Developer"))
}

type loginJSON struct {
	Email    interface{} `json:"email"`
	Password interface{} `json:"password"`
}

type responseJSON struct {
	UserID      interface{} `json:"user_id"`
	AccessToken interface{} `json:"access_token,omitempty"`
}

// LoginHandler is dummy implementation on handling login
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
