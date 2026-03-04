package fraudprotection

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	goredis "github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

const (
	thresholdCacheTTL         = 5 * time.Minute
	metricsNameSMSOTPVerified = "sms_otp_verified"
	auditMetricsTable         = "_audit_metrics"
)

type MetricsStore struct {
	AuditWriteDatabase *auditdb.WriteHandle
	AuditReadDatabase  *auditdb.ReadHandle
	SQLBuilder         *auditdb.SQLBuilderApp
	WriteSQLExecutor   *auditdb.WriteSQLExecutor
	ReadSQLExecutor    *auditdb.ReadSQLExecutor
	Redis              *appredis.Handle
	AppID              config.AppID
	Clock              clock.Clock
}

// RecordVerified inserts 2 rows into _audit_metrics in a single statement —
// one for the IP dimension and one for the phone country dimension.
// Called after an OTP is verified (fire-and-forget; caller ignores the returned error).
func (s *MetricsStore) RecordVerified(ctx context.Context, ip, phoneCountry string) error {
	now := s.Clock.NowUTC()
	ipKey := fmt.Sprintf("ip:%s", ip)
	countryKey := fmt.Sprintf("phone_country:%s", phoneCountry)
	id1 := uuid.New()
	id2 := uuid.New()

	tableName := s.SQLBuilder.TableName(auditMetricsTable)
	builder := s.SQLBuilder.
		Insert(tableName).
		Columns("id", "name", "key", "created_at").
		Values(id1, metricsNameSMSOTPVerified, ipKey, now).
		Values(id2, metricsNameSMSOTPVerified, countryKey, now)

	return s.AuditWriteDatabase.WithTx(ctx, func(ctx context.Context) error {
		_, err := s.WriteSQLExecutor.ExecWith(ctx, builder)
		return err
	})
}

// GetVerifiedByCountry24h returns the number of verified OTPs for a phone country
// in the past 24 hours. Result cached in Redis for 5 minutes.
func (s *MetricsStore) GetVerifiedByCountry24h(ctx context.Context, country string) (int64, error) {
	pgKey := fmt.Sprintf("phone_country:%s", country)
	since := s.Clock.NowUTC().Add(-24 * time.Hour)
	return s.queryVerifiedCount(ctx, pgKey, "24h", since)
}

// GetVerifiedByCountry1h returns the number of verified OTPs for a phone country
// in the past 1 hour. Result cached in Redis for 5 minutes.
func (s *MetricsStore) GetVerifiedByCountry1h(ctx context.Context, country string) (int64, error) {
	pgKey := fmt.Sprintf("phone_country:%s", country)
	since := s.Clock.NowUTC().Add(-1 * time.Hour)
	return s.queryVerifiedCount(ctx, pgKey, "1h", since)
}

// GetVerifiedByIP24h returns the number of verified OTPs from a specific IP
// in the past 24 hours. Result cached in Redis for 5 minutes.
func (s *MetricsStore) GetVerifiedByIP24h(ctx context.Context, ip string) (int64, error) {
	pgKey := fmt.Sprintf("ip:%s", ip)
	since := s.Clock.NowUTC().Add(-24 * time.Hour)
	return s.queryVerifiedCount(ctx, pgKey, "24h", since)
}

// GetVerifiedByCountryPast14DaysRollingMax returns the maximum single-day verified count
// for a phone country across the past 14 days. Result cached in Redis for 5 minutes.
func (s *MetricsStore) GetVerifiedByCountryPast14DaysRollingMax(ctx context.Context, country string) (int64, error) {
	pgKey := fmt.Sprintf("phone_country:%s", country)
	cacheKey := s.thresholdCacheKey(pgKey, "14d_max")

	cached, err := s.getCachedCount(ctx, cacheKey)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, goredis.Nil) {
		return 0, err
	}

	// Cache miss — query PostgreSQL.
	since := s.Clock.NowUTC().Add(-14 * 24 * time.Hour)
	tableName := s.SQLBuilder.TableName(auditMetricsTable)
	appID := string(s.AppID)

	subquery := sq.Select("DATE_TRUNC('day', created_at) AS day", "COUNT(*) AS daily_count").
		From(tableName).
		Where("app_id = ?", appID).
		Where("name = ?", metricsNameSMSOTPVerified).
		Where("key = ?", pgKey).
		Where("created_at >= ?", since).
		GroupBy("DATE_TRUNC('day', created_at)").
		PlaceholderFormat(sq.Dollar)

	query := sq.Select("COALESCE(MAX(daily_count), 0)").
		FromSelect(subquery, "t").
		PlaceholderFormat(sq.Dollar)

	var maxCount int64
	err = s.AuditReadDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		row, err := s.ReadSQLExecutor.QueryRowWith(ctx, query)
		if err != nil {
			return err
		}
		return row.Scan(&maxCount)
	})
	if err != nil {
		return 0, err
	}

	_ = s.setCachedCount(ctx, cacheKey, maxCount)
	return maxCount, nil
}

func (s *MetricsStore) queryVerifiedCount(ctx context.Context, pgKey string, window string, since time.Time) (int64, error) {
	cacheKey := s.thresholdCacheKey(pgKey, window)

	cached, err := s.getCachedCount(ctx, cacheKey)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, goredis.Nil) {
		return 0, err
	}

	// Cache miss — query PostgreSQL.
	var count int64
	err = s.AuditReadDatabase.ReadOnly(ctx, func(ctx context.Context) error {
		query := s.SQLBuilder.
			Select("COUNT(*)").
			From(s.SQLBuilder.TableName(auditMetricsTable)).
			Where("name = ?", metricsNameSMSOTPVerified).
			Where("key = ?", pgKey).
			Where("created_at >= ?", since)

		row, err := s.ReadSQLExecutor.QueryRowWith(ctx, query)
		if err != nil {
			return err
		}
		return row.Scan(&count)
	})
	if err != nil {
		return 0, err
	}

	_ = s.setCachedCount(ctx, cacheKey, count)
	return count, nil
}

func (s *MetricsStore) getCachedCount(ctx context.Context, cacheKey string) (int64, error) {
	var count int64
	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		val, err := conn.Get(ctx, cacheKey).Result()
		if err != nil {
			return err
		}
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		count = n
		return nil
	})
	return count, err
}

func (s *MetricsStore) setCachedCount(ctx context.Context, cacheKey string, count int64) error {
	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return conn.Set(ctx, cacheKey, strconv.FormatInt(count, 10), thresholdCacheTTL).Err()
	})
}

func (s *MetricsStore) thresholdCacheKey(pgKey string, window string) string {
	return fmt.Sprintf("app:%s:fraud_protection:threshold_cache:sms_otp_verified:%s:%s", string(s.AppID), window, pgKey)
}
