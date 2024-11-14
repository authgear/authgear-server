package internal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type CreateCustomDomainOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	AppID          string
	Domain         string
	ApexDomain     string
}

func CreateCustomDomain(ctx context.Context, opts CreateCustomDomainOptions) (err error) {
	db := openDB(opts.DatabaseURL, opts.DatabaseSchema)

	tx, err := db.BeginTx(ctx, nil)
	defer func() {
		if err != nil {
			err = tx.Rollback()
		}
	}()

	exists, err := checkDomainExistance(ctx, tx, opts.AppID, opts.Domain)
	if err != nil {
		return
	}

	if !exists {
		err = createCustomDomain(ctx, tx, opts.AppID, opts.Domain, opts.ApexDomain)
		if err != nil {
			return
		}
		fmt.Printf("created: %s\n", opts.Domain)
	} else {
		fmt.Printf("skipped: %s\n", opts.Domain)
	}

	err = tx.Commit()
	if err != nil {
		return
	}

	return
}

func createCustomDomain(ctx context.Context, tx *sql.Tx, appID string, domain string, apexDomain string) error {
	isCustom := true
	verificationNonce := service.MakeVerificationNonce()

	builder := newSQLBuilder().
		Insert(pq.QuoteIdentifier("_portal_domain")).
		Columns(
			"id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce", "is_custom",
		).
		Values(
			uuid.New(),
			appID,
			time.Now().UTC(),
			domain,
			apexDomain,
			verificationNonce,
			isCustom,
		)

	q, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}
