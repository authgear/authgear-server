package api

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AnonymousUserSignupAPIHandler), "*"),
	wire.Struct(new(AnonymousUserPromotionCodeAPIHandler), "*"),
	wire.Struct(new(PresignImagesUploadHandler), "*"),

	wire.Struct(new(WorkflowNewHandler), "*"),
	wire.Struct(new(WorkflowGetHandler), "*"),
	wire.Struct(new(WorkflowInputHandler), "*"),
	wire.Struct(new(WorkflowWebsocketHandler), "*"),
	wire.Struct(new(WorkflowV2Handler), "*"),

	wire.Struct(new(AuthenticationFlowV1CreateHandler), "*"),
	wire.Struct(new(AuthenticationFlowV1InputHandler), "*"),
	wire.Struct(new(AuthenticationFlowV1GetHandler), "*"),
	wire.Struct(new(AuthenticationFlowV1WebsocketHandler), "*"),

	wire.Struct(new(AccountManagementV1IdentificationHandler), "*"),
	wire.Struct(new(AccountManagementV1IdentificationOAuthHandler), "*"),
)
