package transport

import (
	"context"
	"errors"
	"net/http"

	gographql "github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGraphQLRoute(route httproute.Route) []httproute.Route {
	route = route.WithMethods("GET", "POST")
	return []httproute.Route{
		route.WithPathPattern("/graphql"),
		route.WithPathPattern("/_api/admin/graphql"),
	}
}

var errRollback = errors.New("rollback transaction")

type GraphQLHandler struct {
	GraphQLContext *graphql.Context
	AppDatabase    *appdb.Handle
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
		graphqlHandler := handler.New(&handler.Config{
			Schema:     graphql.Schema,
			Pretty:     false,
			GraphiQL:   false,
			Playground: false,
			ResultCallbackFn: func(ctx context.Context, params *gographql.Params, result *gographql.Result, responseBody []byte) {
				if result.HasErrors() {
					doRollback = true
				}
			},
		})

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
