package messaging

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
)

type SMSMessage struct {
	message
	taskQueue task.Queue
	events    EventService

	Type nonblocking.MessageType
	sms.SendOptions

	IsNotBilled bool
}

func (m *SMSMessage) Send() error {
	err := m.events.DispatchEvent(&nonblocking.SMSSentEventPayload{
		Sender:      m.Sender,
		Recipient:   m.To,
		Type:        m.Type,
		IsNotBilled: m.IsNotBilled,
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
