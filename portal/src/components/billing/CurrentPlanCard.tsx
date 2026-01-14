import React, { useCallback, useContext, useMemo } from "react";
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
import {
  Context as MessageContext,
  FormattedMessage,
} from "../../intl";
import { useId } from "@fluentui/react-hooks";
import LinkButton from "../../LinkButton";
import {
  SMSCost,
  SMSUsage,
  WhatsappCost,
  WhatsappUsage,
  getMAULimit,
  getSMSCost,
  getSMSUsage,
  getWhatsappCost,
  getWhatsappUsage,
  isPlan,
  isStripePlan,
} from "../../util/plan";
import {
  SubscriptionItemPriceType,
  UsageType,
  SubscriptionUsage,
  Usage,
} from "../../graphql/portal/globalTypes.generated";
import { useNavigate } from "react-router-dom";

interface CurrentPlanCardProps {
  planName: string;
  thisMonthUsage: Usage | undefined;
  thisMonthSubscriptionUsage: SubscriptionUsage | undefined;
  previousMonthSubscriptionUsage: SubscriptionUsage | undefined;
  hasSubscription: boolean;
}

export function CurrentPlanCard({
  planName,
  thisMonthUsage,
  thisMonthSubscriptionUsage,
  previousMonthSubscriptionUsage,
  hasSubscription,
}: CurrentPlanCardProps): React.ReactElement {
  const baseAmount = useMemo(() => {
    if (!isStripePlan(planName)) {
      return undefined;
    }
    // show subscription fee only when subscription is active
    if (!hasSubscription) {
      return undefined;
    }
    const amountCent =
      thisMonthSubscriptionUsage?.items.find(
        (a) => a.type === SubscriptionItemPriceType.Fixed
      )?.unitAmount ?? undefined;
    if (amountCent == null) {
      return undefined;
    }
    return amountCent / 100;
  }, [planName, thisMonthSubscriptionUsage, hasSubscription]);

  const smsCost = useMemo(() => {
    if (thisMonthSubscriptionUsage == null) {
      return undefined;
    }
    // show sms cost only when subscription is active
    if (!hasSubscription) {
      return undefined;
    }
    return getSMSCost(planName, thisMonthSubscriptionUsage);
  }, [planName, thisMonthSubscriptionUsage, hasSubscription]);

  const smsUsage = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getSMSUsage(thisMonthUsage);
  }, [thisMonthUsage]);

  const whatsappCost = useMemo(() => {
    if (thisMonthSubscriptionUsage == null) {
      return undefined;
    }
    // show whatsapp cost only when subscription is active
    if (!hasSubscription) {
      return undefined;
    }
    return getWhatsappCost(planName, thisMonthSubscriptionUsage);
  }, [planName, thisMonthSubscriptionUsage, hasSubscription]);

  const whatsappUsage = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getWhatsappUsage(thisMonthUsage);
  }, [thisMonthUsage]);

  const mauCurrent = useMemo(() => {
    return thisMonthSubscriptionUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === UsageType.Mau
    )?.quantity;
  }, [thisMonthSubscriptionUsage]);

  const mauLimit = useMemo(() => {
    return getMAULimit(planName);
  }, [planName]);

  const mauPrevious = useMemo(() => {
    return previousMonthSubscriptionUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === UsageType.Mau
    )?.quantity;
  }, [previousMonthSubscriptionUsage]);

  return (
    <div className={styles.cardContainer}>
      <FixedCostSection planName={planName} baseAmount={baseAmount} />
      <MeteredCostSection
        smsCost={smsCost}
        smsUsage={smsUsage}
        whatsappCost={whatsappCost}
        whatsappUsage={whatsappUsage}
      />
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
  const { renderToString, locale } = useContext(MessageContext);
  const displayedPlanName = useMemo(() => {
    if (!isPlan(planName)) {
      return planName;
    }
    switch (planName) {
      case "free":
      case "free-approved":
        return renderToString("CurrentPlanCard.plan.free");
      case "developers":
      case "developers2025":
        return renderToString("CurrentPlanCard.plan.developers");
      case "business":
      case "business2025":
        return renderToString("CurrentPlanCard.plan.business");
      case "startups":
        return renderToString("CurrentPlanCard.plan.startups");
      case "enterprise":
        return renderToString("CurrentPlanCard.plan.enterprise");
    }
  }, [planName, renderToString]);

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
                values={{ price: baseAmount.toLocaleString(locale) }}
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
              values={{ plan: displayedPlanName }}
            />
          }
          value={
            baseAmount != null ? (
              <FormattedMessage
                id="CurrentPlanCard.subscriptionFee.planPrice"
                values={{ price: baseAmount.toLocaleString(locale) }}
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

function formatMessagePrice(locale: string, price: number) {
  return price.toLocaleString(locale, {
    minimumFractionDigits: 2,
  });
}

function MeteredCostSection({
  smsCost,
  smsUsage,
  whatsappCost,
  whatsappUsage,
}: {
  smsCost: SMSCost | undefined;
  smsUsage: SMSUsage | undefined;
  whatsappCost: WhatsappCost | undefined;
  whatsappUsage: WhatsappUsage | undefined;
}) {
  const { locale } = useContext(MessageContext);

  const totalCost = useMemo(() => {
    if (smsCost == null || whatsappCost == null) {
      return undefined;
    }
    return smsCost.totalCost + whatsappCost.totalCost;
  }, [smsCost, whatsappCost]);

  return (
    <section className={styles.card}>
      <div className="space-y-2">
        <Text block={true} variant="mediumPlus" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.title" />
        </Text>
        <div className="flex items-end">
          {totalCost != null ? (
            <>
              <Text variant="xxLarge">
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.value"
                  values={{ price: totalCost.toLocaleString(locale) }}
                />
              </Text>
              <Text variant="large" className="ml-2 font-semibold">
                <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.unit" />
              </Text>
            </>
          ) : (
            <Text variant="large" className="ml-2 font-semibold">
              -
            </Text>
          )}
        </div>
      </div>
      <div className="space-y-2">
        {smsCost != null || smsUsage != null ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.sms.northAmerica" />
            }
            value={
              smsCost != null ? (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                  values={{
                    unitPrice: formatMessagePrice(
                      locale,
                      smsCost.northAmericaUnitCost
                    ),
                    quantity: smsCost.northAmericaCount,
                    total: formatMessagePrice(
                      locale,
                      smsCost.northAmericaTotalCost
                    ),
                  }}
                />
              ) : (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSCount"
                  values={{
                    quantity: smsUsage!.northAmericaCount,
                  }}
                />
              )
            }
          />
        ) : null}
        {smsCost != null || smsUsage != null ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.sms.other" />
            }
            value={
              smsCost != null ? (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                  values={{
                    unitPrice: formatMessagePrice(
                      locale,
                      smsCost.otherRegionsUnitCost
                    ),
                    quantity: smsCost.otherRegionsCount,
                    total: formatMessagePrice(
                      locale,
                      smsCost.otherRegionsTotalCost
                    ),
                  }}
                />
              ) : (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSCount"
                  values={{
                    quantity: smsUsage!.otherRegionsCount,
                  }}
                />
              )
            }
          />
        ) : null}
        {whatsappCost != null || whatsappUsage != null ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.whatsapp.northAmerica" />
            }
            value={
              whatsappCost != null ? (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                  values={{
                    unitPrice: formatMessagePrice(
                      locale,
                      whatsappCost.northAmericaUnitCost
                    ),
                    quantity: whatsappCost.northAmericaCount,
                    total: formatMessagePrice(
                      locale,
                      whatsappCost.northAmericaTotalCost
                    ),
                  }}
                />
              ) : (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSCount"
                  values={{
                    quantity: whatsappUsage!.northAmericaCount,
                  }}
                />
              )
            }
          />
        ) : null}
        {whatsappCost != null || whatsappUsage != null ? (
          <CostItemRow
            label={
              <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.whatsapp.other" />
            }
            value={
              whatsappCost != null ? (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
                  values={{
                    unitPrice: formatMessagePrice(
                      locale,
                      whatsappCost.otherRegionsUnitCost
                    ),
                    quantity: whatsappCost.otherRegionsCount,
                    total: formatMessagePrice(
                      locale,
                      whatsappCost.otherRegionsTotalCost
                    ),
                  }}
                />
              ) : (
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.whatsappSMSCount"
                  values={{
                    quantity: whatsappUsage!.otherRegionsCount,
                  }}
                />
              )
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
  const navigate = useNavigate();
  const onUpgrade = useCallback(() => {
    navigate({ hash: "Subscription" });
  }, [navigate]);

  return (
    <section className={styles.card}>
      <UsageMeter
        title={<FormattedMessage id="CurrentPlanCard.mau.title" />}
        current={mauCurrent}
        limit={mauLimit}
        previous={mauPrevious}
        warnPercentage={0.8}
        tooltip={<FormattedMessage id="CurrentPlanCard.mau.tooltip" />}
        onClickUpgrade={onUpgrade}
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
    current != null && limit != null ? current / limit : null;
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
          {percentComplete != null ? (
            <ProgressIndicator
              className="w-full"
              percentComplete={percentComplete}
            />
          ) : null}
          <Text block={true} styles={usageStyles} variant="medium">
            {limit != null && current != null
              ? `${current} / ${limit}`
              : limit == null && current != null
              ? `${current}`
              : null}
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
          ) : percentComplete != null && percentComplete >= warnPercentage ? (
            <LinkButton onClick={onClickUpgrade}>
              <FormattedMessage id="CurrentPlanCard.mau.approachingLimit" />
            </LinkButton>
          ) : null}
        </ThemeProvider>
      </div>
    </TooltipHost>
  );
}
