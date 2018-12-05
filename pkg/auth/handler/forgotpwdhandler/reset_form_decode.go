package forgotpwdhandler

import (
	"net/http"
	"strconv"
	"time"
)

func decodeForgotPasswordResetFormRequest(request *http.Request) (payload ForgotPasswordResetPayload, err error) {
	if err = request.ParseForm(); err != nil {
		return
	}

	p := ForgotPasswordResetPayload{}
	p.UserID = request.Form.Get("user_id")
	p.Code = request.Form.Get("code")
	p.NewPassword = request.Form.Get("new_password")

	expireAtStr := request.Form.Get("expire_at")
	var expireAt int64
	if expireAtStr != "" {
		if expireAt, err = strconv.ParseInt(expireAtStr, 10, 64); err != nil {
			return
		}
	}

	p.ExpireAt = expireAt
	p.ExpireAtTime = time.Unix(expireAt, 0).UTC()

	payload = p
	return
}
