package handler

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type rolePayload struct {
	Roles []string `json:"roles"`
}

func (payload *rolePayload) Decode(data map[string]interface{}) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  payload,
		TagName: "json",
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(data); err != nil {
		return err
	}
	if payload.Roles == nil {
		return errors.New("Missing roles key in request")
	}
	return nil
}

// RoleDefaultHandler enable system administrator to set default user role
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "role:default",
//     "master_key": "MASTER_KEY",
//     "access_token": "ACCESS_TOKEN",
//     "roles": [
//        "writer",
//        "user"
//     ]
// }
// EOF
//
// {
//     "result": [
//        "writer",
//        "user"
//     ]
// }
type RoleDefaultHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *RoleDefaultHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
	}
}

func (h *RoleDefaultHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RoleDefaultHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("RoleDefaultHandler %v", h)
	payload := &rolePayload{}
	err := payload.Decode(rpayload.Data)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}

	err = rpayload.DBConn.SetDefaultRoles(payload.Roles)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
	}
	response.Result = payload.Roles
}

// RoleAdminHandler enable system administrator to set which roles can perform
// administrative action, like change others user role.
// curl -X POST -H "Content-Type: application/json" \
//   -d @- http://localhost:3000/ <<EOF
// {
//     "action": "role:default",
//     "master_key": "MASTER_KEY",
//     "access_token": "ACCESS_TOKEN",
//     "roles": [
//        "admin",
//        "moderator"
//     ]
// }
// EOF
//
// {
//     "result": [
//        "admin",
//        "moderator"
//     ]
// }
type RoleAdminHandler struct {
	DevOnly       router.Processor `preprocessor:"dev_only"`
	DBConn        router.Processor `preprocessor:"dbconn"`
	preprocessors []router.Processor
}

func (h *RoleAdminHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.DevOnly,
		h.DBConn,
	}
}

func (h *RoleAdminHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *RoleAdminHandler) Handle(rpayload *router.Payload, response *router.Response) {
	log.Debugf("RoleAdminHandler %v", h)
	payload := &rolePayload{}
	err := payload.Decode(rpayload.Data)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.BadRequest, err.Error())
		return
	}

	err = rpayload.DBConn.SetAdminRoles(payload.Roles)
	if err != nil {
		response.Err = skyerr.NewUnknownErr(err)
	}
	response.Result = payload.Roles
}
