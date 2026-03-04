package transport

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/portal/superadmin/graphql"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type GraphQLHandler struct {
	GraphQLContext *graphql.Context
}

func (h *GraphQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		graphiql := &graphqlutil.GraphiQL{Title: "GraphiQL: Superadmin API - Authgear"}
		graphiql.ServeHTTP(w, r)
		return
	}
	q := r.URL.Query()
	q.Del("query")
	r.URL.RawQuery = q.Encode()

	graphqlHandler := &graphqlutil.Handler{Schema: graphql.Schema}
	ctx := graphql.WithContext(r.Context(), h.GraphQLContext)
	graphqlHandler.ContextHandler(ctx, w, r)
}
