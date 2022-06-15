package cmdusage

import (
	"context"
	"fmt"
	"log"
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

	portalcmd.Root.AddCommand(cmdUsage)
}

var typeList = strings.Join([]string{
	string(libusage.RecordTypeActiveUser),
	string(libusage.RecordTypeSMSSent),
	string(libusage.RecordTypeEmailSent),
	string(libusage.RecordTypeWhatsappOTPVerified),
}, "|")

var cmdUsageCollectCount = &cobra.Command{
	Use:   fmt.Sprintf("collect-count [%s] [period]", typeList),
	Short: "Collect usage count record",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
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
			context.Background(),
			dbPool,
			dbCredentials,
			redisPool,
			analyticRedisCredentials,
		)

		type collectorFuncType func(date *time.Time) (updatedCount int, err error)
		collectorFuncMap := map[libusage.RecordType]map[periodical.Type]collectorFuncType{
			libusage.RecordTypeActiveUser: {
				periodical.Monthly: countCollector.CollectMonthlyActiveUser,
				periodical.Weekly:  countCollector.CollectWeeklyActiveUser,
				periodical.Daily:   countCollector.CollectDailyActiveUser,
			},
		}

		collectorFunc, ok := collectorFuncMap[libusage.RecordType(recordType)][periodicalType]
		if !ok {
			return fmt.Errorf("invalid arguments; record type: %s; period: %s", recordType, period)
		}

		log.Printf("Start collecting usage records")

		updatedCount, err := collectorFunc(date)
		if err != nil {
			return err
		}

		log.Printf("Number of records have been updated: %d", updatedCount)

		return nil
	},
}
