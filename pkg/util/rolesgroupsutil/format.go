package rolesgroupsutil

import (
	"context"
	"fmt"
	"strings"
)

var KeyReservedPrefix = "authgear:"

type FormatKey struct{}

func (_ FormatKey) CheckFormat(ctx context.Context, value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	if strings.HasPrefix(str, KeyReservedPrefix) {
		return fmt.Errorf("key cannot start with the preserved prefix: `%v`", KeyReservedPrefix)
	}

	return nil
}
