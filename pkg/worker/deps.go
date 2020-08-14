package worker

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms"
	"github.com/authgear/authgear-server/pkg/worker/tasks"
)

var DependencySet = wire.NewSet(
	deps.TaskDependencySet,
	deps.CommonDependencySet,

	mail.DependencySet,
	sms.DependencySet,

	tasks.DependencySet,
	wire.Bind(new(tasks.MailSender), new(*mail.Sender)),
	wire.Bind(new(tasks.SMSClient), new(*sms.Client)),
)
