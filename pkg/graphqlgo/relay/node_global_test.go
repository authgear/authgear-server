package relay_test

import (
	context0 "context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"

	"github.com/authgear/authgear-server/pkg/graphqlgo/relay"
)

type photo2 struct {
	PhotoId int `json:"photoId"`
	Width   int `json:"width"`
}

var globalIDTestUserData = map[string]*user{
	"1": &user{1, "John Doe"},
	"2": &user{2, "Jane Smith"},
}
var globalIDTestPhotoData = map[string]*photo2{
	"1": &photo2{1, 300},
	"2": &photo2{2, 400},
}

// declare types first, define later in init()
// because they all depend on nodeTestDef
var globalIDTestUserType *graphql.Object
var globalIDTestPhotoType *graphql.Object

var globalIDTestDef = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
	IDFetcher: func(globalID string, info graphql.ResolveInfo, ctx context0.Context) (any, error) {
		resolvedGlobalID := relay.FromGlobalID(globalID)
		if resolvedGlobalID == nil {
			return nil, errors.New("Unknown node id")
		}

		switch resolvedGlobalID.Type {
		case "User":
			return globalIDTestUserData[resolvedGlobalID.ID], nil
		case "Photo":
			return globalIDTestPhotoData[resolvedGlobalID.ID], nil
		default:
			return nil, errors.New("Unknown node type")
		}
	},
	TypeResolve: func(p graphql.ResolveTypeParams) *graphql.Object {
		switch p.Value.(type) {
		case *user:
			return globalIDTestUserType
		case *photo2:
			return globalIDTestPhotoType
		default:
			panic(fmt.Sprintf("Unknown object type `%v`", p.Value))
		}
	},
})
var globalIDTestQueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"node": globalIDTestDef.NodeField,
		"allObjects": &graphql.Field{
			Type: graphql.NewList(globalIDTestDef.NodeInterface),
			Resolve: func(p graphql.ResolveParams) (any, error) {
				return []any{
					globalIDTestUserData["1"],
					globalIDTestUserData["2"],
					globalIDTestPhotoData["1"],
					globalIDTestPhotoData["2"],
				}, nil
			},
		},
	},
})

// becareful not to define schema here, since globalIDTestUserType and globalIDTestPhotoType wouldn't be defined till init()
var globalIDTestSchema graphql.Schema

func init() {
	globalIDTestUserType = graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("User", nil),
			"name": &graphql.Field{
				Type: graphql.String,
			},
		},
		Interfaces: []*graphql.Interface{globalIDTestDef.NodeInterface},
	})
	photoIDFetcher := func(obj any, info graphql.ResolveInfo, ctx context0.Context) (string, error) {
		switch obj := obj.(type) {
		case *photo2:
			return fmt.Sprintf("%v", obj.PhotoId), nil
		}
		return "", errors.New("Not a photo")
	}
	globalIDTestPhotoType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Photo",
		Fields: graphql.Fields{
			"id": relay.GlobalIDField("Photo", photoIDFetcher),
			"width": &graphql.Field{
				Type: graphql.Int,
			},
		},
		Interfaces: []*graphql.Interface{globalIDTestDef.NodeInterface},
	})

	globalIDTestSchema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query: globalIDTestQueryType,
		Types: []graphql.Type{globalIDTestUserType, globalIDTestPhotoType},
	})
}

func TestGlobalIDFields_GivesDifferentIDs(t *testing.T) {
	query := `{
      allObjects {
        id
      }
    }`
	expected := &graphql.Result{
		Data: map[string]any{
			"allObjects": []any{
				map[string]any{
					"id": "VXNlcjox",
				},
				map[string]any{
					"id": "VXNlcjoy",
				},
				map[string]any{
					"id": "UGhvdG86MQ",
				},
				map[string]any{
					"id": "UGhvdG86Mg",
				},
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        globalIDTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}

func TestGlobalIDFields_RefetchesTheIDs(t *testing.T) {
	query := `{
      user: node(id: "VXNlcjox") {
        id
        ... on User {
          name
        }
      },
      photo: node(id: "UGhvdG86MQ") {
        id
        ... on Photo {
          width
        }
      }
    }`
	expected := &graphql.Result{
		Data: map[string]any{
			"user": map[string]any{
				"id":   "VXNlcjox",
				"name": "John Doe",
			},
			"photo": map[string]any{
				"id":    "UGhvdG86MQ",
				"width": 300,
			},
		},
	}
	result := graphql.Do(graphql.Params{
		Schema:        globalIDTestSchema,
		RequestString: query,
	})
	//nolint:govet // Vendored code, do not bother to fix.
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
