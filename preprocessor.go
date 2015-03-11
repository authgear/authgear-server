package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/router"
)

type connPreprocessor struct {
	DBOpener func(string, string, string) (oddb.Conn, error)
	DBImpl   string
	AppName  string
	Option   string
}

func (p connPreprocessor) Preprocess(payload *router.Payload, response *router.Response) (int, error) {
	log.Println("GetDB Conn")

	conn, err := p.DBOpener(p.DBImpl, p.AppName, p.Option)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	payload.DBConn = conn

	log.Println("Get DB OK")

	return http.StatusOK, nil
}

type tokenStorePreprocessor struct {
	authtoken.Store
}

func (p tokenStorePreprocessor) Preprocess(payload *router.Payload, response *router.Response) (int, error) {
	payload.TokenStore = p.Store
	return http.StatusOK, nil
}

func authenticateUser(payload *router.Payload, response *router.Response) (int, error) {
	tokenString := payload.AccessToken()
	if tokenString == "" { // no access token, leave it unauthenticated
		return http.StatusOK, nil
	}

	store := payload.TokenStore
	token := authtoken.Token{}

	if err := store.Get(tokenString, &token); err != nil {
		return http.StatusUnauthorized, err
	}

	conn := payload.DBConn
	userinfo := oddb.UserInfo{}
	if err := conn.GetUser(token.UserInfoID, &userinfo); err != nil {
		// we got a valid access token but cannot find the user specified
		// there must be data inconsistency.
		log.Printf("Cannot find UserInfo via connection specified in Token = %#v", token)
		return http.StatusInternalServerError, err
	}
	payload.UserInfo = &userinfo

	return http.StatusOK, nil
}

func injectDatabase(payload *router.Payload, response *router.Response) (int, error) {
	conn := payload.DBConn

	databaseID, _ := payload.Data["database_id"].(string)
	switch databaseID {
	case "_public":
		payload.Database = conn.PublicDB()
	case "_private":
		if payload.UserInfo != nil {
			payload.Database = conn.PrivateDB(payload.UserInfo.ID)
		} else {
			return http.StatusUnauthorized, errors.New("Authentication is needed for private DB access")
		}
	}

	return http.StatusOK, nil
}
