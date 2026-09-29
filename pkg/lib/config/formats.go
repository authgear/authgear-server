package config

import (
	"context"
	"fmt"
	"net"
	"strconv"

	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"

	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	jsonschemaformat.DefaultChecker["phone"] = FormatPhone{}
	jsonschemaformat.DefaultChecker["x_host_port"] = FormatHostPort{}
}

// FormatHostPort checks that the input is a "host:port" address.
type FormatHostPort struct{}

func (f FormatHostPort) CheckFormat(ctx context.Context, value any) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	host, port, err := net.SplitHostPort(str)
	if err != nil {
		return err
	}
	if host == "" {
		return fmt.Errorf("expect non-empty host")
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 1 || p > 65535 {
		return fmt.Errorf("expect port in range 1-65535")
	}
	return nil
}

// FormatPhone checks if input is a phone number in E.164 format.
// If the input is not a string, it is not an error.
// To enforce string, use other JSON schema constructs.
// This design allows this format to validate optional phone number.
type FormatPhone struct{}

func (f FormatPhone) CheckFormat(ctx context.Context, value any) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	appCtx, ok := GetAppContext(ctx)
	if ok {
		cfg := appCtx.Config.AppConfig.UI.PhoneInput.Validation
		switch cfg.Implementation {
		case PhoneInputValidationImplementationLibphonenumber:
			switch cfg.Libphonenumber.ValidationMethod {
			case LibphonenumberValidationMethodIsPossibleNumber:
				_, err := phone.Parse_IsPossibleNumber_ReturnE164(str)
				if err != nil {
					return err
				}
			case LibphonenumberValidationMethodIsValidNumber:
				err := phone.Require_IsPossibleNumber_IsValidNumber_UserInputInE164(str)
				if err != nil {
					return err
				}
			default:
				panic(fmt.Errorf("unknown validation method: %s", cfg.Libphonenumber.ValidationMethod))

			}
		default:
			panic(fmt.Errorf("unknown validation implementation: %s", cfg.Implementation))
		}
	} else {
		// If AppContext is not available, validate with strictest rule
		err := phone.Require_IsPossibleNumber_IsValidNumber_UserInputInE164(str)
		if err != nil {
			return err
		}
	}
	return nil
}
