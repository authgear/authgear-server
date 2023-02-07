package api

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(AnonymousUserSignupAPIHandler), "*"),
	NewAnonymousUserSignupAPIHandlerLogger,
	wire.Struct(new(AnonymousUserPromotionCodeAPIHandler), "*"),
	NewAnonymousUserPromotionCodeAPILogger,
	wire.Struct(new(PresignImagesUploadHandler), "*"),
	NewPresignImagesUploadHandlerLogger,
	wire.Struct(new(MagicLinkVerificationAPIHandler), "*"),
	NewMagicLinkVerificationAPILogger,

	wire.Struct(new(WorkflowNewHandler), "*"),
	wire.Struct(new(WorkflowGetHandler), "*"),
	wire.Struct(new(WorkflowInputHandler), "*"),
)
