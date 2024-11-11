package cmdpricing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
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

var metadataWhatsappNorthAmerica = map[string]string{
	libstripe.MetadataKeyPriceType:      string(model.PriceTypeUsage),
	libstripe.MetadataKeyUsageType:      string(model.UsageTypeWhatsapp),
	libstripe.MetadatakeyWhatsappRegion: string(model.WhatsappRegionNorthAmerica),
}

var metadataWhatsappOtherRegions = map[string]string{
	libstripe.MetadataKeyPriceType:      string(model.PriceTypeUsage),
	libstripe.MetadataKeyUsageType:      string(model.UsageTypeWhatsapp),
	libstripe.MetadatakeyWhatsappRegion: string(model.WhatsappRegionOtherRegions),
}

var metadataMAU = map[string]string{
	libstripe.MetadataKeyPriceType: string(model.PriceTypeUsage),
	libstripe.MetadataKeyUsageType: string(model.UsageTypeMAU),
}

// SafeOffset is introduced to work around a weird bug of Stripe.
// When the subscription has its first invoice issued,
// It is an error to create an usage record of timestamp equal to current_period_start.
// So we add a 1-second offset to ensure the timestamp is WITHIN the current period.
const SafeOffset = 1 * time.Second

func ToStripeConvention(t time.Time) *int64 {
	unix := t.Unix()
	return &unix
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

func (s *StripeService) ListAppIDs(ctx context.Context) (appIDs []string, err error) {
	err = s.Handle.ReadOnly(ctx, func(ctx context.Context) error {
		srcs, err := s.Store.ListAll(ctx)
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

func (s *StripeService) getStripeSubscriptionID(ctx context.Context, appID string) (stripeSubscriptionID string, err error) {
	err = s.Handle.ReadOnly(ctx, func(ctx context.Context) error {
		q := s.SQLBuilder.Select(
			"stripe_subscription_id",
		).
			From(s.SQLBuilder.TableName("_portal_subscription")).
			Where("app_id = ?", appID)

		row, err := s.SQLExecutor.QueryRowWith(ctx, q)
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

func (s *StripeService) getUsageRecords(ctx context.Context, f func(builder squirrel.SelectBuilder) squirrel.SelectBuilder) (out []*usage.UsageRecord, err error) {
	err = s.Handle.ReadOnly(ctx, func(ctx context.Context) (err error) {
		q := f(s.SQLBuilder.Select(
			"id",
			"app_id",
			"name",
			"period",
			"start_time",
			"end_time",
			"count",
			"stripe_timestamp",
		).From(s.SQLBuilder.TableName("_portal_usage_record")))
		rows, err := s.SQLExecutor.QueryWith(ctx, q)
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

func (s *StripeService) getDailyUsageRecordsForUpload(
	ctx context.Context,
	appID string,
	recordName usage.RecordName,
	subscriptionCreatedAt time.Time,
	midnight time.Time,
	stripeTimestamp time.Time,
) (out []*usage.UsageRecord, err error) {
	return s.getUsageRecords(ctx, func(b squirrel.SelectBuilder) squirrel.SelectBuilder {
		// We have two conditions here.
		// 1st condition is to retrieve usage records that have not been uploaded.
		// 2nd condition is to retrieve usage records that have been uploaded the same day, so that if this command
		// is ever re-run, we still sum up to a correct quantity.
		return b.Where(
			"app_id = ? AND name = ? AND period = ? AND start_time > ? AND ((stripe_timestamp IS NULL AND end_time <= ?) OR stripe_timestamp IS NOT NULL AND stripe_timestamp = ?)",
			appID,
			string(recordName),
			string(periodical.Daily),
			subscriptionCreatedAt,
			midnight,
			stripeTimestamp,
		)
	})
}

func (s *StripeService) getMonthlyUsageRecordsForUpload(ctx context.Context, appID string, recordName usage.RecordName, currentPeriodEnd time.Time) (out []*usage.UsageRecord, err error) {
	return s.getUsageRecords(ctx, func(b squirrel.SelectBuilder) squirrel.SelectBuilder {
		return b.Where(
			"app_id = ? AND name = ? AND period = ? AND end_time = ?",
			appID,
			string(recordName),
			string(periodical.Monthly),
			currentPeriodEnd,
		)
	})
}

func (s *StripeService) markStripeTimestamp(ctx context.Context, usageRecords []*usage.UsageRecord, stripeTimestamp time.Time) (err error) {
	err = s.Handle.WithTx(ctx, func(ctx context.Context) (err error) {
		var ids []string
		for _, record := range usageRecords {
			ids = append(ids, record.ID)
		}

		q := s.SQLBuilder.Update(
			s.SQLBuilder.TableName("_portal_usage_record"),
		).
			Set("stripe_timestamp", stripeTimestamp).
			Where("id = ANY (?)", pq.Array(ids))

		result, err := s.SQLExecutor.ExecWith(ctx, q)
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

func (s *StripeService) uploadDailyUsageRecordToSubscriptionItem(
	ctx context.Context,
	appID string,
	subscription *stripe.Subscription,
	si *stripe.SubscriptionItem,
	recordName usage.RecordName,
	midnight time.Time,
) (err error) {
	timestamp := midnight.Add(SafeOffset)
	subscriptionCreatedAt := time.Unix(subscription.Created, 0).UTC()
	records, err := s.getDailyUsageRecordsForUpload(
		ctx,
		appID,
		recordName,
		subscriptionCreatedAt,
		midnight,
		timestamp,
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

	_, err = s.ClientAPI.UsageRecords.New(&stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(si.ID),
		Action:           stripe.String(stripe.UsageRecordActionSet),
		Quantity:         stripe.Int64(int64(quantity)),
		Timestamp:        ToStripeConvention(timestamp),
	})
	if err != nil {
		return
	}

	err = s.markStripeTimestamp(ctx, records, timestamp)
	if err != nil {
		return
	}

	return
}

func (s *StripeService) uploadMonthlyUsageRecordToSubscriptionItem(
	ctx context.Context,
	appID string,
	subscription *stripe.Subscription,
	si *stripe.SubscriptionItem,
	recordName usage.RecordName,
) (err error) {
	currentPeriodStart := time.Unix(
		subscription.CurrentPeriodStart,
		0,
	).UTC()
	currentPeriodEnd := time.Unix(
		subscription.CurrentPeriodEnd,
		0,
	).UTC()
	timestamp := currentPeriodStart.Add(SafeOffset)

	records, err := s.getMonthlyUsageRecordsForUpload(
		ctx,
		appID,
		recordName,
		currentPeriodEnd,
	)

	// Skip when there are no records.
	if len(records) <= 0 {
		return
	}

	quantity := 0
	for _, record := range records {
		quantity += record.Count
	}

	if freeQuantityStr, ok := si.Price.Metadata[libstripe.MetadataKeyFreeQuantity]; ok {
		var freeQuantity int
		freeQuantity, err = strconv.Atoi(freeQuantityStr)
		if err != nil {
			return fmt.Errorf("price %v has invalid free_quantity %#v: %w", si.Price.ID, freeQuantityStr, err)
		}

		quantity = quantity - freeQuantity
		if quantity < 0 {
			quantity = 0
		}
	}

	// We encounter this error
	// {"status":400,"message":"Cannot create the usage record with this timestamp because timestamps must be after the subscription's last invoice period (or current period start time).","param":"timestamp","request_id":"redacted","type":"invalid_request_error"}
	fields := map[string]interface{}{
		"app_id":               appID,
		"current_period_start": currentPeriodStart.Format(time.RFC3339),
		"current_period_end":   currentPeriodEnd.Format(time.RFC3339),
	}
	if subscription.LatestInvoice != nil {
		fields["latest_invoice_period_start"] = time.Unix(
			subscription.LatestInvoice.PeriodStart,
			0,
		).UTC().Format(time.RFC3339)
		fields["latest_invoice_period_end"] = time.Unix(
			subscription.LatestInvoice.PeriodEnd,
			0,
		).UTC().Format(time.RFC3339)
	}
	s.Logger.WithFields(fields).Infof("subscription timestamps")

	_, err = s.ClientAPI.UsageRecords.New(&stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(si.ID),
		Action:           stripe.String(stripe.UsageRecordActionSet),
		Quantity:         stripe.Int64(int64(quantity)),
		Timestamp:        ToStripeConvention(timestamp),
	})
	if err != nil {
		return
	}

	err = s.markStripeTimestamp(ctx, records, timestamp)
	if err != nil {
		return
	}

	return
}

func (s *StripeService) getStripeSubscription(ctx context.Context, stripeSubscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Params: stripe.Params{
			Context: ctx,
			Expand: []*string{
				stripe.String("items.data.price.product"),
				stripe.String("latest_invoice"),
			},
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

	stripeSubscriptionID, err := s.getStripeSubscriptionID(ctx, appID)
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
	).UTC()

	canUploadDailyUsage := !midnight.Before(currentPeriodStart)

	if canUploadDailyUsage {
		if smsNorthAmerica, ok := s.findSubscriptionItem(stripeSubscription, metadataSMSNorthAmerica); ok {
			err = s.uploadDailyUsageRecordToSubscriptionItem(
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
		}

		if smsOtherRegions, ok := s.findSubscriptionItem(stripeSubscription, metadataSMSOtherRegions); ok {
			err = s.uploadDailyUsageRecordToSubscriptionItem(
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
		}

		if whatsappNorthAmerica, ok := s.findSubscriptionItem(stripeSubscription, metadataWhatsappNorthAmerica); ok {
			err = s.uploadDailyUsageRecordToSubscriptionItem(
				ctx,
				appID,
				stripeSubscription,
				whatsappNorthAmerica,
				usage.RecordNameWhatsappSentNorthAmerica,
				midnight,
			)
			if err != nil {
				return
			}
		}

		if whatsappOtherRegions, ok := s.findSubscriptionItem(stripeSubscription, metadataWhatsappOtherRegions); ok {
			err = s.uploadDailyUsageRecordToSubscriptionItem(
				ctx,
				appID,
				stripeSubscription,
				whatsappOtherRegions,
				usage.RecordNameWhatsappSentOtherRegions,
				midnight,
			)
			if err != nil {
				return
			}
		}
	}

	if mau, ok := s.findSubscriptionItem(stripeSubscription, metadataMAU); ok {
		err = s.uploadMonthlyUsageRecordToSubscriptionItem(
			ctx,
			appID,
			stripeSubscription,
			mau,
			usage.RecordNameActiveUser,
		)
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
