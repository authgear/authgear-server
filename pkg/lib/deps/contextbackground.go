package deps

import (
	"context"
)

// Some dependencies of redis queue requires a *http.Request, which
// requires a context.Context to construct.
// Since the *http.Request is already a stub, we can also provide a stub context here.
var contextForRedisQueue = context.TODO()
