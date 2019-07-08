package hook

import (
	"encoding/json"
	"io"

	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type AuthPayload struct {
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
	Context map[string]interface{} `json:"context"`
}

func NewDefaultAuthPayload(
	event string,
	user model.User,
	requestID string,
	path string,
	reqPayload interface{},
) (AuthPayload, error) {
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

	reqContext := map[string]interface{}{
		"id":      requestID,
		"path":    path,
		"payload": reqPayload,
	}

	context := map[string]interface{}{
		"req": reqContext,
	}

	payload = AuthPayload{
		Event:   event,
		Data:    data,
		Context: context,
	}

	return payload, nil
}

func (a AuthPayload) Encode() interface{} {
	return a
}

type AuthRespPayload struct {
	User *model.User
}

func (a *AuthRespPayload) Decode(r io.Reader) error {
	return json.NewDecoder(r).Decode(a.User)
}
