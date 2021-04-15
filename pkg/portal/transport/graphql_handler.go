package transport

import (
	"context"
	"errors"
	"net/http"

	gographql "github.com/graphql-go/graphql"
	graphqlgohandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/lib/config"
	db "github.com/authgear/authgear-server/pkg/lib/infra/db/global"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGraphQLRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/api/graphql")
}

type GraphQLHandler struct {
	DevMode        config.DevMode
	GraphQLContext *graphql.Context
	Database       *db.Handle
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.Database.WithTx(func() error {
		doRollback := false
		graphqlHandler := graphqlgohandler.New(&graphqlgohandler.Config{
			Schema:   graphql.Schema,
			Pretty:   false,
			GraphiQL: bool(h.DevMode),
			ResultCallbackFn: func(ctx context.Context, params *gographql.Params, result *gographql.Result, responseBody []byte) {
				if result.HasErrors() {
					doRollback = true
				}
			},
		})

		ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
		graphqlHandler.ContextHandler(ctx, w, r)

		if doRollback {
			return errRollback
		}
		return nil
	})
	if err != nil && !errors.Is(err, errRollback) {
		panic(err)
	}
}
