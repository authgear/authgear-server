package fraudprotection

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Service), "*"),
	wire.Struct(new(MetricsStore), "*"),
	wire.Struct(new(LeakyBucketStore), "*"),
	wire.Bind(new(MetricsQuerier), new(*MetricsStore)),
	wire.Bind(new(LeakyBucketer), new(*LeakyBucketStore)),
)
