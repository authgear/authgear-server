package log

import (
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

type FormatHook struct {
	MaskPatterns []MaskPattern
	Mask         string
}

func (h *FormatHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *FormatHook) Fire(entry *logrus.Entry) error {
	if len(entry.Data) > 0 {
		fields := make(logrus.Fields, len(entry.Data))
		for k, v := range entry.Data {
			if err, ok := v.(error); ok {
				// should be a safe JSON value (no need to mask)
				details := errorutil.GetSafeDetails(err)
				if len(details) > 0 {
					fields["details"] = details
				}
			}

			v, err := ensureJSON(v)
			if err != nil {
				return err
			}
			v = h.maskJSON(v)
			fields[k] = v
		}
		entry.Data = fields
	}

	entry.Message = h.maskString(entry.Message)
	return nil
}

func ensureJSON(d interface{}) (interface{}, error) {
	switch d := d.(type) {
	case int, int32, int64, float32, float64, string, bool, nil:
		return d, nil
	case error:
		return errorutil.Summary(d), nil
	default:
		// round-trip to ensure JSON
		b, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &d)
		return d, err
	}
}

func (h *FormatHook) maskJSON(json interface{}) interface{} {
	switch value := json.(type) {
	case string:
		json = h.maskString(value)
	case []interface{}:
		for i, v := range value {
			value[i] = h.maskJSON(v)
		}
	case map[string]interface{}:
		for k, v := range value {
			value[k] = h.maskJSON(v)
		}
	}
	return json
}

func (h *FormatHook) maskString(s string) string {
	for _, p := range h.MaskPatterns {
		s = p.Mask(s, h.Mask)
	}
	return s
}
