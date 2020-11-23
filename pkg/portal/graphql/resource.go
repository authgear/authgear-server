package graphql

import (
	"encoding/base64"
	"errors"

	"github.com/graphql-go/graphql"

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
				return r.DescriptedPath.Path, nil
			},
		},
		"languageTag": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				r := p.Source.(*model.AppResource)
				match, ok := r.DescriptedPath.Descriptor.MatchResource(r.DescriptedPath.Path)
				if !ok || match.LanguageTag == "" {
					return nil, nil
				}
				return match.LanguageTag, nil
			},
		},
		"data": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := GQLContext(p.Context)
				r := p.Source.(*model.AppResource)
				result, err := r.Context.Resources.Read(r.DescriptedPath.Descriptor, resource.AppFile{
					Path:              r.DescriptedPath.Path,
					AllowedSecretKeys: ctx.SecretKeyAllowlist,
				})
				if errors.Is(err, resource.ErrResourceNotFound) {
					return nil, nil
				} else if err != nil {
					return nil, err
				}

				return base64.StdEncoding.EncodeToString(result.([]byte)), nil
			},
		},
		"effectiveData": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				r := p.Source.(*model.AppResource)
				result, err := r.Context.Resources.Read(r.DescriptedPath.Descriptor, resource.EffectiveFile{
					Path:       r.DescriptedPath.Path,
					DefaultTag: r.Context.Config.AppConfig.Localization.FallbackLanguage,
				})
				if errors.Is(err, resource.ErrResourceNotFound) {
					return nil, nil
				} else if err != nil {
					return nil, err
				}
				return base64.StdEncoding.EncodeToString(result.([]byte)), nil
			},
		},
	},
})
