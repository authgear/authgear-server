package service

import (
	"context"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/usage"
	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/util/periodical"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type UsageUsageStore interface {
	FetchUsageRecordsInRange(
		ctx context.Context,
		appID string,
		recordName usage.RecordName,
		period periodical.Type,
		fromStartTime time.Time,
		toEndTime time.Time,
	) ([]*usage.UsageRecord, error)
	FetchUsageRecords(
		ctx context.Context,
		appID string,
		recordName usage.RecordName,
		period periodical.Type,
		startTime time.Time,
	) ([]*usage.UsageRecord, error)
}

type UsageService struct {
	GlobalDatabase *globaldb.Handle
	UsageStore     UsageUsageStore
}

func (s *UsageService) GetUsage(
	ctx context.Context,
	appID string,
	date time.Time,
) (
	*model.Usage,
	error,
) {
	var usage *model.Usage
	var err error
	err = s.GlobalDatabase.WithTx(ctx, func(ctx context.Context) error {
		usage, err = s.getUsageWithTx(ctx, appID, date)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return usage, nil
}

func (s *UsageService) getUsageWithTx(
	ctx context.Context,
	appID string,
	date time.Time,
) (
	*model.Usage,
	error,
) {
	firstDayOfMonth := timeutil.FirstDayOfTheMonth(date)
	start := firstDayOfMonth
	end := start.AddDate(0, 1, 0)

	smsNorthAmericaRecords, err := s.UsageStore.FetchUsageRecordsInRange(
		ctx,
		appID,
		usage.RecordNameSMSSentNorthAmerica,
		periodical.Daily,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}
	smsNorthAmericaItem := &model.UsageItem{
		UsageType:      model.UsageTypeSMS,
		SMSRegion:      model.SMSRegionNorthAmerica,
		WhatsappRegion: model.WhatsappRegionNone,
		Quantity:       sumUsageRecord(smsNorthAmericaRecords),
	}

	smsOtherRegionRecords, err := s.UsageStore.FetchUsageRecordsInRange(
		ctx,
		appID,
		usage.RecordNameSMSSentOtherRegions,
		periodical.Daily,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}
	smsOtherRegionItem := &model.UsageItem{
		UsageType:      model.UsageTypeSMS,
		SMSRegion:      model.SMSRegionOtherRegions,
		WhatsappRegion: model.WhatsappRegionNone,
		Quantity:       sumUsageRecord(smsOtherRegionRecords),
	}

	whatsappNorthAmericaRecords, err := s.UsageStore.FetchUsageRecordsInRange(
		ctx,
		appID,
		usage.RecordNameWhatsappSentNorthAmerica,
		periodical.Daily,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}
	whatsappNorthAmericaItem := &model.UsageItem{
		UsageType:      model.UsageTypeWhatsapp,
		SMSRegion:      model.SMSRegionNone,
		WhatsappRegion: model.WhatsappRegionNorthAmerica,
		Quantity:       sumUsageRecord(whatsappNorthAmericaRecords),
	}

	whatsappOtherRegionRecords, err := s.UsageStore.FetchUsageRecordsInRange(
		ctx,
		appID,
		usage.RecordNameWhatsappSentOtherRegions,
		periodical.Daily,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}
	whatsappOtherRegionItem := &model.UsageItem{
		UsageType:      model.UsageTypeWhatsapp,
		SMSRegion:      model.SMSRegionNone,
		WhatsappRegion: model.WhatsappRegionOtherRegions,
		Quantity:       sumUsageRecord(whatsappOtherRegionRecords),
	}

	mauRecord, err := s.UsageStore.FetchUsageRecords(
		ctx,
		appID,
		usage.RecordNameActiveUser,
		periodical.Monthly,
		start,
	)
	if err != nil {
		return nil, err
	}
	mauItem := &model.UsageItem{
		UsageType:      model.UsageTypeMAU,
		SMSRegion:      model.SMSRegionNone,
		WhatsappRegion: model.WhatsappRegionNone,
		Quantity:       sumUsageRecord(mauRecord),
	}

	return &model.Usage{
		Items: []model.UsageItem{
			*smsNorthAmericaItem,
			*smsOtherRegionItem,
			*whatsappNorthAmericaItem,
			*whatsappOtherRegionItem,
			*mauItem,
		},
	}, nil
}
