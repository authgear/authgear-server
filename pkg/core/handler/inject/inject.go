package inject

import (
	"context"
	"reflect"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

func DefaultInject(
	h handler.Handler,
	dependencyGraph handler.ProviderGraph,
	ctx context.Context,
	configuration config.TenantConfiguration,
) {
	injectDependency(h, dependencyGraph, ctx, configuration)
	injectAuthorizationPolicy(h)
}

func injectDependency(
	h handler.Handler,
	dependencyGraph handler.ProviderGraph,
	ctx context.Context,
	configuration config.TenantConfiguration,
) {
	t := reflect.TypeOf(h).Elem()
	v := reflect.ValueOf(h).Elem()

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		dependencyName := t.Field(i).Tag.Get("dependency")
		if dependencyName == "" {
			continue
		}

		field := v.Field(i)
		dependency := dependencyGraph.Provide(dependencyName, ctx, configuration)
		field.Set(reflect.ValueOf(dependency))
	}
}

func injectAuthorizationPolicy(h handler.Handler) {
	t := reflect.TypeOf(h).Elem()
	v := reflect.ValueOf(h).Elem()

	policyType := reflect.TypeOf((*authz.Policy)(nil)).Elem()

	numField := t.NumField()
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.Type == policyType {
			allow := field.Tag.Get("allow")
			deny := field.Tag.Get("deny")
			policy := authz.NewDefaultPolicy(
				strings.Split(allow, ","),
				strings.Split(deny, ","),
			)
			v.Field(i).Set(reflect.ValueOf(policy))
		}
	}
}
