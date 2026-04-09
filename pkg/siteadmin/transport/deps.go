package transport

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(AppsListHandler), "*"),
	wire.Struct(new(AppGetHandler), "*"),
	wire.Struct(new(CollaboratorsListHandler), "*"),
	wire.Struct(new(CollaboratorAddHandler), "*"),
	wire.Struct(new(CollaboratorRemoveHandler), "*"),
	wire.Struct(new(MessagingUsageHandler), "*"),
	wire.Struct(new(MonthlyActiveUsersUsageHandler), "*"),
	wire.Struct(new(AuthzMiddleware), "*"),
)
