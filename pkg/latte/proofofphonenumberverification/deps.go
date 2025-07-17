package proofofphonenumberverification

import (
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	NewHookHTTPClient,
	wire.Struct(new(Service), "*"),
	wire.Struct(new(ProofOfPhoneNumberVerificationWebHook), "*"),
)
