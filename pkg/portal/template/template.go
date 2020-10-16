package template

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateCollaboratorInvitationEmailTXT = template.RegisterPlainText(
	"messages/collaborator_invitation_email.txt",
)

var TemplateCollaboratorInvitationEmailHTML = template.RegisterPlainText(
	"messages/collaborator_invitation_email.html",
)
