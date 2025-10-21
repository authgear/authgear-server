package testrunner

import (
	"encoding/json"
	"fmt"

	authflowclient "github.com/authgear/authgear-server/e2e/pkg/e2eclient"
	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

func MatchAuthflowOutput(output Output, flowResult *authflowclient.FlowResponse, flowError error) (resultViolations []MatchViolation, errorViolations []MatchViolation, err error) {
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

func MatchUserImportOutput(output AdminAPIUserImportOutput, userImportResult *authflowclient.UserImportResponseResult, userImportError error) (resultViolations []MatchViolation, errorViolations []MatchViolation, err error) {
	if output.Result != "" {
		if userImportResult == nil {
			return nil, nil, fmt.Errorf("expected user import result, got nil")
		}

		userImportResultJSON, err := json.Marshal(userImportResult)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal user import result: %w", err)
		}

		resultViolations, err = MatchJSON(string(userImportResultJSON), output.Result)
		if err != nil {
			return nil, nil, err
		}
	}

	if output.Error != "" {
		if userImportError == nil {
			return nil, nil, fmt.Errorf("expected user import error, got nil")
		}

		apiError := apierrors.AsAPIError(userImportError)
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

func MatchAdminAPIOutput(output AdminAPIOutput, resp *authflowclient.GraphQLResponse) (violations []MatchViolation, err error) {
	if output.Result != "" {
		if resp == nil {
			return nil, fmt.Errorf("expected admin api result, got nil")
		}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal admin api result: %w", err)
		}
		violations, err = MatchJSON(string(respJSON), output.Result)
		if err != nil {
			return nil, err
		}
	}
	return violations, nil
}
