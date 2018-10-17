package pq

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/logging"
)

type tenantConfigurationValue struct {
	TenantConfiguration config.TenantConfiguration
	Valid               bool
}

func (v tenantConfigurationValue) Value() (driver.Value, error) {
	if !v.Valid {
		return nil, nil
	}

	b := bytes.Buffer{}
	if err := json.NewEncoder(&b).Encode(v.TenantConfiguration); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (v *tenantConfigurationValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		logger := logging.LoggerEntry("gateway")
		logger.Errorf("Unsupported Scan pair: %T -> %T", value, v.TenantConfiguration)
	}

	err := json.Unmarshal(b, &v.TenantConfiguration)
	if err == nil {
		v.Valid = true
	}
	return err
}
