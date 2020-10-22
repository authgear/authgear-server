package resource

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

var TemplateCollaboratorInvitationEmailTXT = PortalRegistry.Register(&template.PlainText{
	Name: "messages/collaborator_invitation_email.txt",
}).(*template.PlainText)

var TemplateCollaboratorInvitationEmailHTML = PortalRegistry.Register(&template.HTML{
	Name: "messages/collaborator_invitation_email.html",
}).(*template.HTML)
