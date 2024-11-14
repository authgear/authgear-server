package oauthrelyingpartyutil

import (
	"context"
)

// jwx@v2 takes context in some of the APIs.
// But jwx@v3 remove context in those APIs.
// And actually, the context argument is unused.
// So we can safely pass it a context.TODO().
var ContextForTheUnusedContextArgumentInJWXV2API = context.TODO()
