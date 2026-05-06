package panicutil

import (
	"fmt"
)

func MakeError(val any) error {
	if val == nil {
		return nil
	}
	if err, ok := val.(error); ok {
		return err
	}
	return fmt.Errorf("%+v", val)
}
