package messaging

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/lib/translation"
)

type SMSMessage struct {
	message
	taskQueue task.Queue
	events    EventService

	Type translation.MessageType
	sms.SendOptions

	IsNotCounted bool
}

func (m *SMSMessage) Send() error {
	err := m.events.DispatchEventImmediately(&nonblocking.SMSSentEventPayload{
		Sender:              m.Sender,
		Recipient:           m.To,
		Type:                string(m.Type),
		IsNotCountedInUsage: m.IsNotCounted,
	})
	if err != nil {
		return err
	}

	m.taskQueue.Enqueue(&tasks.SendMessagesParam{
		SMSMessages: []sms.SendOptions{m.SendOptions},
	})
	m.isSent = true

	return nil
}
