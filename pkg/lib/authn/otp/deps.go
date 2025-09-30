package otp

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(MessageSender), "*"),
	wire.Struct(new(CodeStoreRedis), "*"),
	wire.Struct(new(LookupStoreRedis), "*"),
	wire.Struct(new(AttemptTrackerRedis), "*"),
	wire.Bind(new(CodeStore), new(*CodeStoreRedis)),
	wire.Bind(new(SenderCodeStore), new(*CodeStoreRedis)),
	wire.Bind(new(LookupStore), new(*LookupStoreRedis)),
	wire.Bind(new(AttemptTracker), new(*AttemptTrackerRedis)),
)
