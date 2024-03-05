package userimport

import (
	"context"
	"encoding/json"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

type UserImportService struct{}

func (s *UserImportService) ImportRecords(ctx context.Context, request *Request) (*Summary, []Detail) {
	summary := Summary{}

	options := &Options{
		Upsert:     request.Upsert,
		Identifier: request.Identifier,
	}
	var details []Detail

	for idx, rawMessage := range request.Records {
		summary.Total += 1

		detail := Detail{
			Index:  idx,
			Record: rawMessage,
		}

		hasDetail := false
		outcome, warnings, err := s.ImportRecord(ctx, options, rawMessage)

		if len(warnings) > 0 {
			detail.Warnings = warnings
			hasDetail = true
		}
		if err != nil {
			detail.Errors = []*apierrors.APIError{apierrors.AsAPIError(err)}
			hasDetail = true
		}

		if hasDetail {
			details = append(details, detail)
		}

		switch outcome {
		case OutcomeInserted:
			summary.Inserted += 1
		case OutcomeUpdated:
			summary.Updated += 1
		case OutcomeSkipped:
			summary.Skipped += 1
		case OutcomeFailed:
			summary.Failed += 1
		default:
			summary.Failed += 1
		}
	}

	return &summary, details
}

func (s *UserImportService) ImportRecord(ctx context.Context, options *Options, rawMessage json.RawMessage) (outcome Outcome, warnings []Warning, err error) {
	var record Record
	err = RecordSchema.Validator().ParseJSONRawMessage(rawMessage, &record)
	if err != nil {
		return
	}

	outcome = OutcomeSkipped
	return
}
