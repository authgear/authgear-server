package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/otelauthgear"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/otelutil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

//go:generate go tool mockgen -source=deno_client.go -destination=deno_client_mock_test.go -package hook

var DenoClientLogger = slogutil.NewLogger("deno-client")

type DenoClient interface {
	Run(ctx context.Context, script string, input interface{}) (out interface{}, err error)
}

type SyncDenoClient interface {
	DenoClient
}

func NewSyncDenoClient(endpoint config.DenoEndpoint, c *config.HookConfig) SyncDenoClient {
	return &DenoClientImpl{
		Endpoint:   string(endpoint),
		HTTPClient: httputil.NewExternalClient(c.SyncTimeout.Duration()),
	}
}

type AsyncDenoClient interface {
	DenoClient
}

func NewAsyncDenoClient(endpoint config.DenoEndpoint) AsyncDenoClient {
	return &DenoClientImpl{
		Endpoint:   string(endpoint),
		HTTPClient: httputil.NewExternalClient(60 * time.Second),
	}
}

type DenoClientImpl struct {
	Endpoint   string
	HTTPClient *http.Client
}

func (c *DenoClientImpl) Run(ctx context.Context, snippet string, input interface{}) (interface{}, error) {
	logger := DenoClientLogger.GetLogger(ctx)

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
		return nil, HookInvalidResponse.NewWithInfo("invalid status code", apierrors.Details{
			"status_code": resp.StatusCode,
		})
	}

	var runResponse RunResponse
	err = json.NewDecoder(resp.Body).Decode(&runResponse)
	if err != nil {
		return nil, err
	}

	logger.With(
		slog.String("response_error", runResponse.Error),
		slog.String("error_code", string(runResponse.ErrorCode)),
		slog.Any("output", runResponse.Output),
		slog.Any("stdout", runResponse.Stdout),
		slog.Any("stderr", runResponse.Stderr),
	).Info(ctx, "run deno script")

	if runResponse.Error != "" {
		otelutil.IntCounterAddOne(ctx,
			otelauthgear.CounterDenoRunCount,
			otelauthgear.WithStatusError(),
			otelauthgear.WithDenoErrorCode(string(runResponse.ErrorCode)),
		)

		return nil, DenoRunError.NewWithInfo(runResponse.Error, apierrors.Details{
			"stdout": runResponse.Stdout,
			"stderr": runResponse.Stderr,
		})
	} else {
		otelutil.IntCounterAddOne(ctx,
			otelauthgear.CounterDenoRunCount,
			otelauthgear.WithStatusOk())
	}

	return runResponse.Output, nil
}

func (c *DenoClientImpl) Check(ctx context.Context, snippet string) error {
	u, err := url.JoinPath(c.Endpoint, "/check")
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(CheckRequest{
		Script: snippet,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u, &buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return HookInvalidResponse.NewWithInfo("invalid status code", apierrors.Details{
			"status_code": resp.StatusCode,
		})
	}

	var checkResponse CheckResponse
	err = json.NewDecoder(resp.Body).Decode(&checkResponse)
	if err != nil {
		return err
	}

	if checkResponse.Stderr != "" {
		return DenoCheckError.New(checkResponse.Stderr)
	}

	return nil
}
