package service

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/portal/db"
	"github.com/authgear/authgear-server/pkg/portal/model"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

var ErrDomainDuplicated = apierrors.AlreadyExists.WithReason("DuplicatedDomain").
	New("requested domain is already in use")

var ErrDomainNotFound = apierrors.NotFound.WithReason("DomainNotFound").
	New("domain not found")

type DomainService struct {
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
	d, err := newDomain(appID, domain)
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
	if err := scn.Scan(
		&d.ID,
		&d.AppID,
		&d.CreatedAt,
		&d.Domain,
		&d.ApexDomain,
		&d.VerificationNonce,
	); err != nil {
		return nil, err
	}

	return d, nil
}

func newDomain(appID string, domainName string) (*domain, error) {
	nonce := make([]byte, 16)
	corerand.SecureRand.Read(nonce)

	apexDomain, err := publicsuffix.EffectiveTLDPlusOne(domainName)
	if err != nil {
		return nil, err
	}

	return &domain{
		ID:                uuid.New(),
		AppID:             appID,
		CreatedAt:         time.Now(),
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
