package cmdpricing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/client"

	"github.com/authgear/authgear-server/pkg/lib/config/configsource"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/libstripe"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("stripe")} }

func NewClientAPI(stripeConfig *portalconfig.StripeConfig, logger Logger) *client.API {
	clientAPI := &client.API{}
	clientAPI.Init(stripeConfig.SecretKey, &stripe.Backends{
		API: stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
			LeveledLogger: logger,
		}),
	})
	return clientAPI
}

var metadataSMSNorthAmerica = map[string]string{
	libstripe.MetadataKeyPriceType: string(model.PriceTypeUsage),
	libstripe.MetadataKeyUsageType: string(model.UsageTypeSMS),
	libstripe.MetadatakeySMSRegion: string(model.SMSRegionNorthAmerica),
}

var metadataSMSOtherRegions = map[string]string{
	libstripe.MetadataKeyPriceType: string(model.PriceTypeUsage),
	libstripe.MetadataKeyUsageType: string(model.UsageTypeSMS),
	libstripe.MetadatakeySMSRegion: string(model.SMSRegionOtherRegions),
}

type StripeService struct {
	ClientAPI   *client.API
	Handle      *globaldb.Handle
	SQLBuilder  *globaldb.SQLBuilder
	SQLExecutor *globaldb.SQLExecutor
	Store       *configsource.Store
	Clock       clock.Clock
	Logger      Logger
}

func (s *StripeService) ListAppIDs() (appIDs []string, err error) {
	err = s.Handle.ReadOnly(func() error {
		srcs, err := s.Store.ListAll()
		if err != nil {
			return err
		}

		for _, src := range srcs {
			appIDs = append(appIDs, src.AppID)
		}

		return nil
	})
	if err != nil {
		return
	}
	return
}

func (s *StripeService) getStripeSubscriptionID(appID string) (stripeSubscriptionID string, err error) {
	err = s.Handle.ReadOnly(func() error {
		q := s.SQLBuilder.Select(
			"stripe_subscription_id",
		).
			From(s.SQLBuilder.TableName("_portal_subscription")).
			Where("app_id = ?", appID)

		row, err := s.SQLExecutor.QueryRowWith(q)
		if err != nil {
			return err
		}

		err = row.Scan(
			&stripeSubscriptionID,
		)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *StripeService) scanUsageRecord(scanner db.Scanner) (*usage.UsageRecord, error) {
	var r usage.UsageRecord

	err := scanner.Scan(
		&r.ID,
		&r.AppID,
		&r.Name,
		&r.Period,
		&r.StartTime,
		&r.EndTime,
		&r.Count,
		&r.StripeTimestamp,
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *StripeService) getUsageRecordsForUpload(appID string, recordName usage.RecordName, subscriptionCreatedAt time.Time, midnight time.Time) (out []*usage.UsageRecord, err error) {
	err = s.Handle.ReadOnly(func() (err error) {
		q := s.SQLBuilder.Select(
			"id",
			"app_id",
			"name",
			"period",
			"start_time",
			"end_time",
			"count",
			"stripe_timestamp",
		).
			From(s.SQLBuilder.TableName("_portal_usage_record")).
			// We have two conditions here.
			// 1st condition is to retrieve usage records that have not been uploaded.
			// 2nd condition is to retrieve usage records that have been uploaded the same day, so that if this command
			// is ever re-run, we still sum up to a correct quantity.
			Where(
				"app_id = ? AND name = ? AND period = ? AND start_time > ? AND ((stripe_timestamp IS NULL AND end_time <= ?) OR stripe_timestamp IS NOT NULL AND stripe_timestamp = ?)",
				appID,
				string(recordName),
				string(periodical.Daily),
				subscriptionCreatedAt,
				midnight,
				midnight,
			)

		rows, err := s.SQLExecutor.QueryWith(q)
		if err != nil {
			return
		}
		defer rows.Close()

		for rows.Next() {
			var r *usage.UsageRecord
			r, err = s.scanUsageRecord(rows)
			if err != nil {
				return
			}
			out = append(out, r)
		}

		return
	})
	return
}

func (s *StripeService) markStripeTimestamp(usageRecords []*usage.UsageRecord, midnight time.Time) (err error) {
	err = s.Handle.WithTx(func() (err error) {
		var ids []string
		for _, record := range usageRecords {
			ids = append(ids, record.ID)
		}

		q := s.SQLBuilder.Update(
			s.SQLBuilder.TableName("_portal_usage_record"),
		).
			Set("stripe_timestamp", midnight).
			Where("id = ANY (?)", pq.Array(ids))

		result, err := s.SQLExecutor.ExecWith(q)
		if err != nil {
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return
		}

		if rowsAffected != int64(len(ids)) {
			err = fmt.Errorf("failed to update usage records: %v", ids)
			return
		}

		return
	})
	return
}

func (s *StripeService) uploadUsageRecordToSubscriptionItem(
	ctx context.Context,
	appID string,
	subscription *stripe.Subscription,
	si *stripe.SubscriptionItem,
	recordName usage.RecordName,
	midnight time.Time,
) (err error) {
	subscriptionCreatedAt := time.Unix(subscription.Created, 0)
	records, err := s.getUsageRecordsForUpload(
		appID,
		recordName,
		subscriptionCreatedAt,
		midnight,
	)
	if err != nil {
		return
	}

	// Skip when there are no records.
	if len(records) <= 0 {
		return
	}

	quantity := 0
	for _, record := range records {
		quantity += record.Count
	}

	timestamp := midnight.Unix()
	_, err = s.ClientAPI.UsageRecords.New(&stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(si.ID),
		Action:           stripe.String(stripe.UsageRecordActionSet),
		Quantity:         stripe.Int64(int64(quantity)),
		Timestamp:        &timestamp,
	})
	if err != nil {
		return
	}

	err = s.markStripeTimestamp(records, midnight)
	if err != nil {
		return
	}

	return
}

func (s *StripeService) getStripeSubscription(ctx context.Context, stripeSubscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
			Expand:  []*string{stripe.String("items.data.price.product")},
		},
	}
	return s.ClientAPI.Subscriptions.Get(stripeSubscriptionID, params)
}

