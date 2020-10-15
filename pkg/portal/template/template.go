package template

import (
	"github.com/authgear/authgear-server/pkg/util/template"
)

const (
	TemplateItemTypePortalCollaboratorInvitationEmailTXT  string = "portal_collaborator_invitation_email.txt"
	TemplateItemTypePortalCollaboratorInvitationEmailHTML string = "portal_collaborator_invitation_email.html"
)

var TemplatePortalCollaboratorInvitationEmailTXT = template.Register(template.T{
	Type: TemplateItemTypePortalCollaboratorInvitationEmailTXT,
})

var TemplatePortalCollaboratorInvitationEmailHTML = template.Register(template.T{
	Type:   TemplateItemTypePortalCollaboratorInvitationEmailHTML,
	IsHTML: true,
})
