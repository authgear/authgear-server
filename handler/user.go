package handler

import (
	"github.com/mitchellh/mapstructure"
	"github.com/oursky/ourd/oderr"
	"github.com/oursky/ourd/router"
)

type queryPayload struct {
	Emails []string `json:"emails"`
}

type updatePayload struct {
	Email string `json:"email"`
}

func UserQueryHandler(payload *router.Payload, response *router.Response) {
	qp := queryPayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &qp,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}

	if err := mapDecoder.Decode(payload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	userinfos, err := payload.DBConn.QueryUser(qp.Emails)
	if err != nil {
		response.Err = err
		return
	}

	results := make([]interface{}, len(userinfos))
	for i, userinfo := range userinfos {
		results[i] = map[string]interface{}{
			"id":   userinfo.ID,
			"type": "user",
			"data": struct {
				ID    string `json:"id"`
				Email string `json:"email"`
			}{userinfo.ID, userinfo.Email},
		}
	}
	response.Result = results
	return
}

func UserUpdateHandler(payload *router.Payload, response *router.Response) {
	p := updatePayload{}
	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &p,
		TagName: "json",
	})
	if err != nil {
		panic(err)
	}

	if err := mapDecoder.Decode(payload.Data); err != nil {
		response.Err = oderr.NewRequestInvalidErr(err)
		return
	}

	payload.UserInfo.Email = p.Email

	if err := payload.DBConn.UpdateUser(payload.UserInfo); err != nil {
		response.Err = err
		return
	}

	return
}
