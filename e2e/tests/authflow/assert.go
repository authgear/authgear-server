package tests

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authflowclient"
)

func PerformAssertion(assertion Assert, value interface{}) error {
	valueStr, ok := value.(string)
	if !ok {
		return errors.New(fmt.Sprintf("field '%s' is not valid\n", assertion.Field))
	}

	switch assertion.Op {
	case AssertOpEq:
		if valueStr != assertion.Value {
			return errors.New(fmt.Sprintf("expected '%s' to be '%s', got '%s'\n", assertion.Field, assertion.Value, valueStr))
		}
	case AssertOpNeq:
		if valueStr == assertion.Value {
			return errors.New(fmt.Sprintf("expected '%s' to not be '%s'\n", assertion.Field, assertion.Value))
		}
	case AssertOpContains:
		if !strings.Contains(valueStr, assertion.Value) {
			return errors.New(fmt.Sprintf("expected '%s' to contain '%s', got '%s'\n", assertion.Field, assertion.Value, valueStr))
		}
	}

	return nil
}

func TranslateAssertValue(flowResponse *authflowclient.FlowResponse, err error, field AssertField) (interface{}, bool) {
	var apiError *apierrors.APIError
	if err != nil {
		apiError = apierrors.AsAPIError(err)
	}

	switch field {
	case AssertFieldActionType:
		if flowResponse == nil {
			return nil, false
		}
		return string(flowResponse.Action.Type), true
	case AssertFieldErrorReason:
		if apiError == nil {
			return nil, false
		}
		return apiError.Reason, true
	default:
		return nil, false
	}
}
