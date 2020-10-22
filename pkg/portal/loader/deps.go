package loader

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewUserLoader,
	NewAppLoader,
	NewDomainLoader,
	NewCollaboratorLoader,
	NewCollaboratorInvitationLoader,
)
