package preprocessor

import (
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type DevOnlyProcessor struct {
	DevMode bool
}

func (p DevOnlyProcessor) Preprocess(payload *router.Payload, response *router.Response) int {
	if !p.DevMode {
		log.Infof("Attempt to access dev only end-point")
		response.Err = skyerr.NewError(skyerr.PermissionDenied,
			"Attempt to access dev only end-point")
		return http.StatusForbidden
	}
	return http.StatusOK
}
