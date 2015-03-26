package main

import (
	"errors"
	"github.com/oursky/ourd/oderr"
	"net/http"
	log "github.com/Sirupsen/logrus"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/router"
)

type apiKeyValidatonPreprocessor struct {
	Key string
}

func (p apiKeyValidatonPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	apiKey := payload.APIKey()
	if apiKey != p.Key {
		log.Debugf("Invalid APIKEY: %v", apiKey)
		response.Err = oderr.NewFmt(oderr.CannotVerifyAPIKey, "Cannot verify api key: %v", apiKey)
		return http.StatusUnauthorized
	}

	return http.StatusOK
}

type connPreprocessor struct {
	DBOpener func(string, string, string) (oddb.Conn, error)
	DBImpl   string
	AppName  string
	Option   string
}

func (p connPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	log.Debugf("GetDB Conn")

	conn, err := p.DBOpener(p.DBImpl, p.AppName, p.Option)
	if err != nil {
		response.Err = err
		return http.StatusServiceUnavailable
	}
	payload.DBConn = conn

	log.Debugf("Get DB OK")

	return http.StatusOK
}

type tokenStorePreprocessor struct {
	authtoken.Store
}

func (p tokenStorePreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	payload.TokenStore = p.Store
	return http.StatusOK
}

func authenticateUser(payload *router.Payload, response *router.Response) int {
	tokenString := payload.AccessToken()
	if tokenString == "" { // no access token, leave it unauthenticated
		return http.StatusOK
	}

	store := payload.TokenStore
	token := authtoken.Token{}

	if err := store.Get(tokenString, &token); err != nil {
		response.Err = err
		return http.StatusUnauthorized
	}

	conn := payload.DBConn
	userinfo := oddb.UserInfo{}
	if err := conn.GetUser(token.UserInfoID, &userinfo); err != nil {
		// we got a valid access token but cannot find the user specified
		// there must be data inconsistency.
		log.Printf("Cannot find UserInfo via connection specified in Token = %#v", token)
		response.Err = err
		return http.StatusInternalServerError
	}
	payload.UserInfo = &userinfo

	return http.StatusOK
}

func injectDatabase(payload *router.Payload, response *router.Response) int {
	conn := payload.DBConn

	databaseID, _ := payload.Data["database_id"].(string)
	switch databaseID {
	case "_public":
		payload.Database = conn.PublicDB()
	case "_private":
		if payload.UserInfo != nil {
			payload.Database = conn.PrivateDB(payload.UserInfo.ID)
		} else {
			response.Err = errors.New("Authentication is needed for private DB access")
			return http.StatusUnauthorized
		}
	}

	return http.StatusOK
}
