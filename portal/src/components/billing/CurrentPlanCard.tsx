import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Text as RadixText } from "@radix-ui/themes";
import { useNavigate } from "react-router-dom";
import styles from "./CurrentPlanCard.module.css";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import { TextButton } from "../v2/Button/TextButton/TextButton";
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
    <div className={styles.costItemRow}>
      <RadixText size="2" className={styles.costItemLabel}>
        {label}
      </RadixText>
      <RadixText size="2" className={styles.costItemValue}>
        {value}
      </RadixText>
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
      case "free2026":
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
      <div className={styles.sectionHeader}>
        <RadixText as="p" size="3" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="CurrentPlanCard.subscriptionFee.title" />
        </RadixText>
        {baseAmount != null ? (
          <div className={styles.priceRow}>
            <RadixText size="8">
              <FormattedMessage
                id="CurrentPlanCard.subscriptionFee.value"
                values={{ price: baseAmount.toLocaleString(locale) }}
              />
            </RadixText>
            <RadixText size="4" weight="medium" className={styles.priceUnit}>
              <FormattedMessage id="CurrentPlanCard.subscriptionFee.unit" />
            </RadixText>
          </div>
        ) : (
          <RadixText size="8">-</RadixText>
        )}
      </div>
      <div className={styles.detailsSection}>
        <RadixText as="p" size="2" weight="medium" className={styles.detailsSectionTitle}>
          <FormattedMessage id="CurrentPlanCard.subscriptionFee.include" />
        </RadixText>
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
      <div className={styles.sectionHeader}>
        <RadixText as="p" size="3" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.title" />
        </RadixText>
        <div className={styles.priceRow}>
          {totalCost != null ? (
            <>
              <RadixText size="8">
                <FormattedMessage
                  id="CurrentPlanCard.whatsappSMSFee.value"
                  values={{ price: totalCost.toLocaleString(locale) }}
                />
              </RadixText>
              <RadixText size="4" weight="medium" className={styles.priceUnit}>
                <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.unit" />
              </RadixText>
            </>
          ) : (
            <RadixText size="4" weight="medium">
              -
            </RadixText>
          )}
        </div>
      </div>
      <div className={styles.detailsSection}>
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
    <section className={cn(styles.card, styles["card--fullWidth"])}>
      <UsageMeter
        title={<FormattedMessage id="CurrentPlanCard.mau.title" />}
        current={mauCurrent}
        limit={mauLimit}
        previous={mauPrevious}
        warnPercentage={0.8}
        tooltip={
          <FormattedMessage
            id="CurrentPlanCard.mau.tooltip"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              br: () => <br />,
            }}
          />
        }
        onClickUpgrade={onUpgrade}
      />
    </section>
  );
}

interface UsageMeterProps {
  title: React.ReactNode;
  tooltip: React.ReactNode;
  current?: number;
  limit?: number;
  previous?: number;
  warnPercentage: number;
  onClickUpgrade?: React.MouseEventHandler<HTMLButtonElement>;
}

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
  const limitReached =
    current != null && limit != null ? current >= limit : false;

  return (
    <Tooltip content={tooltip}>
      <div className={styles.usageMeter}>
        <RadixText
          as="p"
          size="3"
          weight="medium"
          className={styles.usageMeterTitle}
        >
          {title}
        </RadixText>
        {percentComplete != null ? (
          <div className={styles.progressBar}>
            <div
              className={cn(
                styles.progressBarFill,
                limitReached ? styles["progressBarFill--warn"] : null
              )}
              style={{ width: `${Math.min(percentComplete, 1) * 100}%` }}
            />
          </div>
        ) : null}
        <RadixText
          as="p"
          size="2"
          className={cn(
            styles.usageText,
            limitReached ? styles["usageText--warn"] : null
          )}
        >
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
        </RadixText>
        {limitReached ? (
          <div className={styles.upgradeLink}>
            <TextButton
              variant="default"
              size="3"
              text={<FormattedMessage id="CurrentPlanCard.mau.limitReached" />}
              onClick={onClickUpgrade}
            />
          </div>
        ) : percentComplete != null && percentComplete >= warnPercentage ? (
          <div className={styles.upgradeLink}>
            <TextButton
              variant="default"
              size="3"
              text={
                <FormattedMessage id="CurrentPlanCard.mau.approachingLimit" />
              }
              onClick={onClickUpgrade}
            />
          </div>
        ) : null}
      </div>
    </Tooltip>
  );
}
