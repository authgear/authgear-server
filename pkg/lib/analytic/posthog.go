package analytic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type PosthogCredentials struct {
	Endpoint string
	APIKey   string
}

type PosthogLogger struct{ *log.Logger }

func NewPosthogLogger(lf *log.Factory) PosthogLogger {
	return PosthogLogger{lf.New("posthog-integration")}
}

type PosthogHTTPClient struct {
	*http.Client
}

func NewPosthogHTTPClient() PosthogHTTPClient {
	return PosthogHTTPClient{
		httputil.NewExternalClient(5 * time.Second),
	}
}

type PosthogIntegration struct {
	PosthogCredentials *PosthogCredentials
	Clock              clock.Clock
	GlobalHandle       *globaldb.Handle
	GlobalDBStore      *GlobalDBStore
	AppDBHandle        *appdb.Handle
	AppDBStore         *AppDBStore
	HTTPClient         PosthogHTTPClient
	Logger             PosthogLogger
	ReadCounterStore   ReadCounterStore
}

type PosthogGroup struct {
	ProjectID         string
	MAU               int
	UserCount         int
	CollaboratorCount int
	ApplicationCount  int
	ProjectPlan       string
}

func (p *PosthogIntegration) SetGroupProperties(ctx context.Context) error {
	endpoint, err := url.Parse(p.PosthogCredentials.Endpoint)
	if err != nil {
		return err
	}

	now := p.Clock.NowUTC()

	appIDs, err := p.getAppIDs(ctx)
	if err != nil {
		return err
	}

	var groups []*PosthogGroup
	for _, appID := range appIDs {
		g, err := p.preparePosthogGroup(ctx, appID, now)
		if err != nil {
			return err
		}

		groups = append(groups, g)

		p.Logger.
			WithField("project_id", appID).
			WithField("mau", g.MAU).
			WithField("user_count", g.UserCount).
			WithField("collaborator_count", g.CollaboratorCount).
			WithField("application_count", g.ApplicationCount).
			WithField("project_plan", g.ProjectPlan).
			Info("prepared group")
	}

	events, err := p.makeEventsFromGroups(groups)
	if err != nil {
		return err
	}

	err = p.Batch(ctx, endpoint, events)
	if err != nil {
		return err
	}

	return nil
}

