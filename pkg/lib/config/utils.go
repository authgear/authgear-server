package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

func newBool(v bool) *bool { return &v }

func newInt(v int) *int { return &v }

func NewSecretMaskLogHook(cfg *SecretConfig) logrus.Hook {
	var patterns []log.MaskPattern
	for _, item := range cfg.Secrets {
		for _, s := range item.Data.SensitiveStrings() {
			if len(s) == 0 {
				continue
			}
			patterns = append(patterns, log.NewPlainMaskPattern(s))
		}
	}

	return &log.FormatHook{
		MaskPatterns: patterns,
		Mask:         "********",
	}
}

type Date time.Time

func (date *Date) IsEmpty() bool {
	empty := Date{}
	return *date == empty
}

func (date *Date) Decode(value string) error {
	t, err := time.Parse(timeutil.LayoutISODate, value)
	if err != nil {
		return err
	}
	*date = Date(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC))
	return nil
}

func (date *Date) MarshalJSON() ([]byte, error) {
	t := time.Time(*date)
	str := fmt.Sprintf(`"%s"`, t.Format(timeutil.LayoutISODate))
	return []byte(str), nil
}
