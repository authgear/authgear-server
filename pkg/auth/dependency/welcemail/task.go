package welcemail

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
)

type SendTaskRequest struct {
	Email       string
	UserProfile userprofile.UserProfile
	Logger      *logrus.Entry
}

type SendTask struct {
	Request  chan SendTaskRequest
	Response chan error

	Sender
}

func NewSendTask(sender Sender) *SendTask {
	return &SendTask{
		Request:  make(chan SendTaskRequest),
		Response: make(chan error),
		Sender:   sender,
	}
}

func (s *SendTask) WaitForRequest(ctx context.Context) {
	go func() {
		var req SendTaskRequest
		select {
		case req = <-s.Request:
		case <-ctx.Done():
			// early return if context is cancelled
			return
		}

		req.Logger.WithFields(logrus.Fields{
			"email": req.Email,
		}).Info("start sending welcome email")

		var err error
		if err = s.Send(req.Email, req.UserProfile); err != nil {
			req.Logger.WithFields(logrus.Fields{
				"error":  err,
				"email":  req.Email,
				"userID": req.UserProfile.ID,
			}).Error("fail to send welcome email")
		}

		select {
		case <-ctx.Done(): // return if no one receive the error
		default:
			s.Response <- err
		}
	}()
}