func (p *PosthogIntegration) SetUserProperties(ctx context.Context, portalAppID string) error {
	endpoint, err := url.Parse(p.PosthogCredentials.Endpoint)
	if err != nil {
		return err
	}

	var users []*User
	err = p.AppDBHandle.WithTx(ctx, func(ctx context.Context) error {
		var err error
		users, err = p.AppDBStore.GetAllUsers(ctx, portalAppID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	events, err := p.makeEventsFromUsers(users)
	if err != nil {
		return err
	}

	err = p.Batch(ctx, endpoint, events)
	if err != nil {
		return err
	}

	return nil
}

func (p *PosthogIntegration) getAppIDs(ctx context.Context) (appIDs []string, err error) {
	err = p.GlobalHandle.WithTx(ctx, func(ctx context.Context) error {
		appIDs, err = p.GlobalDBStore.GetAppIDs(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (p *PosthogIntegration) preparePosthogGroup(ctx context.Context, appID string, now time.Time) (*PosthogGroup, error) {
	year := now.Year()
	month := now.Month()

	mau, _, err := p.ReadCounterStore.GetMonthlyActiveUserCount(ctx, config.AppID(appID), year, int(month))
	if err != nil {
		return nil, err
	}

	var userCount int
	err = p.AppDBHandle.WithTx(ctx, func(ctx context.Context) error {
		count, err := p.AppDBStore.GetUserCountBeforeTime(ctx, appID, &now)
		if err != nil {
			return err
		}
		userCount = count
		return nil
	})
	if err != nil {
		return nil, err
	}

	var collaboratorCount int
	err = p.GlobalHandle.WithTx(ctx, func(ctx context.Context) error {
		count, err := p.GlobalDBStore.GetCollaboratorCount(ctx, appID)
		if err != nil {
			return err
		}
		collaboratorCount = count
		return nil
	})
	if err != nil {
		return nil, err
	}

	var appConfigSource *AppConfigSource
	err = p.GlobalHandle.WithTx(ctx, func(ctx context.Context) error {
		s, err := p.GlobalDBStore.GetAppConfigSource(ctx, appID)
		if err != nil {
			return err
		}
		appConfigSource = s
		return nil
	})
	if err != nil {
		return nil, err
	}

	authgearYAMLBytes := appConfigSource.Data["authgear.yaml"]

	m := make(map[string]interface{})
	err = yaml.Unmarshal(authgearYAMLBytes, &m)
	if err != nil {
		return nil, err
	}

	var applicationCount int
	if oauthConfig, ok := m["oauth"].(map[string]interface{}); ok {
		if clients, ok := oauthConfig["clients"].([]interface{}); ok {
			applicationCount = len(clients)
		}
	}

	g := &PosthogGroup{
		ProjectID:         appID,
		MAU:               mau,
		UserCount:         userCount,
		CollaboratorCount: collaboratorCount,
		ApplicationCount:  applicationCount,
		ProjectPlan:       appConfigSource.PlanName,
	}

	return g, nil
}

func (p *PosthogIntegration) makeEventsFromGroups(groups []*PosthogGroup) ([]json.RawMessage, error) {
	var events []json.RawMessage

	for _, g := range groups {
		group_set := map[string]interface{}{
			"mau":                g.MAU,
			"user_count":         g.UserCount,
			"collaborator_count": g.CollaboratorCount,
			"application_count":  g.ApplicationCount,
		}
		if g.ProjectPlan != "" {
			group_set["project_plan"] = g.ProjectPlan
		}

		event := map[string]interface{}{
			"event":       "$groupidentify",
			"distinct_id": "groups_setup_id",
			"properties": map[string]interface{}{
				"$geoip_disable": true,
				"$group_type":    "project",
				"$group_key":     g.ProjectID,
				"$group_set":     group_set,
			},
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			return nil, err
		}

		events = append(events, json.RawMessage(eventBytes))
	}

	return events, nil
}

func (p *PosthogIntegration) makeEventsFromUsers(users []*User) ([]json.RawMessage, error) {
	var events []json.RawMessage

	for _, u := range users {
		set := map[string]interface{}{}
		if u.Email != "" {
			set["email"] = u.Email
		}

		event := map[string]interface{}{
			"event":       "$identify",
			"distinct_id": u.ID,
			"properties": map[string]interface{}{
				"$set":           set,
				"$geoip_disable": true,
			},
		}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			return nil, err
		}

		events = append(events, json.RawMessage(eventBytes))
	}

	return events, nil
}

type PosthogBatchRequest struct {
	APIKey string            `json:"api_key"`
	Batch  []json.RawMessage `json:"batch,omitempty"`
}

func (p *PosthogIntegration) Batch(ctx context.Context, endpoint *url.URL, events []json.RawMessage) error {
	u := *endpoint
	u.Path = "/batch"

	// The hard limit is 20MB.
	// Here we make an assumption that the size of 1000 events will not exceed the limit.

	var chunks [][]json.RawMessage
	chunkSize := 1000
	for i, chunkNum := 0, 0; i < len(events); i, chunkNum = i+chunkSize, chunkNum+1 {
		start := i
		end := i + chunkSize
		if end > len(events) {
			end = len(events)
		}

		chunk := events[start:end]
		chunks = append(chunks, chunk)
	}

	for _, chunk := range chunks {
		if len(chunk) <= 0 {
			p.Logger.WithField("batch_size", len(chunk)).Info("skipped an empty batch")
			continue
		}

		body := PosthogBatchRequest{
			APIKey: p.PosthogCredentials.APIKey,
			Batch:  chunk,
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}

		r, err := http.NewRequestWithContext(ctx, "POST", u.String(), bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}

		r.Header.Set("Content-Type", "application/json")

		resp, err := p.HTTPClient.Do(r)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("failed to upload to posthog: %v", string(respBody))
		}

		p.Logger.WithField("batch_size", len(chunk)).Info("uploaded a batch to posthog")
	}

	return nil
}
