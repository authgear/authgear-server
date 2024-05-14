package cmdanalytic

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/analytic"
	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	libanalytic "github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

var cmdAnalyticPosthog = &cobra.Command{
	Use:   "posthog",
	Short: "Commands for Posthog integration",
}

var cmdAnalyticPosthogGroup = &cobra.Command{
	Use:   "group",
	Short: "Set group properties",
	Args:  cobra.NoArgs,
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

		posthogEndpoint, err := binder.GetRequiredString(cmd, portalcmd.ArgPosthogEndpoint)
		if err != nil {
			return err
		}

		posthogAPIKey, err := binder.GetRequiredString(cmd, portalcmd.ArgPosthogAPIKey)
		if err != nil {
			return err
		}

		posthogCredentials := &libanalytic.PosthogCredentials{
			Endpoint: posthogEndpoint,
			APIKey:   posthogAPIKey,
		}

		dbPool := db.NewPool()
		redisPool := redis.NewPool()

		posthogIntegration := analytic.NewPosthogIntegration(
			context.Background(),
			dbPool,
			dbCredentials,
			auditDBCredentials,
			redisPool,
			analyticRedisCredentials,
			posthogCredentials,
		)

		err = posthogIntegration.SetGroupProperties()
		if err != nil {
			return err
		}

		return nil
	},
}

var cmdAnalyticPosthogUser = &cobra.Command{
	Use:   "user",
	Short: "Set user properties",
	Args:  cobra.NoArgs,
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

		posthogEndpoint, err := binder.GetRequiredString(cmd, portalcmd.ArgPosthogEndpoint)
		if err != nil {
			return err
		}

		posthogAPIKey, err := binder.GetRequiredString(cmd, portalcmd.ArgPosthogAPIKey)
		if err != nil {
			return err
		}

		portalAppID, err := binder.GetRequiredString(cmd, portalcmd.ArgAnalyticPortalAppID)
		if err != nil {
			return err
		}

		posthogCredentials := &libanalytic.PosthogCredentials{
			Endpoint: posthogEndpoint,
			APIKey:   posthogAPIKey,
		}

		dbPool := db.NewPool()
		redisPool := redis.NewPool()

		posthogIntegration := analytic.NewPosthogIntegration(
			context.Background(),
			dbPool,
			dbCredentials,
			auditDBCredentials,
			redisPool,
			analyticRedisCredentials,
			posthogCredentials,
		)

		err = posthogIntegration.SetUserProperties(portalAppID)
		if err != nil {
			return err
		}

		return nil
	},
}
