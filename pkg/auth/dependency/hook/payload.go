package hook

import (
	"encoding/json"
	"io"

	"github.com/skygeario/skygear-server/pkg/auth/response"
)

type ReqBodyEncoder interface {
	Encode() interface{}
}

type RespDecoder interface {
	Decode(r io.Reader) error
}

type AuthPayload struct {
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
	Context map[string]interface{} `json:"context"`
}

func NewDefaultAuthPayload(event string, user response.User) (AuthPayload, error) {
	var data map[string]interface{}
	var userBytes []byte
	var err error
	var payload AuthPayload
	userBytes, err = json.Marshal(user)
	if err != nil {
		return payload, err
	}
	err = json.Unmarshal(userBytes, &data)
	if err != nil {
		return payload, err
	}

	payload = AuthPayload{
		Event: event,
		Data:  data,
	}

	return payload, nil
}

func (a AuthPayload) Encode() interface{} {
	return a
}

type AuthRespPayload struct {
	User *response.User
}

func (a *AuthRespPayload) Decode(r io.Reader) error {
	return json.NewDecoder(r).Decode(a.User)
}
