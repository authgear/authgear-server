package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"golang.org/x/net/publicsuffix"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const DomainVerificationTimeout = 10 * time.Second

var ErrDomainDuplicated = apierrors.AlreadyExists.WithReason("DuplicatedDomain").
	New("requested domain is already in use")

var ErrDomainVerified = apierrors.AlreadyExists.WithReason("DomainVerified").
	New("requested domain is already verified")

var ErrDomainNotFound = apierrors.NotFound.WithReason("DomainNotFound").
	New("domain not found")

var ErrDomainNotCustom = apierrors.Forbidden.WithReason("DomainNotCustom").
	New("requested domain is not a custom domain")

var DomainVerificationFailed = apierrors.Forbidden.WithReason("DomainVerificationFailed")
var InvalidDomain = apierrors.Invalid.WithReason("InvalidDomain")

type DomainConfigService interface {
	CreateDomain(ctx context.Context, appID string, domainID string, domain string, isCustom bool) error
	DeleteDomain(ctx context.Context, domain *apimodel.Domain) error
}

type DomainService struct {
	Clock          clock.Clock
	DomainConfig   DomainConfigService
	SQLBuilder     *globaldb.SQLBuilder
	SQLExecutor    *globaldb.SQLExecutor
	GlobalDatabase *globaldb.Handle
}

// GetMany acquires connection.
func (s *DomainService) GetMany(ctx context.Context, ids []string) ([]*apimodel.Domain, error) {
	var rawIDs []string
	for _, id := range ids {
		_, rawID, ok := parseDomainID(id)
		if ok {
			rawIDs = append(rawIDs, rawID)
		}
	}

	var pendingDomains []*apimodel.Domain
	var domains []*apimodel.Domain
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		pendingDomains, err = s.listDomains(ctx, rawIDs, false)
		if err != nil {
			return err
		}
		domains, err = s.listDomains(ctx, rawIDs, true)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var out []*apimodel.Domain
	out = append(out, pendingDomains...)
	out = append(out, domains...)
	return out, nil
}

// ListDomains acquires connection.
func (s *DomainService) ListDomains(ctx context.Context, appID string) ([]*apimodel.Domain, error) {
	var pendingDomains []*apimodel.Domain
	var domains []*apimodel.Domain
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		pendingDomains, err = s.listDomainsByAppID(ctx, appID, false)
		if err != nil {
			return err
		}

		domains, err = s.listDomainsByAppID(ctx, appID, true)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	result := append(pendingDomains, domains...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Domain < result[j].Domain
	})

	return result, nil
}

// CreateCustomDomain acquires connection.
func (s *DomainService) CreateCustomDomain(ctx context.Context, appID string, domain string) (*apimodel.Domain, error) {
	var out *apimodel.Domain
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		out, err = s.CreateDomain(ctx, appID, domain, false, true)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return out, nil
}

// CreateDomain assumes acquired connection.
func (s *DomainService) CreateDomain(ctx context.Context, appID string, domain string, isVerified bool, isCustom bool) (*apimodel.Domain, error) {
	d, err := newDomain(appID, domain, s.Clock.NowUTC(), isCustom)
	if err != nil {
		return nil, err
	}

	if !isCustom {
		// For non-custom domain, assume the domain is always an apex domain,
		// in case the domain suffix is not yet in PSL.
		d.ApexDomain = d.Domain
	}

	err = s.createDomain(ctx, d, isVerified)

	if err != nil {
		return nil, err
	}

	domainModel := d.toModel(isVerified)
	if isVerified {
		err = s.DomainConfig.CreateDomain(ctx, appID, domainModel.ID, domainModel.Domain, domainModel.IsCustom)
		if err != nil {
			return nil, err
		}
	}
	return domainModel, nil
}

// DeleteDomain assumes acquired connection.
func (s *DomainService) DeleteDomain(ctx context.Context, appID string, id string) error {
	isVerified, id, ok := parseDomainID(id)
	if !ok {
		return ErrDomainNotFound
	}

	d, err := s.getDomain(ctx, appID, id, isVerified)
	if err != nil {
		return err
	}

	err = s.deleteDomain(ctx, d, isVerified)
	if err != nil {
		return err
	}

	err = s.DomainConfig.DeleteDomain(ctx, d.toModel(isVerified))
	if err != nil {
		return err
	}

	return nil
}

