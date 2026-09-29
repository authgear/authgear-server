package audit

import (
	"github.com/google/wire"

	"github.com/authgear/authgear-server/pkg/lib/telemetry/auditlogstreaming"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(Sink), "*"),
	wire.Struct(new(ReadStore), "*"),
	wire.Struct(new(WriteStore), "*"),
	wire.Struct(new(Query), "*"),
	auditlogstreaming.ProducerDependencySet,
)
