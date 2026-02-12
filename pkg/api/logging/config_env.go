package logging

import (
	"fmt"
	"slices"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

const (
	LogHandlerConsole = "console"
	LogHandlerOTLP    = "otlp"
)

var ALLOWED_LOG_HANDLERS = []string{
	LogHandlerConsole,
	LogHandlerOTLP,
}

type LogEnvironmentConfig struct {
	Level        string      `envconfig:"LEVEL" default:"warn"`
	Handlers     LogHandlers `envconfig:"HANDLERS" default:"console"`
	ConsoleLevel string      `envconfig:"HANDLER_CONSOLE_LEVEL"`
	OTLPLevel    string      `envconfig:"HANDLER_OTLP_LEVEL"`
	OTLPEndpoint string      `envconfig:"HANDLER_OTLP_ENDPOINT"`
}

func LoadConfig() *LogEnvironmentConfig {
	cfg := &LogEnvironmentConfig{}
	_ = envconfig.Process("LOG", cfg)
	return cfg
}

type LogHandlers []string

func (s LogHandlers) List() []string {
	return []string(s)
}

func (s *LogHandlers) UnmarshalText(text []byte) error {
	str := string(text)

	parts := strings.Split(str, ",")
	var handlers []string
	for _, p := range parts {
		h := strings.TrimSpace(p)
		if h == "" {
			continue
		}
		if !slices.Contains(ALLOWED_LOG_HANDLERS, h) {
			panic(fmt.Errorf("invalid log handler: %s", h))
		}
		handlers = append(handlers, h)
	}
	*s = handlers
	return nil
}
