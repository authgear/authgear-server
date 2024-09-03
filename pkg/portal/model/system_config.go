package model

import (
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type SystemConfig struct {
	AuthgearClientID         string         `json:"authgearClientID"`
	AuthgearEndpoint         string         `json:"authgearEndpoint"`
	SentryDSN                string         `json:"sentryDSN,omitempty"`
	AppHostSuffix            string         `json:"appHostSuffix"`
	AvailableLanguages       []string       `json:"availableLanguages"`
	BuiltinLanguages         []string       `json:"builtinLanguages"`
	Themes                   interface{}    `json:"themes,omitempty"`
	Translations             interface{}    `json:"translations,omitempty"`
	SearchEnabled            bool           `json:"searchEnabled"`
	Web3Enabled              bool           `json:"web3Enabled"`
	AuditLogEnabled          bool           `json:"auditLogEnabled"`
	AnalyticEnabled          bool           `json:"analyticEnabled"`
	AnalyticEpoch            *timeutil.Date `json:"analyticEpoch,omitempty"`
	GitCommitHash            string         `json:"gitCommitHash,omitempty"`
	GTMContainerID           string         `json:"gtmContainerID,omitempty"`
	UIImplementation         string         `json:"uiImplementation,omitempty"`
	UISettingsImplementation string         `json:"uiSettingsImplementation,omitempty"`
}
