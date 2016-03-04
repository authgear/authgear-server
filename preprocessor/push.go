// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
