package preprocessor

import (
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
)

type ConnPreprocessor struct {
	AppName  string
	DBOpener func(string, string, string) (skydb.Conn, error)
	DBImpl   string
	Option   string
}

func (p ConnPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	log.Debugf("Opening DBConn: {%v %v %v}", p.DBImpl, p.AppName, p.Option)

	conn, err := p.DBOpener(p.DBImpl, p.AppName, p.Option)
	if err != nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedUnableToOpenDatabase, err.Error())
		return http.StatusServiceUnavailable
	}
	payload.DBConn = conn

	log.Debugf("Get DB OK")

	return http.StatusOK
}
