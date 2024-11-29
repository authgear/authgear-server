package service

import "github.com/authgear/authgear-server/pkg/lib/usage"

func sumUsageRecord(records []*usage.UsageRecord) int {
	sum := 0
	for _, record := range records {
		sum += record.Count
	}
	return sum
}
