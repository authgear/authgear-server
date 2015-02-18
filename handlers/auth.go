package handlers

import (
	"github.com/oursky/ourd/oddb"
	"github.com/twinj/uuid"
	"log"
)

type authResponse struct {
	UserID      string `json:"user_id,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

type signupPayload struct {
	Meta map[string]interface{}
	Data map[string]interface{}
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
func SignupHandler(playload *Payload, response *Response) {
	var (
		resp authResponse
	)
	log.Println("SignupHandler")
	var p = signupPayload{
		Meta: playload.Meta,
		Data: playload.Data,
	}
	// TODO: check user not exist already
	data := make(map[string]interface{})
	data["email"] = p.Email()
	// TODO: hash the password
	data["password"] = p.Password()
	u := uuid.NewV4()
	user := oddb.Record{
		"user",
		"user:" + uuid.Formatter(u, uuid.Clean),
		data,
	}
	playload.DBConn.PublicDB().Save(&user)
	resp.UserID = "rickmak-oursky"
	resp.AccessToken = "validToken"
	response.Result = resp
	return
}

type loginPayload struct {
	Meta map[string]interface{}
	Data map[string]interface{}
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
func LoginHandler(playload *Payload, response *Response) {
	var (
		resp authResponse
	)
	log.Println("LoginHandler")
	var p = loginPayload{
		Meta: playload.Meta,
		Data: playload.Data,
	}
	if p.Email() != "rick.mak@gmail.com" {
		panic("User Not exist")
	}
	resp.UserID = "rickmak-oursky"
	resp.AccessToken = "validToken"
	response.Result = resp
	return
}
