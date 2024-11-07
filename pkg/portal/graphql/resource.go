package graphql

import (
	"encoding/base64"
	"errors"

	"github.com/graphql-go/graphql"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/checksum"
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
				ctx := p.Context
				gqlCtx := GQLContext(ctx)
				r := p.Source.(*model.AppResource)
				resMgr := gqlCtx.AppResMgrFactory.NewManagerWithAppContext(r.Context)
				result, err := resMgr.ReadAppFile(r.DescriptedPath.Descriptor,
					&resource.AppFile{
						Path: r.DescriptedPath.Path,
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
					Path: r.DescriptedPath.Path,
				})
				if errors.Is(err, resource.ErrResourceNotFound) {
					return nil, nil
				} else if err != nil {
					return nil, err
				}
				return base64.StdEncoding.EncodeToString(result.([]byte)), nil
			},
		},
		"checksum": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context
				gqlCtx := GQLContext(ctx)
				r := p.Source.(*model.AppResource)
				resMgr := gqlCtx.AppResMgrFactory.NewManagerWithAppContext(r.Context)
				result, err := resMgr.ReadAppFile(r.DescriptedPath.Descriptor,
					&resource.AppFile{
						Path: r.DescriptedPath.Path,
					})
				if errors.Is(err, resource.ErrResourceNotFound) {
					return nil, nil
				} else if err != nil {
					return nil, err
				}

				return checksum.CRC32IEEEInHex(result.([]byte)), nil
			},
			Description: "The checksum of the resource file. It is an opaque string that will be used to detect conflict.",
		},
	},
})
