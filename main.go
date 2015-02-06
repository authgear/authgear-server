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

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Developer"))
}

type RequestJson struct {
	Action   interface{} `json:"action"`
	ApiKey   interface{} `json:"api_key"`
	Email    interface{} `json:"email"`
	Password interface{} `json:"password"`
}

type ResponseJson struct {
	UserId      interface{} `json:"user_id"`
	AccessToken interface{} `json:"access_token,omitempty"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var (
		httpStatus = http.StatusOK
		reqJson    RequestJson
		respJson   ResponseJson
		errString  string
	)
	defer func() {
		if httpStatus != http.StatusOK {
			w.Write([]byte(errString))
		}
	}()
	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, &reqJson); err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
		return
	}

	respJson.UserId = "rickmak-oursky"
	b, err := json.Marshal(respJson)
	if err != nil {
		httpStatus = http.StatusBadRequest
		errString = err.Error()
	}
	w.Write(b)
}
