import React, { useMemo } from "react";
import {
  IButtonProps,
  ITooltipHostProps,
  PartialTheme,
  ProgressIndicator,
  Text,
  ThemeProvider,
  TooltipHost,
  useTheme,
} from "@fluentui/react";
import styles from "./CurrentPlanCard.module.css";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useId } from "@fluentui/react-hooks";
import LinkButton from "../../LinkButton";
import {
  SMSCost,
  WhatsappCost,
  getMAULimit,
  getSMSCost,
  getWhatsappCost,
  isStripePlan,
} from "../../util/plan";
import {
  SubscriptionItemPriceType,
  UsageType,
  SubscriptionUsage,
} from "../../graphql/portal/globalTypes.generated";

interface CurrentPlanCardProps {
  planName: string;
  thisMonthUsage: SubscriptionUsage | undefined;
  previousMonthUsage: SubscriptionUsage | undefined;
}

export function CurrentPlanCard({
  planName,
  thisMonthUsage,
  previousMonthUsage,
}: CurrentPlanCardProps): React.ReactElement {
  const baseAmount = useMemo(() => {
    if (!isStripePlan(planName)) {
      return undefined;
    }

    const amountCent =
      thisMonthUsage?.items.find(
        (a) => a.type === SubscriptionItemPriceType.Fixed
      )?.unitAmount ?? undefined;
    if (amountCent == null) {
      return undefined;
    }
    return amountCent / 100;
  }, [planName, thisMonthUsage]);

  const smsCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getSMSCost(planName, thisMonthUsage);
  }, [planName, thisMonthUsage]);

  const whatsappCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getWhatsappCost(planName, thisMonthUsage);
  }, [planName, thisMonthUsage]);

  const mauCurrent = useMemo(() => {
    return thisMonthUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === UsageType.Mau
    )?.quantity;
  }, [thisMonthUsage]);

  const mauLimit = useMemo(() => {
    return getMAULimit(planName);
  }, [planName]);

  const mauPrevious = useMemo(() => {
    return previousMonthUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === UsageType.Mau
    )?.quantity;
  }, [previousMonthUsage]);

  return (
    <div className={styles.cardContainer}>
      <FixedCostSection planName={planName} baseAmount={baseAmount} />
      <MeteredCostSection smsCost={smsCost} whatsappCost={whatsappCost} />
      <MAUUsageSection
        mauCurrent={mauCurrent}
        mauLimit={mauLimit}
        mauPrevious={mauPrevious}
      />
    </div>
  );
}

function CostItemRow({
  label,
  value,
}: {
  label: React.ReactNode;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-end justify-between">
      <Text variant="medium" className="font-semibold">
        {label}
      </Text>
      <Text variant="medium">{value}</Text>
    </div>
  );
}

function FixedCostSection({
  planName,
  baseAmount,
}: {
  planName: string;
  baseAmount: number | undefined;
}) {
  return (
    <section className={styles.card}>
      <div className="space-y-2">
        <Text block={true} variant="mediumPlus" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.subscriptionFee.title" />
        </Text>
        {baseAmount != null ? (
          <div className="flex items-end">
            <Text variant="xxLarge">
              <FormattedMessage
                id="CurrentPlanCard.subscriptionFee.value"
                values={{ price: baseAmount }}
              />
            </Text>
            <Text variant="large" className="ml-2 font-semibold">
              <FormattedMessage id="CurrentPlanCard.subscriptionFee.unit" />
            </Text>
          </div>
        ) : (
          <Text variant="xxLarge">-</Text>
        )}
      </div>
      <div className="space-y-2">
        <Text block={true} variant="medium" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.subscriptionFee.include" />
        </Text>
        <CostItemRow
          label={
            <FormattedMessage
              id="CurrentPlanCard.subscriptionFee.plan"
              values={{ plan: planName }}
            />
          }
          value={
            baseAmount != null ? (
              <FormattedMessage
                id="CurrentPlanCard.subscriptionFee.planPrice"
                values={{ price: baseAmount }}
              />
            ) : (
              "-"
            )
          }
        />
      </div>
    </section>
  );
}

