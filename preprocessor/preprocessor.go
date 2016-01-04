package preprocessor

import (
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

func InjectUserIfPresent(payload *router.Payload, response *router.Response) int {
	if payload.UserInfoID == "" {
		log.Debugln("injectUser: empty UserInfoID, skipping")
		return http.StatusOK
	}

	conn := payload.DBConn
	userinfo := skydb.UserInfo{}
	if err := conn.GetUser(payload.UserInfoID, &userinfo); err != nil {
		log.Errorf("Cannot find UserInfo.ID = %#v\n", payload.UserInfoID)
		response.Err = skyerr.NewError(skyerr.UnexpectedUserInfoNotFound, err.Error())
		return http.StatusInternalServerError
	}

	payload.UserInfo = &userinfo

	return http.StatusOK
}

func InjectDatabase(payload *router.Payload, response *router.Response) int {
	conn := payload.DBConn

	databaseID, _ := payload.Data["database_id"].(string)
	switch databaseID {
	case "_public":
		payload.Database = conn.PublicDB()
	case "_private":
		if payload.UserInfo != nil {
			payload.Database = conn.PrivateDB(payload.UserInfo.ID)
		} else {
			response.Err = skyerr.NewError(skyerr.NotAuthenticated, "Authentication is needed for private DB access")
			return http.StatusUnauthorized
		}
	}

	return http.StatusOK
}

func RequireUserForWrite(payload *router.Payload, response *router.Response) int {
	if payload.UserInfo == nil {
		response.Err = skyerr.NewError(skyerr.PermissionDenied, "write is not allowed")
		return http.StatusUnauthorized
	}

	return http.StatusOK
}