func (s *StripeService) findSubscriptionItem(subscription *stripe.Subscription, metadata map[string]string) (*stripe.SubscriptionItem, bool) {
Loop:
	for _, subscriptionItem := range subscription.Items.Data {
		product := subscriptionItem.Price.Product

		// Find SubscriptionItem by checking if the metadata of the product is a superset of input metadata.
		for key, value := range metadata {
			v := product.Metadata[key]
			if value != v {
				continue Loop
			}
		}

		si := subscriptionItem
		return si, true
	}
	return nil, false
}

func (s *StripeService) uploadUsage(ctx context.Context, appID string) (err error) {
	midnight := timeutil.TruncateToDate(s.Clock.NowUTC())

	stripeSubscriptionID, err := s.getStripeSubscriptionID(appID)
	if errors.Is(err, sql.ErrNoRows) {
		s.Logger.Infof("%v: skip upload usage due to no subscription", appID)
		err = nil
		return
	}
	if err != nil {
		return
	}

	stripeSubscription, err := s.getStripeSubscription(ctx, stripeSubscriptionID)
	if err != nil {
		return
	}

	currentPeriodStart := time.Unix(
		stripeSubscription.CurrentPeriodStart,
		0,
	)

	if midnight.Before(currentPeriodStart) {
		s.Logger.Infof("%v: skip upload usage due to current_period_start", appID)
		return
	}

	smsNorthAmerica, ok := s.findSubscriptionItem(stripeSubscription, metadataSMSNorthAmerica)
	if !ok {
		err = fmt.Errorf("%v: subscription %v is missing SMS %v price", appID, stripeSubscription.ID, model.SMSRegionNorthAmerica)
		return
	}

	smsOtherRegions, ok := s.findSubscriptionItem(stripeSubscription, metadataSMSOtherRegions)
	if !ok {
		err = fmt.Errorf("%v: subscription %v is missing SMS %v price", appID, stripeSubscription.ID, model.SMSRegionOtherRegions)
		return
	}

	err = s.uploadUsageRecordToSubscriptionItem(
		ctx,
		appID,
		stripeSubscription,
		smsNorthAmerica,
		usage.RecordNameSMSSentNorthAmerica,
		midnight,
	)
	if err != nil {
		return
	}

	err = s.uploadUsageRecordToSubscriptionItem(
		ctx,
		appID,
		stripeSubscription,
		smsOtherRegions,
		usage.RecordNameSMSSentOtherRegions,
		midnight,
	)
	if err != nil {
		return
	}

	return
}

func (s *StripeService) UploadUsage(ctx context.Context, appID string) (err error) {
	err = s.uploadUsage(ctx, appID)
	if err != nil {
		s.Logger.WithError(err).Errorf("failed to upload usage for %v", appID)
	}
	return
}
