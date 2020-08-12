package deps

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/auth/task"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
)

var taskDeps = wire.NewSet(
	wire.NewSet(
		commonDeps,
		mail.DependencySet,
		sms.DependencySet,
	),

	task.DependencySet,
	wire.Bind(new(task.MailSender), new(*mail.Sender)),
	wire.Bind(new(task.SMSClient), new(*sms.Client)),
)