function MeteredCostSection({
  smsCost,
  whatsappCost,
}: {
  smsCost: SMSCost | undefined;
  whatsappCost: WhatsappCost | undefined;
}) {
  const totalCost = useMemo(() => {
    return (smsCost?.totalCost ?? 0) + (whatsappCost?.totalCost ?? 0);
  }, [smsCost, whatsappCost]);

  return (
    <section className={styles.card}>
      <div className="space-y-2">
        <Text block={true} variant="mediumPlus" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.title" />
        </Text>
        <div className="flex items-end">
          <Text variant="xxLarge">
            <FormattedMessage
              id="CurrentPlanCard.whatsappSMSFee.value"
              values={{ price: totalCost }}
            />
          </Text>
          <Text variant="large" className="ml-2 font-semibold">
            <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.unit" />
          </Text>
        </div>
      </div>
      <div className="space-y-2">
        {smsCost ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.sms.northAmerica" />
            }
            value={
              <FormattedMessage
                id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                values={{
                  unitPrice: smsCost.northAmericaUnitCost,
                  quantity: smsCost.northAmericaCount,
                  total: smsCost.northAmericaTotalCost,
                }}
              />
            }
          />
        ) : null}
        {smsCost ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.sms.other" />
            }
            value={
              <FormattedMessage
                id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                values={{
                  unitPrice: smsCost.otherRegionsUnitCost,
                  quantity: smsCost.otherRegionsCount,
                  total: smsCost.otherRegionsTotalCost,
                }}
              />
            }
          />
        ) : null}
        {whatsappCost ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.whatsapp.northAmerica" />
            }
            value={
              <FormattedMessage
                id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                values={{
                  unitPrice: whatsappCost.northAmericaUnitCost,
                  quantity: whatsappCost.northAmericaCount,
                  total: whatsappCost.northAmericaTotalCost,
                }}
              />
            }
          />
        ) : null}
        {whatsappCost ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.whatsapp.other" />
            }
            value={
              <FormattedMessage
                id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                values={{
                  unitPrice: whatsappCost.otherRegionsUnitCost,
                  quantity: whatsappCost.otherRegionsCount,
                  total: whatsappCost.otherRegionsTotalCost,
                }}
              />
            }
          />
        ) : null}
      </div>
    </section>
  );
}

function MAUUsageSection({
  mauCurrent,
  mauLimit,
  mauPrevious,
}: {
  mauCurrent: number | undefined;
  mauLimit: number | undefined;
  mauPrevious: number | undefined;
}) {
  return (
    <section className={styles.card}>
      <UsageMeter
        title={<FormattedMessage id="CurrentPlanCard.mau.title" />}
        current={mauCurrent}
        limit={mauLimit}
        previous={mauPrevious}
        warnPercentage={0.8}
        tooltip={<FormattedMessage id="CurrentPlanCard.mau.tooltip" />}
      />
    </section>
  );
}

interface UsageMeterProps {
  title: React.ReactNode;
  tooltip: ITooltipHostProps["content"];
  current?: number;
  limit?: number;
  previous?: number;
  warnPercentage: number;
  onClickUpgrade?: IButtonProps["onClick"];
}

const USAGE_METER_THEME_WARN: PartialTheme = {
  palette: {
    themePrimary: "#F9597A",
  },
};

function UsageMeter(props: UsageMeterProps): React.ReactElement {
  const {
    title,
    tooltip,
    current,
    limit,
    previous,
    warnPercentage,
    onClickUpgrade,
  } = props;
  const percentComplete =
    current != null && limit != null ? current / limit : 0;
  const id = useId("usage-meter");
  const calloutProps = useMemo(() => {
    return {
      target: `#${id}`,
    };
  }, [id]);
  const currentTheme = useTheme();
  const limitReached =
    current != null && limit != null ? current >= limit : false;
  const theme = limitReached ? USAGE_METER_THEME_WARN : currentTheme;
  const usageStyles = {
    root: {
      color: limitReached
        ? USAGE_METER_THEME_WARN.palette?.themePrimary
        : currentTheme.palette.neutralSecondary,
    },
  };
  return (
    <TooltipHost
      hostClassName="col-span-2"
      content={tooltip}
      calloutProps={calloutProps}
    >
      <div className="flex flex-col">
        <Text
          id={id}
          block={true}
          variant="mediumPlus"
          className="self-start font-semibold mb-2"
        >
          {title}
        </Text>
        <ThemeProvider theme={theme}>
          <ProgressIndicator
            className="w-full"
            percentComplete={percentComplete}
          />
          <Text block={true} styles={usageStyles} variant="medium">
            {current != null ? `${current}` : "-"}
            {" / "}
            {limit != null ? `${limit}` : "-"}
            {previous != null ? (
              <FormattedMessage
                id="CurrentPlanCard.mau.previous"
                values={{
                  count: previous,
                }}
              />
            ) : null}
          </Text>
          {limitReached ? (
            <LinkButton onClick={onClickUpgrade}>
              <FormattedMessage id="CurrentPlanCard.mau.limitReached" />
            </LinkButton>
          ) : percentComplete >= warnPercentage ? (
            <LinkButton onClick={onClickUpgrade}>
              <FormattedMessage id="CurrentPlanCard.mau.approachingLimit" />
            </LinkButton>
          ) : null}
        </ThemeProvider>
      </div>
    </TooltipHost>
  );
}
