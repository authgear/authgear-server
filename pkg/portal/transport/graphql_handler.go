package transport

import (
	"net/http"

	graphqlgohandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGraphQLRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/api/graphql")
}

type GraphQLHandler struct {
	Config         *config.ServerConfig
	GraphQLContext *graphql.Context
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	graphqlHandler := graphqlgohandler.New(&graphqlgohandler.Config{
		Schema:   graphql.Schema,
		Pretty:   h.Config.DevMode,
		GraphiQL: h.Config.DevMode,
	})
	ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
	graphqlHandler.ContextHandler(ctx, w, r)
}
