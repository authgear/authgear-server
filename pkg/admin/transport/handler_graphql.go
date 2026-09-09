package transport

import (
	"context"
	"errors"
	"net/http"

	gographql "github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ConfigureGraphQLRoute(route httproute.Route) []httproute.Route {
	route = route.WithMethods("GET", "POST")
	return []httproute.Route{
		route.WithPathPattern("/graphql"),
		route.WithPathPattern("/_api/admin/graphql"),
	}
}

var errRollback = errors.New("rollback transaction")

type MutationRateLimiter interface {
	AllowN(ctx context.Context, spec ratelimit.BucketSpec, n int) (*ratelimit.FailedReservation, error)
}

type GraphQLHandler struct {
	GraphQLContext        *graphql.Context
	AppDatabase           *appdb.Handle
	RateLimiter           MutationRateLimiter
	RemoteIP              httputil.RemoteIP
	AdminAPIFeatureConfig *config.AdminAPIFeatureConfig
}

// checkMutationRateLimit takes mutationFieldCount tokens from each Admin API
// mutation bucket, so a document carrying several mutation fields costs one
// token per field rather than one per request.
//
// Per-IP is checked before per-project so the tighter, caller-scoped bucket
// short-circuits first, matching the DCR register handler's ordering. A
// rejected request therefore charges only the bucket that rejected it.
func (h *GraphQLHandler) checkMutationRateLimit(ctx context.Context, mutationFieldCount int) error {
	// Queries cost nothing, and never reach Redis.
	if mutationFieldCount <= 0 {
		return nil
	}

	rateLimits := h.AdminAPIFeatureConfig.GetRateLimits().GetMutation().GetAll()
	if rateLimits == nil {
		return nil
	}

	specs := []ratelimit.BucketSpec{
		NewBucketSpecAdminAPIMutationAllPerIP(rateLimits, string(h.RemoteIP)),
		NewBucketSpecAdminAPIMutationAllPerProject(rateLimits),
	}
	for _, spec := range specs {
		failed, err := h.RateLimiter.AllowN(ctx, spec, mutationFieldCount)
		if err != nil {
			return err
		}
		if failed != nil {
			return failed.Error()
		}
	}

	return nil
}

func (h *GraphQLHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		graphiql := &graphqlutil.GraphiQL{
			Title:    "GraphiQL: Admin API - Authgear",
			IsPortal: r.Header.Get("X-Authgear-Portal-Is-Proxied") == "true",
		}
		graphiql.ServeHTTP(rw, r)
		return
	} else {
		// graphql-go/handler will use "query=" when it is present.
		// This causes GraphiQL unable to fetch the schema.
		q := r.URL.Query()
		q.Del("query")
		r.URL.RawQuery = q.Encode()
	}

	ctx := r.Context()
	err := h.AppDatabase.WithTx(ctx, func(ctx context.Context) error {
		doRollback := false
		graphqlHandler := &graphqlutil.Handler{
			Schema: graphql.Schema,
			BeforeExecuteFn: func(ctx context.Context, mutationFieldCount int) error {
				return h.checkMutationRateLimit(ctx, mutationFieldCount)
			},
			ResultCallbackFn: func(ctx context.Context, params *gographql.Params, result *gographql.Result, responseBody []byte) {
				if result.HasErrors() {
					doRollback = true
				}
			},
		}

		ctx = graphql.WithContext(ctx, h.GraphQLContext)
		graphqlHandler.ContextHandler(ctx, rw, r)

		if doRollback {
			return errRollback
		}
		return nil
	})

	if err != nil && !errors.Is(err, errRollback) {
		panic(err)
	}
}
