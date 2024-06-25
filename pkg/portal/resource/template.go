package resource

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateCollaboratorInvitationEmailTXT = PortalRegistry.Register(&template.MessagePlainText{
	Name: "messages/collaborator_invitation_email.txt",
}).(*template.MessagePlainText)

var TemplateCollaboratorInvitationEmailHTML = PortalRegistry.Register(&template.MessageHTML{
	Name: "messages/collaborator_invitation_email.html",
}).(*template.MessageHTML)
