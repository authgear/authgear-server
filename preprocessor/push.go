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
	if p.NotificationSender == nil {
		response.Err = skyerr.NewError(skyerr.UnexpectedPushNotificationNotConfigured, "Unable to send push notification because APNS is not configured or there was a problem configuring the APNS.\n")
		return http.StatusInternalServerError
	}

	payload.NotificationSender = p.NotificationSender
	return http.StatusOK
}
