package config

import (
	"context"
	"fmt"

	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"

	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	jsonschemaformat.DefaultChecker["phone"] = FormatPhone{}
}

// FormatPhone checks if input is a phone number in E.164 format.
// If the input is not a string, it is not an error.
// To enforce string, use other JSON schema constructs.
// This design allows this format to validate optional phone number.
type FormatPhone struct{}

func (f FormatPhone) CheckFormat(ctx context.Context, value interface{}) error {
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
