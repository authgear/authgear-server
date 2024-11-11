package cmdusage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/usage"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	libusage "github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/util/cobrasentry"
	"github.com/authgear/authgear-server/pkg/util/periodical"
)

var cmdUsage = &cobra.Command{
	Use:   "usage",
	Short: "Usage Commands",
}

func init() {
	binder := portalcmd.GetBinder()

	cmdUsage.AddCommand(cmdUsageCollectCount)
	binder.BindString(cmdUsageCollectCount.Flags(), portalcmd.ArgDatabaseURL)
	binder.BindString(cmdUsageCollectCount.Flags(), portalcmd.ArgDatabaseSchema)
	binder.BindString(cmdUsageCollectCount.Flags(), portalcmd.ArgAuditDatabaseURL)
	binder.BindString(cmdUsageCollectCount.Flags(), portalcmd.ArgAuditDatabaseSchema)
	binder.BindString(cmdUsageCollectCount.Flags(), portalcmd.ArgAnalyticRedisURL)
	binder.BindString(cmdUsageCollectCount.Flags(), cobrasentry.ArgSentryDSN)

	portalcmd.Root.AddCommand(cmdUsage)
}

var typeList = strings.Join([]string{
	string(libusage.RecordTypeActiveUser),
	string(libusage.RecordTypeSMSSent),
	string(libusage.RecordTypeEmailSent),
	string(libusage.RecordTypeWhatsappSent),
}, "|")

var cmdUsageCollectCount = &cobra.Command{
	Use:   fmt.Sprintf("collect-count [%s] [period]", typeList),
	Short: "Collect usage count record",
	Args:  cobra.ExactArgs(2),
	RunE: cobrasentry.RunEWrap(portalcmd.GetBinder, func(ctx context.Context, cmd *cobra.Command, args []string) (err error) {
		hub := cobrasentry.GetHub(ctx)
		logger := cobrasentry.NewLoggerFactory(hub).New("cmd-portal-usage")

		binder := portalcmd.GetBinder()
		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		dbCredentials := &config.DatabaseCredentials{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}

		auditDBURL, err := binder.GetRequiredString(cmd, portalcmd.ArgAuditDatabaseURL)
		if err != nil {
			return err
		}

		auditDBSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgAuditDatabaseSchema)
		if err != nil {
			return err
		}

		auditDBCredentials := &config.AuditDatabaseCredentials{
			DatabaseURL:    auditDBURL,
			DatabaseSchema: auditDBSchema,
		}

		var analyticRedisCredentials *config.AnalyticRedisCredentials
		analyticRedisURL := binder.GetString(cmd, portalcmd.ArgAnalyticRedisURL)
		if analyticRedisURL != "" {
			analyticRedisCredentials = &config.AnalyticRedisCredentials{
				RedisURL: analyticRedisURL,
			}
		}

		recordType := args[0]
		period := args[1]
		parser := analytic.NewPeriodicalArgumentParser()
		periodicalType, date, err := parser.Parse(period)
		if err != nil {
			return err
		}

		dbPool := db.NewPool()
		redisPool := redis.NewPool()
		countCollector := usage.NewCountCollector(
			dbPool,
			dbCredentials,
			auditDBCredentials,
			redisPool,
			analyticRedisCredentials,
			hub,
		)

		type collectorFuncType func(ctx context.Context, date *time.Time) (updatedCount int, err error)
		collectorFuncMap := map[libusage.RecordType]map[periodical.Type]collectorFuncType{
			libusage.RecordTypeActiveUser: {
				periodical.Monthly: countCollector.CollectMonthlyActiveUser,
				periodical.Weekly:  countCollector.CollectWeeklyActiveUser,
				periodical.Daily:   countCollector.CollectDailyActiveUser,
			},
			libusage.RecordTypeSMSSent: {
				periodical.Daily: countCollector.CollectDailySMSSent,
			},
			libusage.RecordTypeEmailSent: {
				periodical.Daily: countCollector.CollectDailyEmailSent,
			},
			libusage.RecordTypeWhatsappSent: {
				periodical.Daily: countCollector.CollectDailyWhatsappSent,
			},
		}

		collectorFunc, ok := collectorFuncMap[libusage.RecordType(recordType)][periodicalType]
		if !ok {
			return fmt.Errorf("invalid arguments; record type: %s; period: %s", recordType, period)
		}

		logger.Info("Start collecting usage records")

		updatedCount, err := collectorFunc(ctx, date)
		if err != nil {
			return err
		}

		logger.Infof("Number of records have been updated: %d", updatedCount)

		return nil
	}),
}
