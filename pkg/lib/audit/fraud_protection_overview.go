package audit

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type FraudProtectionOverview struct {
	SendSMS FraudProtectionOverviewSendSMS `json:"sendSMS"`
}

type FraudProtectionOverviewSendSMS struct {
	TotalActions   int                         `json:"totalActions"`
	BlockedActions int                         `json:"blockedActions"`
	WarnedActions  int                         `json:"warnedActions"`
	TopSourceIPs   []FraudProtectionOverviewIP `json:"topSourceIPs"`
}

type FraudProtectionOverviewIP struct {
	IPAddress string `json:"ipAddress"`
	Total     int    `json:"total"`
	Blocked   int    `json:"blocked"`
	Flagged   int    `json:"flagged"`
}

type FraudProtectionOverviewQueryOptions struct {
	RangeFrom *time.Time
	RangeTo   *time.Time
	Actions   []model.FraudProtectionAction
}

func (o FraudProtectionOverviewQueryOptions) Apply(q db.SelectBuilder) db.SelectBuilder {
	if o.RangeFrom != nil {
		q = q.Where("created_at >= ?", o.RangeFrom)
	}
	if o.RangeTo != nil {
		q = q.Where("created_at < ?", o.RangeTo)
	}

	q = q.Where("activity_type = ?", string(nonblocking.FraudProtectionDecisionRecorded))

	if len(o.Actions) > 0 {
		actions := make([]string, 0, len(o.Actions))
		for _, action := range o.Actions {
			actions = append(actions, string(action))
		}
		q = q.Where("data#>>'{payload,record,action}' = ANY (?)", pq.Array(actions))
	}

	return q
}

func (s *ReadStore) fraudProtectionDecisionRecordsQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"created_at",
			"user_id",
			"activity_type",
			"data",
			"COALESCE(host(ip_address)::text, '') AS ip_address",
			"COALESCE(data#>>'{payload,record,decision}', '') AS decision",
			"COALESCE(jsonb_array_length((data->'payload'->'record'->'triggered_warnings')::jsonb), 0) AS warning_count",
		).
		From(s.SQLBuilder.TableName("_audit_log"))
}

func (s *ReadStore) GetFraudProtectionOverview(ctx context.Context, opts FraudProtectionOverviewQueryOptions) (*FraudProtectionOverview, error) {
	baseQuery := s.fraudProtectionDecisionRecordsQuery()
	baseQuery = opts.Apply(baseQuery)

	totalQuery := s.SQLBuilder.
		Select(
			"COUNT(*) AS total_actions",
			"COUNT(*) FILTER (WHERE decision = 'blocked') AS blocked_actions",
			"COUNT(*) FILTER (WHERE decision = 'allowed' AND warning_count > 0) AS warning_actions",
		).
		FromSelect(baseQuery, "records")

	row, err := s.SQLExecutor.QueryRowWith(ctx, totalQuery)
	if err != nil {
		return nil, err
	}

	var totalActions int64
	var blockedActions int64
	var warnedActions int64
	if err := row.Scan(&totalActions, &blockedActions, &warnedActions); err != nil {
		return nil, err
	}

	topIPsQuery := s.SQLBuilder.
		Select(
			"ip_address",
			"COUNT(*) AS total_actions",
			"COUNT(*) FILTER (WHERE decision = 'blocked') AS blocked_actions",
			"COUNT(*) FILTER (WHERE decision = 'allowed' AND warning_count > 0) AS warning_actions",
		).
		FromSelect(baseQuery, "records").
		Where("ip_address <> ''").
		GroupBy("ip_address").
		OrderBy("total_actions DESC", "ip_address ASC").
		Limit(10)

	rows, err := s.SQLExecutor.QueryWith(ctx, topIPsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topSourceIPs := make([]FraudProtectionOverviewIP, 0)
	for rows.Next() {
		var ip sql.NullString
		var item FraudProtectionOverviewIP
		var total int64
		var blocked int64
		var warnings int64
		if err := rows.Scan(&ip, &total, &blocked, &warnings); err != nil {
			return nil, err
		}
		if ip.Valid {
			item.IPAddress = ip.String
		}
		item.Total = int(total)
		item.Blocked = int(blocked)
		item.Flagged = int(warnings)
		topSourceIPs = append(topSourceIPs, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &FraudProtectionOverview{
		SendSMS: FraudProtectionOverviewSendSMS{
			TotalActions:   int(totalActions),
			BlockedActions: int(blockedActions),
			WarnedActions:  int(warnedActions),
			TopSourceIPs:   topSourceIPs,
		},
	}, nil
}
