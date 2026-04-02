package audit

import (
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type FraudProtectionDecisionRecord struct {
	ID        string                            `json:"id"`
	CreatedAt time.Time                         `json:"createdAt"`
	Record    model.FraudProtectionDecisionRecord `json:"record"`
}

type FraudProtectionDecisionRecordQueryOptions struct {
	RangeFrom     *time.Time
	RangeTo       *time.Time
	SortDirection model.SortDirection
	Decisions     []model.FraudProtectionDecision
	Search        *string
	ReasonCodes   []string
}

func (o FraudProtectionDecisionRecordQueryOptions) Apply(q db.SelectBuilder) db.SelectBuilder {
	q = q.Where("activity_type = ?", string(nonblocking.FraudProtectionDecisionRecorded))

	if o.RangeFrom != nil {
		q = q.Where("created_at >= ?", o.RangeFrom)
	}
	if o.RangeTo != nil {
		q = q.Where("created_at < ?", o.RangeTo)
	}
	if len(o.Decisions) > 0 {
		decisions := make([]string, 0, len(o.Decisions))
		for _, decision := range o.Decisions {
			decisions = append(decisions, string(decision))
		}
		q = q.Where("data#>>'{payload,record,decision}' = ANY (?)", pq.Array(decisions))
	}
	if o.Search != nil && *o.Search != "" {
		search := strings.TrimSpace(*o.Search)
		searchUpper := strings.ToUpper(search)
		q = q.Where(
			"(data#>>'{payload,record,action_detail,recipient}' LIKE ? OR data#>>'{payload,record,action_detail,phone_number_country_code}' LIKE ? OR data#>>'{payload,record,geo_location_code}' LIKE ? OR data#>>'{payload,record,ip_address}' LIKE ?)",
			search+"%",
			searchUpper+"%",
			searchUpper+"%",
			search+"%",
		)
	}
	if len(o.ReasonCodes) > 0 {
		q = q.Where(
			`EXISTS (
				SELECT 1
				FROM jsonb_array_elements_text(COALESCE(data->'payload'->'record'->'triggered_warnings', '[]'::jsonb)) AS warning(code)
				WHERE warning.code = ANY (?)
			)`,
			pq.Array(o.ReasonCodes),
		)
	}

	return q
}
