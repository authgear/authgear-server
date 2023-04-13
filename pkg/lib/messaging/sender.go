package messaging

import (
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("messaging")}
}

type EventService interface {
	DispatchEvent(payload event.Payload) error
}

type Sender struct {
	RateLimits RateLimits
	TaskQueue  task.Queue
	Events     EventService
}

func (s *Sender) PrepareEmail(email string, msgType nonblocking.MessageType) (*EmailMessage, error) {
	msg, err := s.RateLimits.checkEmail(email)
	if err != nil {
		return nil, err
	}

	return &EmailMessage{
		message:     *msg,
		taskQueue:   s.TaskQueue,
		events:      s.Events,
		SendOptions: mail.SendOptions{Recipient: email},
		Type:        msgType,
	}, nil
}

func (s *Sender) PrepareSMS(phoneNumber string, msgType nonblocking.MessageType) (*SMSMessage, error) {
	msg, err := s.RateLimits.checkSMS(phoneNumber)
	if err != nil {
		return nil, err
	}

	return &SMSMessage{
		message:     *msg,
		taskQueue:   s.TaskQueue,
		events:      s.Events,
		SendOptions: sms.SendOptions{To: phoneNumber},
		Type:        msgType,
	}, nil
}
