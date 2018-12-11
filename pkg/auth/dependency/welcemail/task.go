package welcemail

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
)

type SendTask struct {
	Response chan error

	Context context.Context
	Sender

	executed bool
}

func NewSendTask(ctx context.Context, sender Sender) *SendTask {
	return &SendTask{
		Response: make(chan error),
		Context:  ctx,
		Sender:   sender,
		executed: false,
	}
}

func (s *SendTask) Execute(
	email string,
	userProfile userprofile.UserProfile,
	logger *logrus.Entry,
) {
	if s.executed {
		panic(errors.New("SendTask cannot be executed more than once"))
	}
	s.executed = true

	go func() {
		logger.WithFields(logrus.Fields{
			"email": email,
		}).Info("start sending welcome email")

		var err error
		if err = s.Send(email, userProfile); err != nil {
			logger.WithFields(logrus.Fields{
				"error":  err,
				"email":  email,
				"userID": userProfile.ID,
			}).Error("fail to send welcome email")
		}

		select {
		case <-s.Context.Done(): // return if no one receive the error
		default:
			s.Response <- err
		}
	}()
}
