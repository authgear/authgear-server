package graphql

import (
	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

var appResource = graphql.NewObject(graphql.ObjectConfig{
	Name:        "AppResource",
	Description: "Resource file for an app",
	Fields: graphql.Fields{
		"path": &graphql.Field{
			Type: graphql.NewNonNull(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				r := p.Source.(*model.AppResource)
				return r.Path, nil
			},
		},
		"data": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				r := p.Source.(*model.AppResource)
				for _, f := range r.Files {
					if f.Fs == r.Context.AppFs {
						return string(f.Data), nil
					}
				}
				return nil, nil
			},
		},
		"effectiveData": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				r := p.Source.(*model.AppResource)
				if r.Descriptor == configsource.SecretConfig {
					return nil, apierrors.NewForbidden("cannot access effective secrets")
				}

				var layers []resource.LayerFile
				for _, f := range r.Files {
					layers = append(layers, resource.LayerFile{
						Path: r.Path,
						Data: f.Data,
					})
				}

				// Expose raw representation of merged data in API
				merged, err := r.Descriptor.Merge(layers, map[string]interface{}{
					resource.ArgMergeRaw: true,
				})
				if err != nil {
					return nil, err
				}

				return string(merged.Data), nil
			},
		},
	},
})
