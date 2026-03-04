package transport

import "github.com/authgear/authgear-server/pkg/util/httproute"

func ConfigureGraphQLRoute(route httproute.Route) []httproute.Route {
	route = route.WithMethods("GET", "POST")
	return []httproute.Route{
		route.WithPathPattern("/api/graphql"),
	}
}
