package transport

import (
	"net/http"

	graphqlgohandler "github.com/graphql-go/handler"

	"github.com/authgear/authgear-server/pkg/portal/graphql"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/httproute"
)

func ConfigureGraphQLRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "POST").
		WithPathPattern("/api/graphql")
}

type GraphQLHandler struct {
	GraphQLContext *graphql.Context
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		graphiql := &graphqlutil.GraphiQL{
			Title:    "GraphiQL: Portal - Authgear",
			IsPortal: true,
		}
		graphiql.ServeHTTP(w, r)
		return
	} else {
		// graphql-go/handler will use "query=" when it is present.
		// This causes GraphiQL unable to fetch the schema.
		q := r.URL.Query()
		q.Del("query")
		r.URL.RawQuery = q.Encode()
	}

	graphqlHandler := graphqlgohandler.New(&graphqlgohandler.Config{
		Schema:   graphql.Schema,
		Pretty:   false,
		GraphiQL: false,
	})

	ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
	graphqlHandler.ContextHandler(ctx, w, r)
}
