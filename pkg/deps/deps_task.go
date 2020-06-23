package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/mail"
	"github.com/skygeario/skygear-server/pkg/sms"
)

var taskDeps = wire.NewSet(
	commonDeps,

	task.DependencySet,
	wire.NewSet(
		mail.DependencySet,
		wire.Bind(new(task.MailSender), new(*mail.Sender)),
	),
	wire.NewSet(
		sms.DependencySet,
		wire.Bind(new(task.SMSClient), new(*sms.Client)),
	),
)
