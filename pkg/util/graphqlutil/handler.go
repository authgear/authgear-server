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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

const (
	ContentTypeJSON    = "application/json"
	ContentTypeGraphQL = "application/graphql"
)

var ErrMethodMustBePost = errors.New("http method must be POST")
var ErrBodyMustBeNonNil = errors.New("http request body must be non-nil")

const maxMutationFieldsPerRequest = 5

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

	if err := validateMutationFieldCount(opts.Query, opts.OperationName, maxMutationFieldsPerRequest); err != nil {
		writeErrorResponse(ctx, w, err)
		return
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

func writeErrorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	apiError := apierrors.AsAPIErrorWithContext(ctx, err)
	logger := logger.GetLogger(ctx)
	logger.WithError(err).With(
		slog.Int("status_code", apiError.Code),
		slog.String("error_name", string(apiError.Kind.Name)),
		slog.String("error_reason", apiError.Kind.Reason),
		slog.String("error_message", apiError.Message),
	).Warn(ctx, "rejecting GraphQL request before execution")

	resp := &api.Response{Error: err}
	bodyBytes, encodeErr := resp.EncodeToJSON(ctx)
	if encodeErr != nil {
		panic(encodeErr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiError.Code)
	_, writeErr := w.Write(bodyBytes)
	if writeErr != nil {
		panic(writeErr)
	}
}

func validateMutationFieldCount(query string, operationName string, limit int) error {
	doc, err := parser.Parse(parser.ParseParams{Source: query})
	if err != nil {
		return err
	}

	operation, err := findOperation(doc, operationName)
	if err != nil || operation == nil {
		return err
	}
	if operation.Operation != ast.OperationTypeMutation {
		return nil
	}

	fragments := make(map[string]*ast.FragmentDefinition)
	for _, def := range doc.Definitions {
		fragment, ok := def.(*ast.FragmentDefinition)
		if !ok || fragment.Name == nil {
			continue
		}
		fragments[fragment.Name.Value] = fragment
	}

	count, err := countTopLevelFields(operation.SelectionSet, fragments, map[string]bool{})
	if err != nil {
		return err
	}
	if count > limit {
		return apierrors.NewBadRequest(
			fmt.Sprintf("too many mutation fields in one request: got %d, limit is %d", count, limit),
		)
	}
	return nil
}

func findOperation(doc *ast.Document, operationName string) (*ast.OperationDefinition, error) {
	var operations []*ast.OperationDefinition
	for _, def := range doc.Definitions {
		if op, ok := def.(*ast.OperationDefinition); ok {
			operations = append(operations, op)
		}
	}

	if operationName == "" {
		if len(operations) == 1 {
			return operations[0], nil
		}
		return nil, nil
	}

	for _, op := range operations {
		if op.Name != nil && op.Name.Value == operationName {
			return op, nil
		}
	}
	return nil, nil
}

func countTopLevelFields(
	selectionSet *ast.SelectionSet,
	fragments map[string]*ast.FragmentDefinition,
	visiting map[string]bool,
) (int, error) {
	if selectionSet == nil {
		return 0, nil
	}

	count := 0
	for _, selection := range selectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			count++
		case *ast.InlineFragment:
			n, err := countTopLevelFields(sel.SelectionSet, fragments, visiting)
			if err != nil {
				return 0, err
			}
			count += n
		case *ast.FragmentSpread:
			if sel.Name == nil {
				continue
			}
			name := sel.Name.Value
			if visiting[name] {
				return 0, apierrors.NewBadRequest(
					fmt.Sprintf("graphql fragment cycle detected at %q", name),
				)
			}
			fragment := fragments[name]
			if fragment == nil {
				continue
			}
			visiting[name] = true
			n, err := countTopLevelFields(fragment.SelectionSet, fragments, visiting)
			delete(visiting, name)
			if err != nil {
				return 0, err
			}
			count += n
		}
	}

	return count, nil
}
