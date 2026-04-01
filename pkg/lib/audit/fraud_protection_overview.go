package audit

import (
	"context"
	"database/sql"

	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type FraudProtectionOverview struct {
	TotalActions   int                           `json:"totalActions"`
	AllowedActions int                           `json:"allowedActions"`
	BlockedActions int                           `json:"blockedActions"`
	WarnedActions int                           `json:"warnedActions"`
	TopSourceIPs  []FraudProtectionOverviewIP `json:"topSourceIPs"`
}

type FraudProtectionOverviewIP struct {
	IPAddress      string `json:"ipAddress"`
	TotalActions   int    `json:"totalActions"`
	BlockedActions int    `json:"blockedActions"`
	WarnedActions int    `json:"warnedActions"`
}

func (s *ReadStore) fraudProtectionDecisionRecordsQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"created_at",
			"user_id",
			"activity_type",
			"data",
			"COALESCE(data#>>'{payload,record,ip_address}', '') AS ip_address",
			"COALESCE(data#>>'{payload,record,decision}', '') AS decision",
			"COALESCE(jsonb_array_length((data->'payload'->'record'->'triggered_warnings')::jsonb), 0) AS warning_count",
		).
		From(s.SQLBuilder.TableName("_audit_log"))
}

func (s *ReadStore) GetFraudProtectionOverview(ctx context.Context, opts QueryPageOptions) (*FraudProtectionOverview, error) {
	baseQuery := s.fraudProtectionDecisionRecordsQuery()
	opts.ActivityTypes = []string{string(nonblocking.FraudProtectionDecisionRecorded)}
	baseQuery = opts.Apply(baseQuery)

	totalQuery := s.SQLBuilder.
		Select(
			"COUNT(*) AS total_actions",
			"COUNT(*) FILTER (WHERE decision = 'allowed' AND warning_count = 0) AS allowed_actions",
			"COUNT(*) FILTER (WHERE decision = 'blocked') AS blocked_actions",
			"COUNT(*) FILTER (WHERE decision = 'allowed' AND warning_count > 0) AS warning_actions",
		).
		FromSelect(baseQuery, "records")

	row, err := s.SQLExecutor.QueryRowWith(ctx, totalQuery)
	if err != nil {
		return nil, err
	}

	var totalActions int64
	var allowedActions int64
	var blockedActions int64
	var warnedActions int64
	if err := row.Scan(&totalActions, &allowedActions, &blockedActions, &warnedActions); err != nil {
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
		OrderBy("COUNT(*) DESC", "ip_address ASC").
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
		item.TotalActions = int(total)
		item.BlockedActions = int(blocked)
		item.WarnedActions = int(warnings)
		topSourceIPs = append(topSourceIPs, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &FraudProtectionOverview{
		TotalActions:   int(totalActions),
		AllowedActions: int(allowedActions),
		BlockedActions: int(blockedActions),
		WarnedActions: int(warnedActions),
		TopSourceIPs:   topSourceIPs,
	}, nil
}
