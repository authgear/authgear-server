package analytic

import (
	"bytes"
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

type PosthogIntegration struct {
	PosthogCredentials *PosthogCredentials
	Clock              clock.Clock
	GlobalHandle       *globaldb.Handle
	GlobalDBStore      *GlobalDBStore
	AppDBHandle        *appdb.Handle
	AppDBStore         *AppDBStore
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

func (p *PosthogIntegration) SetGroupProperties() error {
	endpoint, err := url.Parse(p.PosthogCredentials.Endpoint)
	if err != nil {
		return err
	}

	now := p.Clock.NowUTC()

	appIDs, err := p.getAppIDs()
	if err != nil {
		return err
	}

	var groups []*PosthogGroup
	for _, appID := range appIDs {
		g, err := p.preparePosthogGroup(appID, now)
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

	err = p.uploadToPosthog(endpoint, groups)
	if err != nil {
		return err
	}

	return nil
}

func (p *PosthogIntegration) getAppIDs() (appIDs []string, err error) {
	err = p.GlobalHandle.WithTx(func() error {
		appIDs, err = p.GlobalDBStore.GetAppIDs()
		if err != nil {
			return err
		}
		return nil
	})
	return
}

func (p *PosthogIntegration) preparePosthogGroup(appID string, now time.Time) (*PosthogGroup, error) {
	year := now.Year()
	month := now.Month()

	mau, _, err := p.ReadCounterStore.GetMonthlyActiveUserCount(config.AppID(appID), year, int(month))
	if err != nil {
		return nil, err
	}

	var userCount int
	err = p.AppDBHandle.WithTx(func() error {
		count, err := p.AppDBStore.GetUserCountBeforeTime(appID, &now)
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
	err = p.GlobalHandle.WithTx(func() error {
		count, err := p.GlobalDBStore.GetCollaboratorCount(appID)
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
	err = p.GlobalHandle.WithTx(func() error {
		s, err := p.GlobalDBStore.GetAppConfigSource(appID)
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

func (p *PosthogIntegration) uploadToPosthog(endpoint *url.URL, groups []*PosthogGroup) error {
	u := *endpoint
	u.Path = "/capture"

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

		body := map[string]interface{}{
			"api_key":     p.PosthogCredentials.APIKey,
			"event":       "$groupidentify",
			"distinct_id": "groups_setup_id",
			"properties": map[string]interface{}{
				"$group_type": "project",
				"$group_key":  g.ProjectID,
				"$group_set":  group_set,
			},
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}

		r, err := http.NewRequest("POST", u.String(), bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}

		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
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

		p.Logger.
			WithField("project_id", g.ProjectID).
			Info("uploaded to posthog")
	}

	return nil
}
