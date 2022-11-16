package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type SyncDenoClient struct {
	*DenoClient
}

func NewSyncDenoClient(endpoint config.DenoEndpoint, c *config.HookConfig, logger Logger) SyncDenoClient {
	return SyncDenoClient{
		&DenoClient{
			Endpoint:   string(endpoint),
			HTTPClient: httputil.NewExternalClient(c.SyncTimeout.Duration()),
			Logger:     logger,
		},
	}
}

type AsyncDenoClient struct {
	*DenoClient
}

func NewAsyncDenoClient(endpoint config.DenoEndpoint, logger Logger) AsyncDenoClient {
	return AsyncDenoClient{
		&DenoClient{
			Endpoint:   string(endpoint),
			HTTPClient: httputil.NewExternalClient(60 * time.Second),
			Logger:     logger,
		},
	}
}

type DenoClient struct {
	Endpoint   string
	HTTPClient *http.Client
	Logger     Logger
}

func (c *DenoClient) Run(ctx context.Context, snippet string, input interface{}) (interface{}, error) {
	u, err := url.JoinPath(c.Endpoint, "/run")
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(RunRequest{
		Script: snippet,
		Input:  input,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, WebHookInvalidResponse.NewWithInfo("invalid status code", apierrors.Details{
			"status_code": resp.StatusCode,
		})
	}

	var runResponse RunResponse
	err = json.NewDecoder(resp.Body).Decode(&runResponse)
	if err != nil {
		return nil, err
	}

	c.Logger.WithFields(map[string]interface{}{
		"error":  runResponse.Error,
		"output": runResponse.Output,
		"stdout": runResponse.Stdout,
		"stderr": runResponse.Stderr,
	}).Info("run deno script")

	if runResponse.Error != "" {
		return nil, DenoRunError.NewWithInfo(runResponse.Error, apierrors.Details{
			"stdout": runResponse.Stdout,
			"stderr": runResponse.Stderr,
		})
	}

	return runResponse.Output, nil
}
