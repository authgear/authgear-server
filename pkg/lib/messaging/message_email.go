package messaging

import (
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
)

type EmailMessage struct {
	message
	taskQueue task.Queue
	events    EventService

	Type nonblocking.MessageType
	mail.SendOptions
}

func (m *EmailMessage) Send() error {
	err := m.events.DispatchEvent(&nonblocking.EmailSentEventPayload{
		Sender:    m.Sender,
		Recipient: m.Recipient,
		Type:      m.Type,
	})
	if err != nil {
		return err
	}

	m.taskQueue.Enqueue(&tasks.SendMessagesParam{
		EmailMessages: []mail.SendOptions{m.SendOptions},
	})
	m.isSent = true

	return nil
}
