package internal

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"golang.org/x/net/publicsuffix"

	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

type SetupPortalOptions struct {
	DatabaseURL    string
	DatabaseSchema string
	DefaultDoamin  string
	CustomDomain   string
	ResourceDir    string
}

func SetupPortal(opt *SetupPortalOptions) {
	// parse id from authgear.yaml
	appID := "accounts"

	db := openDB(opt.DatabaseURL, opt.DatabaseSchema)

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("failed to connect db: %s", err)
	}

	if err := createDomain(ctx, tx, appID, opt.DefaultDoamin, false); err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to create default domain: %s", err)
	}

	if err := createDomain(ctx, tx, appID, opt.CustomDomain, true); err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to create custom domain: %s", err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func createDomain(ctx context.Context, tx *sql.Tx, appID string, domain string, isCustom bool) error {
	apexDomain, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return fmt.Errorf("invalid domain: %w", err)
	}
	if !isCustom {
		// For non-custom domain, assume the domain is always an apex domain,
		// in case the domain suffix is not yet in PSL.
		apexDomain = domain
	}

	nonce := make([]byte, 16)
	corerand.SecureRand.Read(nonce)
	verificationNonce := hex.EncodeToString(nonce)

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
			false,
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
