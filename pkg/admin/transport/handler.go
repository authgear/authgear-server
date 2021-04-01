package transport

import (
	"context"
	"errors"
	"net/http"

	gographql "github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGraphQLRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/graphql")
}

var errRollback = errors.New("rollback transaction")

type GraphQLHandler struct {
	GraphQLContext *graphql.Context
	Database       *db.Handle
}

func (h *GraphQLHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	err := h.Database.WithTx(func() error {
		doRollback := false
		graphqlHandler := handler.New(&handler.Config{
			Schema:   graphql.Schema,
			Pretty:   false,
			GraphiQL: true,
			ResultCallbackFn: func(ctx context.Context, params *gographql.Params, result *gographql.Result, responseBody []byte) {
				if result.HasErrors() {
					doRollback = true
				}
			},
		})

		ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
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
