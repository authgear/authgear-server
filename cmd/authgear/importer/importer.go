package importer

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type ImportOptions struct {
	EmailMarkAsVerified bool
}

type Importer struct {
	AppID         config.AppID
	Handle        *appdb.Handle
	SQLBuilderApp *appdb.SQLBuilderApp
	SQLExecutor   *appdb.SQLExecutor
	EmailConfig   *config.LoginIDEmailConfig
}

func (i *Importer) ImportRecord(ctx context.Context, record []string, opts ImportOptions, now time.Time) error {
	userID := record[0]
	rawEmail := record[1]
	name := record[2]
	passwordHash := record[3]

	emailChecker := loginid.EmailChecker{
		Config: i.EmailConfig,
	}
	validationCtx := &validation.Context{}
	emailChecker.Validate(validationCtx, rawEmail)
	err := validationCtx.Error(fmt.Sprintf("invalid email: %v", rawEmail))
	if err != nil {
		return err
	}

	emailNormalizer := loginid.EmailNormalizer{
		Config: i.EmailConfig,
	}

	loginID, err := emailNormalizer.Normalize(rawEmail)
	if err != nil {
		return err
	}

	uniqueKey, err := emailNormalizer.ComputeUniqueKey(loginID)
	if err != nil {
		return err
	}

	claims, err := Claims(loginID)
	if err != nil {
		return err
	}

	stdAttrs, err := StandardAttributes(name, loginID)
	if err != nil {
		return err
	}

	insertStmts := []db.InsertBuilder{
		i.SQLBuilderApp.Insert(i.SQLBuilderApp.TableName("_auth_user")).Columns(
			"id",
			"created_at",
			"updated_at",
			"is_disabled",
			"standard_attributes",
			"require_reindex_after",
		).Values(
			userID,
			now,
			now,
			false,
			stdAttrs,
			now,
		).Suffix("ON CONFLICT (id) DO UPDATE SET standard_attributes = EXCLUDED.standard_attributes"),
		i.SQLBuilderApp.Insert(i.SQLBuilderApp.TableName("_auth_identity")).Columns(
			"id",
			"type",
			"user_id",
			"created_at",
			"updated_at",
		).Values(
			userID,
			"login_id",
			userID,
			now,
			now,
		).Suffix("ON CONFLICT DO NOTHING"),
		i.SQLBuilderApp.Insert(i.SQLBuilderApp.TableName("_auth_identity_login_id")).Columns(
			"id",
			"login_id_key",
			"login_id",
			"claims",
			"original_login_id",
			"unique_key",
			"login_id_type",
		).Values(
			userID,
			"email",
			loginID,
			claims,
			rawEmail,
			uniqueKey,
			"email",
		).Suffix("ON CONFLICT (id) DO UPDATE SET login_id = EXCLUDED.login_id, claims = EXCLUDED.claims, original_login_id = EXCLUDED.original_login_id, unique_key = EXCLUDED.unique_key"),
		i.SQLBuilderApp.Insert(i.SQLBuilderApp.TableName("_auth_authenticator")).Columns(
			"id",
			"type",
			"user_id",
			"created_at",
			"updated_at",
			"is_default",
			"kind",
		).Values(
			userID,
			"password",
			userID,
			now,
			now,
			true,
			"primary",
		).Suffix("ON CONFLICT DO NOTHING"),
		i.SQLBuilderApp.Insert(i.SQLBuilderApp.TableName("_auth_authenticator_password")).Columns(
			"id",
			"password_hash",
		).Values(
			userID,
			passwordHash,
		).Suffix("ON CONFLICT (id) DO UPDATE SET password_hash = EXCLUDED.password_hash"),
	}

	if opts.EmailMarkAsVerified {
		insertStmts = append(insertStmts, i.SQLBuilderApp.Insert(i.SQLBuilderApp.TableName("_auth_verified_claim")).Columns(
			"id",
			"user_id",
			"name",
			"value",
			"created_at",
		).Values(
			userID,
			userID,
			"email",
			loginID,
			now,
		).Suffix("ON CONFLICT (id) DO UPDATE SET value = EXCLUDED.value"))
	}

	err = i.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		for _, stmt := range insertStmts {
			_, err = i.SQLExecutor.ExecWith(ctx, stmt)
			if err != nil {
				return
			}
		}
		return
	})
	if err != nil {
		return err
	}

	return nil
}

func (i *Importer) ImportFromCSV(ctx context.Context, csvPath string, opts ImportOptions) error {
	now := time.Now().UTC()

	f, err := os.Open(csvPath)
	if err != nil {
		return err
	}

	headerSkipped := false
	r := csv.NewReader(f)
	numTotal := 0
	numGood := 0
	numBad := 0
	for {
		record, err := r.Read()
		// This is an infinite loop.
		// According to csv.Reader.Read, err is io.EOF when the file reaches EOF.
		// So this is the condition we exit the loop normally.
		if errors.Is(err, io.EOF) {
			break
		}
		// Otherwise we encounter any other io error that we do not know how to handle.
		// In this case we exit the loop.
		if err != nil {
			return err
		}

		// Skip the header if we haven't do so.
		if !headerSkipped {
			headerSkipped = true
			continue
		}

		numTotal++
		err = i.ImportRecord(ctx, record, opts, now)
		if err != nil {
			numBad++
			fmt.Fprintf(os.Stderr, "%v:%v: %v\n", csvPath, numTotal, err)
		} else {
			numGood++
		}
	}

	if numBad > 0 {
		return fmt.Errorf("total: %v; imported: %v; error: %v", numTotal, numGood, numBad)
	}

	fmt.Printf("total: %v; imported: %v; error: %v\n", numTotal, numGood, numBad)
	return nil
}

func StandardAttributes(name string, email string) ([]byte, error) {
	attrs := map[string]interface{}{
		"name":  name,
		"email": email,
	}
	return json.Marshal(attrs)
}

func Claims(email string) ([]byte, error) {
	claims := map[string]interface{}{
		"email": email,
	}
	return json.Marshal(claims)
}
