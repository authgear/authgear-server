package service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/db"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
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

var DomainVerificationFailed = apierrors.Forbidden.WithReason("DomainVerificationFailed")

type DomainService struct {
	Context     context.Context
	Clock       clock.Clock
	SQLBuilder  *db.SQLBuilder
	SQLExecutor *db.SQLExecutor
}

func (s *DomainService) ListDomains(appID string) ([]*model.Domain, error) {
	pendingDomains, err := s.listDomains(appID, false)
	if err != nil {
		return nil, err
	}

	domains, err := s.listDomains(appID, true)
	if err != nil {
		return nil, err
	}

	return append(pendingDomains, domains...), nil
}

func (s *DomainService) CreateDomain(appID string, domain string) (*model.Domain, error) {
	d, err := newDomain(appID, domain, s.Clock.NowUTC())
	if err != nil {
		return nil, err
	}

	err = s.createDomain(d, false)
	if err != nil {
		return nil, err
	}

	return d.toModel(false), nil
}

func (s *DomainService) DeleteDomain(appID string, id string) error {
	isVerified, id, ok := parseDomainID(id)
	if !ok {
		return ErrDomainNotFound
	}

	err := s.deleteDomain(appID, id, isVerified)
	if err != nil {
		return err
	}

	// TODO(domain): cleanup ingress
	return nil
}

func (s *DomainService) listDomains(appID string, isVerified bool) ([]*model.Domain, error) {
	rows, err := s.SQLExecutor.QueryWith(
		s.SQLBuilder.
			Select("id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce").
			Where("app_id = ?", appID).
			From(s.SQLBuilder.FullTableName(domainTableName(isVerified))),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*model.Domain
	for rows.Next() {
		d, err := scanDomain(rows)
		if err != nil {
			return nil, err
		}
		domains = append(domains, d.toModel(isVerified))
	}

	return domains, nil
}

func (s *DomainService) VerifyDomain(appID string, id string) (*model.Domain, error) {
	isVerified, id, ok := parseDomainID(id)
	if !ok {
		return nil, ErrDomainNotFound
	}

	if isVerified {
		return nil, ErrDomainVerified
	}

	d, err := s.getDomain(appID, id, false)
	if err != nil {
		return nil, err
	}

	err = s.verifyDomain(d)
	if err != nil {
		return nil, DomainVerificationFailed.Errorf("domain verification failed: %w", err)
	}

	// Migrate the domain from pending domains to domains
	err = s.deleteDomain(d.AppID, d.ID, false)
	if err != nil {
		return nil, err
	}

	d.CreatedAt = s.Clock.NowUTC()
	err = s.createDomain(d, true)
	if err != nil {
		return nil, err
	}

	// TODO(domain): create ingress
	return d.toModel(true), nil
}

func (s *DomainService) verifyDomain(domain *domain) error {
	ctx, cancel := context.WithTimeout(s.Context, DomainVerificationTimeout)
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

func (s *DomainService) getDomain(appID string, id string, isVerified bool) (*domain, error) {
	row, err := s.SQLExecutor.QueryRowWith(
		s.SQLBuilder.
			Select("id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce").
			Where("app_id = ? AND id = ?", appID, id).
			From(s.SQLBuilder.FullTableName(domainTableName(isVerified))),
	)
	if err != nil {
		return nil, err
	}

	return scanDomain(row)
}

func (s *DomainService) createDomain(d *domain, isVerified bool) error {
	tableName := domainTableName(isVerified)
	dupeQuery := s.SQLBuilder.
		Select("COUNT(*)").
		From(s.SQLBuilder.FullTableName(tableName))

	dupeQuery = dupeQuery.Where("apex_domain = ?", d.ApexDomain)
	if !isVerified {
		// Limit duplication query to within app for pending domains
		dupeQuery = dupeQuery.Where("app_id = ?", d.AppID)
	}

	scanner, err := s.SQLExecutor.QueryRowWith(dupeQuery)
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

	_, err = s.SQLExecutor.ExecWith(s.SQLBuilder.
		Insert(s.SQLBuilder.FullTableName(tableName)).
		Columns(
			"id",
			"app_id",
			"created_at",
			"domain",
			"apex_domain",
			"verification_nonce",
		).
		Values(
			d.ID,
			d.AppID,
			d.CreatedAt,
			d.Domain,
			d.ApexDomain,
			d.VerificationNonce,
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *DomainService) deleteDomain(appID string, id string, isVerified bool) error {
	_, err := s.SQLExecutor.ExecWith(s.SQLBuilder.
		Delete(s.SQLBuilder.FullTableName(domainTableName(isVerified))).
		Where("app_id = ? AND id = ?", appID, id),
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
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrDomainNotFound
	} else if err != nil {
		return nil, err
	}

	return d, nil
}

func newDomain(appID string, domainName string, createdAt time.Time) (*domain, error) {
	nonce := make([]byte, 16)
	corerand.SecureRand.Read(nonce)

	apexDomain, err := publicsuffix.EffectiveTLDPlusOne(domainName)
	if err != nil {
		return nil, err
	}

	return &domain{
		ID:                uuid.New(),
		AppID:             appID,
		CreatedAt:         createdAt,
		Domain:            domainName,
		ApexDomain:        apexDomain,
		VerificationNonce: hex.EncodeToString(nonce),
	}, nil
}

func (d *domain) toModel(isVerified bool) *model.Domain {
	var prefix string
	if isVerified {
		prefix = "verified:"
	} else {
		prefix = "pending:"
	}

	return &model.Domain{
		ID:                    prefix + d.ID,
		CreatedAt:             d.CreatedAt,
		Domain:                d.Domain,
		ApexDomain:            d.ApexDomain,
		VerificationDNSRecord: domainVerificationDNSRecord(d.VerificationNonce),
		IsVerified:            isVerified,
	}
}

func parseDomainID(modelID string) (isVerified bool, id string, ok bool) {
	parts := strings.Split(modelID, ":")
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
		return "domain"
	}
	return "pending_domain"
}

func domainVerificationDNSRecord(nonce string) string {
	return fmt.Sprintf("authgear-verification=%s", nonce)
}