func (s *DomainService) listDomains(ctx context.Context, ids []string, isVerified bool) ([]*apimodel.Domain, error) {
	rows, err := s.SQLExecutor.QueryWith(
		ctx,
		s.SQLBuilder.
			Select("id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce", "is_custom").
			Where("id = ANY (?)", pq.Array(ids)).
			From(s.SQLBuilder.TableName(domainTableName(isVerified))),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*apimodel.Domain
	for rows.Next() {
		d, err := scanDomain(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d.toModel(isVerified))
	}

	return domains, nil
}

func (s *DomainService) listDomainsByAppID(ctx context.Context, appID string, isVerified bool) ([]*apimodel.Domain, error) {
	rows, err := s.SQLExecutor.QueryWith(
		ctx,
		s.SQLBuilder.
			Select("id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce", "is_custom").
			Where("app_id = ?", appID).
			From(s.SQLBuilder.TableName(domainTableName(isVerified))),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*apimodel.Domain
	for rows.Next() {
		d, err := scanDomain(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d.toModel(isVerified))
	}

	return domains, nil
}

// VerifyDomain acquires connection.
func (s *DomainService) VerifyDomain(ctx context.Context, appID string, id string) (*apimodel.Domain, error) {
	isVerified, id, ok := parseDomainID(id)
	if !ok {
		return nil, ErrDomainNotFound
	}

	if isVerified {
		return nil, ErrDomainVerified
	}

	var d *domain
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		d, err = s.getDomain(ctx, appID, id, false)
		if err != nil {
			return err
		}
		return nil
	})

	err = s.verifyDomain(ctx, d)
	if err != nil {
		return nil, DomainVerificationFailed.Errorf("domain verification failed: %w", err)
	}

	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		// Migrate the domain from pending domains to domains
		err = s.deleteDomain(ctx, d, false)
		if err != nil {
			return err
		}

		d.CreatedAt = s.Clock.NowUTC()
		err = s.createDomain(ctx, d, true)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	domainModel := d.toModel(true)
	err = s.DomainConfig.CreateDomain(ctx, appID, domainModel.ID, domainModel.Domain, domainModel.IsCustom)
	if err != nil {
		return nil, err
	}
	return domainModel, nil
}

func (s *DomainService) verifyDomain(ctx context.Context, domain *domain) error {
	ctx, cancel := context.WithTimeout(ctx, DomainVerificationTimeout)
	defer cancel()

	resolver := &net.Resolver{}
	txtRecords, err := resolver.LookupTXT(ctx, domain.ApexDomain)
	if err != nil {
		return fmt.Errorf("failed to fetch TXT record: %w", err)
	}

	expectedRecord := domainVerificationDNSRecord(domain.VerificationNonce)
	found := false
	for _, record := range txtRecords {
		if record == expectedRecord {
			found = true
			break
		}
	}
	if !found {
		return errors.New("expected TXT record not found")
	}

	return nil
}

func (s *DomainService) getDomain(ctx context.Context, appID string, id string, isVerified bool) (*domain, error) {
	row, err := s.SQLExecutor.QueryRowWith(
		ctx,
		s.SQLBuilder.
			Select("id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce", "is_custom").
			Where("app_id = ? AND id = ?", appID, id).
			From(s.SQLBuilder.TableName(domainTableName(isVerified))),
	)
	if err != nil {
		return nil, err
	}

	return scanDomain(row)
}

