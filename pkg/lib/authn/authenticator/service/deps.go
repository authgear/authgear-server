package service

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(AntiBruteForceAuthenticateBucketMaker), "*"),
	wire.Struct(new(Service), "*"),
)
