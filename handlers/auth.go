package handlers

import (
	"encoding/json"
	"log"
)

type authResponse struct {
	UserID      string `json:"user_id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type signupPayload struct {
	Data map[string]interface{}
	Raw []byte
}

func (p *signupPayload) RouteAction() string {
	return "auth:signup"
}

func (p *signupPayload) Email() string {
	return p.Data["email"].(string)
}

func (p *signupPayload) Password() string {
	return p.Data["password"].(string)
}

/*
SignupHandler is dummy implementation on handling user signup
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "auth:signup",
    "email": "rick.mak@gmail.com",
    "password": "123456"
}
EOF
*/
func SignupHandler(response Responser, playload Payload) {
	var (
		resp authResponse
	)
	defer func() {
		if e := recover(); e != nil {
			log.Println("Error ", e)
			response.Write([]byte(e.(string)))
		}
	}()

	resp.UserID = "rickmak-oursky"
	resp.AccessToken = "validToken"
	b, err := json.Marshal(resp)
	if err != nil {
		panic("Response Error: " + err.Error())
	}
	response.Write(b)
}

type loginPayload struct {
	Data map[string]interface{}
	Raw []byte
}

func (p *loginPayload) RouteAction() string {
	return "auth:login"
}

func (p *loginPayload) Email() string {
	return p.Data["email"].(string)
}

func (p *loginPayload) Password() string {
	return p.Data["password"].(string)
}

/*
LoginHandler is dummy implementation on handling login
curl -X POST -H "Content-Type: application/json" \
  -d @- http://localhost:3000/ <<EOF
{
    "action": "auth:login",
    "email": "rick.mak@gmail.com",
    "password": "123456"
}
EOF
*/
func LoginHandler(response Responser, playload Payload) {
	var (
		resp authResponse
		p    loginPayload
	)
	defer func() {
		if e := recover(); e != nil {
			log.Println("Error ", e)
			response.Write([]byte(e.(string)))
		}
	}()

	p = loginPayload(playload)
	if p.Email() != "rick.mak@gmail.com" {
		panic("User Not exist")
	}
	resp.UserID = "rickmak-oursky"
	resp.AccessToken = "validToken"
	b, err := json.Marshal(resp)
	if err != nil {
		panic("Response Error: " + err.Error())
	}
	response.Write(b)
}
