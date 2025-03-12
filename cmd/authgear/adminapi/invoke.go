package adminapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/admin/authz"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
)

type Invoker struct {
	DatabaseHandle *globaldb.Handle
	Store          *configsource.Store
	Adder          *authz.Adder
}

func (i *Invoker) FetchAdminAPIKeys(ctx context.Context, appID string) (*config.AdminAPIAuthKey, error) {
	var dbSrc *configsource.DatabaseSource
	var err error
	err = i.DatabaseHandle.ReadOnly(ctx, func(ctx context.Context) error {
		dbSrc, err = i.Store.GetDatabaseSourceByAppID(ctx, appID)
		return err
	})
	if err != nil {
		return nil, err
	}

	authgearSecretsYAML := dbSrc.Data[configsource.AuthgearSecretYAML]
	secretConfig, err := config.ParsePartialSecret(ctx, authgearSecretsYAML)
	if err != nil {
		return nil, err
	}

	authKey, ok := secretConfig.LookupData(config.AdminAPIAuthKeyKey).(*config.AdminAPIAuthKey)
	if !ok {
		return nil, fmt.Errorf("key %v not found in %v", config.AdminAPIAuthKeyKey, configsource.AuthgearSecretYAML)
	}

	return authKey, nil
}

type InvokeOptions struct {
	AppID         string
	Endpoint      string
	Host          string
	AdminAPIKey   *config.AdminAPIAuthKey
	Query         string
	OperationName string
	VariablesJSON string
}

type InvokeResult struct {
	// The body of HTTPResponse MUST NOT BE used.
	HTTPResponse   *http.Response
	HTTPBody       []byte
	DumpedResponse []byte
}

func (r *InvokeResult) Error() string {
	return string(r.DumpedResponse)
}

func (i *Invoker) Invoke(ctx context.Context, options InvokeOptions) (*InvokeResult, error) {
	u, err := url.Parse(options.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	u.Path = "/_api/admin/graphql"

	body := map[string]interface{}{
		"query": options.Query,
	}
	if options.VariablesJSON != "" {
		body["variables"] = json.RawMessage(options.VariablesJSON)
	}
	if options.OperationName != "" {
		body["operationName"] = options.OperationName
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	if options.Host != "" {
		req.Host = options.Host
	}

	req.Header.Set("Content-Type", "application/json")
	err = i.Adder.AddAuthz(config.AdminAPIAuthJWT, config.AppID(options.AppID), options.AdminAPIKey, nil, req.Header)
	if err != nil {
		return nil, err
	}

	client := utilhttputil.NewExternalClient(5 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dumpedResponse, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &InvokeResult{
		HTTPResponse:   resp,
		HTTPBody:       respBody,
		DumpedResponse: dumpedResponse,
	}

	var respJSON map[string]interface{}
	err = json.Unmarshal(respBody, &respJSON)
	if err != nil {
		return nil, errors.Join(err, result)
	}

	_, hasErrors := respJSON["errors"]
	if hasErrors {
		return nil, fmt.Errorf("GraphQL response contains errors: %w", result)
	}

	return result, nil
}
