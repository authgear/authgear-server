package admin

import (
	"net/http"

	"github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/admin/graphql"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGraphQLRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/graphql")
}

var graphqlHandler = handler.New(&handler.Config{
	Schema:     graphql.Schema,
	Pretty:     true,
	GraphiQL:   true,
	Playground: false,
})

type GraphQLHandler struct {
	GraphQLContext *graphql.Context
}

func (h *GraphQLHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
	graphqlHandler.ContextHandler(ctx, rw, r)
}