func (s *DomainService) createDomain(ctx context.Context, d *domain, isVerified bool) error {
	tableName := domainTableName(isVerified)
	dupeQuery := s.SQLBuilder.
		Select("COUNT(*)").
		From(s.SQLBuilder.TableName(tableName))

	dupeQuery = dupeQuery.Where("apex_domain = ?", d.ApexDomain)
	if !isVerified {
		// Limit duplication query to within app for pending domains
		dupeQuery = dupeQuery.Where("app_id = ?", d.AppID)
	}

	scanner, err := s.SQLExecutor.QueryRowWith(ctx, dupeQuery)
	if err != nil {
		return err
	}

	var count uint64
	if err = scanner.Scan(&count); err != nil {
		return err
	}

	if count >= 1 {
		return ErrDomainDuplicated
	}

	_, err = s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Insert(s.SQLBuilder.TableName(tableName)).
		Columns(
			"id",
			"app_id",
			"created_at",
			"domain",
			"apex_domain",
			"verification_nonce",
			"is_custom",
		).
		Values(
			d.ID,
			d.AppID,
			d.CreatedAt,
			d.Domain,
			d.ApexDomain,
			d.VerificationNonce,
			d.IsCustom,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *DomainService) deleteDomain(ctx context.Context, d *domain, isVerified bool) error {
	if !d.IsCustom {
		return ErrDomainNotCustom
	}

	_, err := s.SQLExecutor.ExecWith(ctx, s.SQLBuilder.
		Delete(s.SQLBuilder.TableName(domainTableName(isVerified))).
		Where("id = ?", d.ID),
	)
	if err != nil {
		return err
	}

	return nil
}

type domain struct {
	ID                string
	AppID             string
	CreatedAt         time.Time
	Domain            string
	ApexDomain        string
	VerificationNonce string
	IsCustom          bool
}

func scanDomain(scn db.Scanner) (*domain, error) {
	d := &domain{}
	err := scn.Scan(
		&d.ID,
		&d.AppID,
		&d.CreatedAt,
		&d.Domain,
		&d.ApexDomain,
		&d.VerificationNonce,
		&d.IsCustom,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDomainNotFound
	} else if err != nil {
		return nil, err
	}

	return d, nil
}

func MakeVerificationNonce() string {
	nonce := make([]byte, 16)
	corerand.SecureRand.Read(nonce)
	verificationNonce := hex.EncodeToString(nonce)
	return verificationNonce
}

func newDomain(appID string, domainName string, createdAt time.Time, isCustom bool) (*domain, error) {
	verificationNonce := MakeVerificationNonce()

	apexDomain, err := publicsuffix.EffectiveTLDPlusOne(domainName)
	if err != nil {
		return nil, InvalidDomain.Errorf("invalid domain: %w", err)
	}

	return &domain{
		ID:                uuid.New(),
		AppID:             appID,
		CreatedAt:         createdAt,
		Domain:            domainName,
		ApexDomain:        apexDomain,
		VerificationNonce: verificationNonce,
		IsCustom:          isCustom,
	}, nil
}

func (d *domain) toModel(isVerified bool) *apimodel.Domain {
	var prefix string
	if isVerified {
		prefix = "verified:"
	} else {
		prefix = "pending:"
	}

	// for default domain, original domain will be used for cookie domain
	// for custom domain, cookie domain is derived from the
	// CookieDomainWithoutPort function
	cookieDomain := d.Domain
	if d.IsCustom {
		cookieDomain = httputil.CookieDomainWithoutPort(d.Domain)
	}

	return &apimodel.Domain{
		// Base64-encoded to avoid invalid k8s resource label invalid chars
		ID:                    base64.RawURLEncoding.EncodeToString([]byte(prefix + d.ID)),
		AppID:                 d.AppID,
		CreatedAt:             d.CreatedAt,
		Domain:                d.Domain,
		CookieDomain:          cookieDomain,
		ApexDomain:            d.ApexDomain,
		VerificationDNSRecord: domainVerificationDNSRecord(d.VerificationNonce),
		IsCustom:              d.IsCustom,
		IsVerified:            isVerified,
	}
}

func parseDomainID(modelID string) (isVerified bool, id string, ok bool) {
	// Base64-encoded to avoid invalid k8s resource label invalid chars
	rawID, err := base64.RawURLEncoding.DecodeString(modelID)
	if err != nil {
		return
	}

	parts := strings.Split(string(rawID), ":")
	if len(parts) != 2 {
		return
	}
	switch parts[0] {
	case "verified":
		isVerified = true
	case "pending":
		isVerified = false
	default:
		return
	}
	id = parts[1]
	ok = true
	return
}

func domainTableName(isVerified bool) string {
	if isVerified {
		return "_portal_domain"
	}
	return "_portal_pending_domain"
}

func domainVerificationDNSRecord(nonce string) string {
	return fmt.Sprintf("authgear-verification=%s", nonce)
}
