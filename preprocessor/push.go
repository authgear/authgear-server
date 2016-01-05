package preprocessor

import (
	"net/http"

	"github.com/oursky/skygear/push"
	"github.com/oursky/skygear/router"
	"github.com/oursky/skygear/skyerr"
)

type NotificationPreprocessor struct {
	NotificationSender push.Sender
}

func (p NotificationPreprocessor) Preprocess(payload *router.Payload, response *router.Response) int {
	routeSender, ok := p.NotificationSender.(push.RouteSender)
	if !ok {
		response.Err = skyerr.NewError(skyerr.UnexpectedPushNotificationNotConfigured, "Unknown notification sender.")
		return http.StatusInternalServerError
	}

	if routeSender.Len() == 0 {
		response.Err = skyerr.NewError(skyerr.UnexpectedPushNotificationNotConfigured, "Unable to send push notification because APNS is not configured or there was a problem configuring the APNS.\n")
		return http.StatusInternalServerError
	}
	return http.StatusOK
}
