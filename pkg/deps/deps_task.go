package deps

import (
	"github.com/google/wire"

	"github.com/skygeario/skygear-server/pkg/auth/task"
	"github.com/skygeario/skygear-server/pkg/mail"
	"github.com/skygeario/skygear-server/pkg/sms"
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
