package userimport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type importResult struct {
	Outcome  Outcome
	Warnings []Warning
}

type IdentityService interface {
	ListByClaim(name string, value string) ([]*identity.Info, error)
}

type UserImportService struct {
	AppDatabase *appdb.Handle
	Identities  IdentityService
	Logger      Logger
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("user-import")}
}

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
		var result importResult
		err := s.AppDatabase.WithTx(func() error {
			err := s.ImportRecordInTxn(ctx, &result, options, rawMessage)
			if err != nil {
				return err
			}
			return nil
		})

		if len(result.Warnings) > 0 {
			detail.Warnings = result.Warnings
			hasDetail = true
		}
		if err != nil {
			if !apierrors.IsAPIError(err) {
				s.Logger.WithError(err).Error(err.Error())
			}
			detail.Errors = []*apierrors.APIError{apierrors.AsAPIError(err)}
			hasDetail = true
		}

		if hasDetail {
			details = append(details, detail)
		}

		switch result.Outcome {
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

func (s *UserImportService) ImportRecordInTxn(ctx context.Context, result *importResult, options *Options, rawMessage json.RawMessage) (err error) {
	var record Record

	err = options.RecordSchema().Validator().ParseJSONRawMessage(rawMessage, &record)
	if err != nil {
		return
	}

	var infos []*identity.Info
	switch options.Identifier {
	case IdentifierEmail:
		emailPtr, _ := record.Email()
		infos, err = s.Identities.ListByClaim(IdentifierEmail, *emailPtr)
	case IdentifierPhoneNumber:
		phoneNumberPtr, _ := record.PhoneNumber()
		infos, err = s.Identities.ListByClaim(IdentifierPhoneNumber, *phoneNumberPtr)
	case IdentifierPreferredUsername:
		preferredUsernamePtr, _ := record.PreferredUsername()
		infos, err = s.Identities.ListByClaim(IdentifierPreferredUsername, *preferredUsernamePtr)
	default:
		err = fmt.Errorf("unknown identifier: %v", options.Identifier)
	}
	if err != nil {
		return
	}

	switch len(infos) {
	case 0:
		return s.insertRecordInTxn(ctx, result, options, record)
	case 1:
		if options.Upsert {
			// TODO(userimport): update
			err = fmt.Errorf("upsert is not implemented yet")
			return
		} else {
			result.Outcome = OutcomeSkipped
			result.Warnings = append(result.Warnings, Warning{
				Message: "skipping because upsert = false and user exists",
			})
			return
		}
	default:
		err = fmt.Errorf("unexpected number of identities found: %v", len(infos))
		return
	}
}

func (s *UserImportService) insertRecordInTxn(ctx context.Context, result *importResult, options *Options, record Record) (err error) {
	err = fmt.Errorf("insert is not implemented yet")
	return
}
