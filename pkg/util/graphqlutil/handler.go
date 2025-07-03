// The file is vendored from https://github.com/graphql-go/handler/blob/v0.2.4/handler.go
//
// Notable changes:
// - Remove the type Config. You construct Handler directly.
// - Remove features that we do not use, such as pretty, graphiql, playground.
// - Remove ContentTypeFormURLEncoded, and its associated handling.
package graphqlutil

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/graphql-go/graphql"
)

const (
	ContentTypeJSON    = "application/json"
	ContentTypeGraphQL = "application/graphql"
)

var ErrMethodMustBePost = errors.New("http method must be POST")
var ErrBodyMustBeNonNil = errors.New("http request body must be non-nil")

type ResultCallbackFn func(ctx context.Context, params *graphql.Params, result *graphql.Result, responseBody []byte)

type Handler struct {
	Schema           *graphql.Schema
	ResultCallbackFn ResultCallbackFn
}

type RequestOptions struct {
	Query         string                 `json:"query" url:"query" schema:"query"`
	Variables     map[string]interface{} `json:"variables" url:"variables" schema:"variables"`
	OperationName string                 `json:"operationName" url:"operationName" schema:"operationName"`
}

// a workaround for getting`variables` as a JSON string
type requestOptionsCompatibility struct {
	Query         string `json:"query" url:"query" schema:"query"`
	Variables     string `json:"variables" url:"variables" schema:"variables"`
	OperationName string `json:"operationName" url:"operationName" schema:"operationName"`
}

func getFromForm(values url.Values) (*RequestOptions, error) {
	query := values.Get("query")
	if query != "" {
		// get variables map
		variables := make(map[string]interface{}, len(values))
		variablesStr := values.Get("variables")
		err := json.Unmarshal([]byte(variablesStr), &variables)
		if err != nil {
			return nil, err
		}

		return &RequestOptions{
			Query:         query,
			Variables:     variables,
			OperationName: values.Get("operationName"),
		}, nil
	}

	return nil, nil
}

// RequestOptions Parses a http.Request into GraphQL request options struct
func NewRequestOptions(r *http.Request) (*RequestOptions, error) {
	optionsFromURLQuery, err := getFromForm(r.URL.Query())
	if err != nil {
		return nil, err
	}
	if optionsFromURLQuery != nil {
		return optionsFromURLQuery, nil
	}

	if r.Method != http.MethodPost {
		return nil, ErrMethodMustBePost
	}

	if r.Body == nil {
		return nil, ErrBodyMustBeNonNil
	}

	// TODO: improve Content-Type handling
	contentTypeStr := r.Header.Get("Content-Type")
	contentTypeTokens := strings.Split(contentTypeStr, ";")
	contentType := contentTypeTokens[0]

	switch contentType {
	case ContentTypeGraphQL:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return &RequestOptions{
			Query: string(body),
		}, nil
	case ContentTypeJSON:
		fallthrough
	default:
		var opts RequestOptions
		body, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &opts)
		if err != nil {
			// Probably `variables` was sent as a string instead of an object.
			// So, we try to be polite and try to parse that as a JSON string
			var optsCompatible requestOptionsCompatibility
			err = json.Unmarshal(body, &optsCompatible)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal([]byte(optsCompatible.Variables), &opts.Variables)
			if err != nil {
				return nil, err
			}
		}
		return &opts, nil
	}
}

// ContextHandler provides an entrypoint into executing graphQL queries with a
// user-provided context.
func (h *Handler) ContextHandler(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get query
	opts, err := NewRequestOptions(r)

	if err != nil {
		// Use panic to handle the error.
		// The mounted middleware should take care of handling the error,
		// including writing the response and logging the error.
		panic(err)
	}

	// execute graphql query
	params := graphql.Params{
		Schema:         *h.Schema,
		RequestString:  opts.Query,
		VariableValues: opts.Variables,
		OperationName:  opts.OperationName,
		Context:        ctx,
	}
	result := graphql.Do(params)

	// use proper JSON Header
	w.Header().Add("Content-Type", "application/json; charset=utf-8")

	var buff []byte
	w.WriteHeader(http.StatusOK)
	buff, err = json.Marshal(result)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(buff)
	if err != nil {
		panic(err)
	}

	if h.ResultCallbackFn != nil {
		h.ResultCallbackFn(ctx, &params, result, buff)
	}
}
