package graphql

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/crc32"

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
				resMgr := ctx.AppResMgrFactory.NewManagerWithAppContext(r.Context)
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
				ctx := GQLContext(p.Context)
				r := p.Source.(*model.AppResource)
				resMgr := ctx.AppResMgrFactory.NewManagerWithAppContext(r.Context)
				result, err := resMgr.ReadAppFile(r.DescriptedPath.Descriptor,
					&resource.AppFile{
						Path: r.DescriptedPath.Path,
					})
				if errors.Is(err, resource.ErrResourceNotFound) {
					return nil, nil
				} else if err != nil {
					return nil, err
				}
				// Calculate the checksum with crc32 IEEE
				checksum := crc32.ChecksumIEEE(result.([]byte))
				// Turn the 32-bit unsigned checksum into 4 bytes in big endian order.
				byteSlice := make([]byte, 4)
				byteSlice = binary.BigEndian.AppendUint32(byteSlice, checksum)

				// Encode the 4 bytes in hex format.
				checksumString := hex.EncodeToString(byteSlice)
				return checksumString, nil
			},
		},
	},
})
