package testrunner

import (
	"encoding/json"
	"fmt"

	authflowclient "github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func MatchOutput(output Output, flowResult *authflowclient.FlowResponse, flowError error) (resultViolations []MatchViolation, errorViolations []MatchViolation, err error) {
	if output.Result != "" {
		if flowResult == nil {
			return nil, nil, fmt.Errorf("expected flow result, got nil")
		}

		flowResultJSON, err := json.Marshal(flowResult)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal flow result: %w", err)
		}

		resultViolations, err = MatchJSON(string(flowResultJSON), output.Result)
		if err != nil {
			return nil, nil, err
		}
	}

	if output.Error != "" {
		if flowError == nil {
			return nil, nil, fmt.Errorf("expected flow error, got nil")
		}

		apiError := apierrors.AsAPIError(flowError)
		apiErrorJSON, err := json.Marshal(&apiError)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal API error: %w", err)
		}

		errorViolations, err = MatchJSON(string(apiErrorJSON), output.Error)
		if err != nil {
			return nil, nil, err
		}
	}

	return errorViolations, resultViolations, nil
}
