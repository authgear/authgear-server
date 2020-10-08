package service

import (
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/authgear/authgear-server/pkg/portal/db"
	"github.com/authgear/authgear-server/pkg/portal/model"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

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

func (s *DomainService) listDomains(appID string, isVerified bool) ([]*model.Domain, error) {
	var tableName string
	if isVerified {
		tableName = "domain"
	} else {
		tableName = "pending_domain"
	}

	rows, err := s.SQLExecutor.QueryWith(
		s.SQLBuilder.
			Select("id", "app_id", "created_at", "domain", "apex_domain", "verification_nonce").
			Where("app_id = ?", appID).
			From(s.SQLBuilder.FullTableName(tableName)),
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
	return &model.Domain{
		ID:                    d.ID,
		CreatedAt:             d.CreatedAt,
		Domain:                d.Domain,
		ApexDomain:            d.ApexDomain,
		VerificationDNSRecord: domainVerificationDNSRecord(d.VerificationNonce),
		IsVerified:            isVerified,
	}
}

func domainVerificationDNSRecord(nonce string) string {
	return fmt.Sprintf("authgear-verification=%s", nonce)
}
