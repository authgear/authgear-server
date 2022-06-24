import React, { useState, useCallback, useMemo } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { DateTime } from "luxon";
import {
  Text,
  DefaultEffects,
  PrimaryButton,
  Dialog,
  DialogType,
  DialogFooter,
  IDialogContentProps,
} from "@fluentui/react";
import { useConst } from "@fluentui/react-hooks";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  SubscriptionItemPriceSmsRegion,
  SubscriptionItemPriceType,
  SubscriptionItemPriceUsageType,
  SubscriptionPlan,
  SubscriptionUsage,
} from "./globalTypes.generated";
import { AppFragmentFragment } from "./query/subscriptionScreenQuery.generated";
import { useSubscriptionScreenQueryQuery } from "./query/subscriptionScreenQuery";
import styles from "./SubscriptionScreen.module.scss";
import SubscriptionCurrentPlanSummary, {
  CostItem,
  CostItemSeparator,
} from "./SubscriptionCurrentPlanSummary";
import SubscriptionPlanCard, {
  CardTag,
  CardTitle,
  CardTagline,
  BasePriceTag,
  MAURestriction,
  UsagePriceTag,
  CTA,
  PlanDetailsTitle,
  PlanDetailsLine,
} from "./SubscriptionPlanCard";

const ALL_KNOWN_PLANS = ["free", "developers", "startups", "business"];
const PAID_PLANS = ALL_KNOWN_PLANS.slice(1);

const MAU_LIMIT: Record<string, number> = {
  free: 5000,
  developers: 1000,
  startups: 5000,
  business: 30000,
};

function previousPlan(planName: string): string | null {
  const idx = ALL_KNOWN_PLANS.indexOf(planName);
  if (idx >= 1) {
    return ALL_KNOWN_PLANS[idx - 1];
  }
  return null;
}

function isKnownPlan(planName: string): boolean {
  return ALL_KNOWN_PLANS.indexOf(planName) >= 0;
}

function isKnownPaidPlan(planName: string): boolean {
  return PAID_PLANS.indexOf(planName) >= 0;
}

function isRecommendedPlan(planName: string): boolean {
  return planName === "startups";
}

function showRecommendedTag(
  planName: string,
  currentPlanName: string
): boolean {
  const a = isRecommendedPlan(planName);
  const i = ALL_KNOWN_PLANS.indexOf(planName);
  const j = ALL_KNOWN_PLANS.indexOf(currentPlanName);
  return a && i >= 0 && j >= 0 && j <= i;
}

interface PlanDetailsLinesProps {
  planName: string;
}

function PlanDetailsLines(props: PlanDetailsLinesProps) {
  const { planName } = props;
  const isKnown = isKnownPaidPlan(planName);
  if (!isKnown) {
    return null;
  }
  let length = 0;
  switch (planName) {
    case "developers":
      length = 5;
      break;
    case "startups":
      length = 4;
      break;
    case "business":
      length = 3;
      break;
  }
  const children = [];
  for (let i = 0; i < length; i++) {
    children.push(
      <PlanDetailsLine key={i}>
        <FormattedMessage
          id={`SubscriptionPlanCard.plan.features.line.${i}.${planName}`}
        />
      </PlanDetailsLine>
    );
  }
  return <>{children}</>;
}

interface SubscriptionPlanCardRenderProps {
  currentPlanName: string;
  subscriptionPlan: SubscriptionPlan;
}

function SubscriptionPlanCardRenderer(props: SubscriptionPlanCardRenderProps) {
  const { currentPlanName, subscriptionPlan } = props;
  const isKnown = isKnownPaidPlan(subscriptionPlan.name);
  if (!isKnown) {
    return null;
  }
  const { name } = subscriptionPlan;

  const basePrice = subscriptionPlan.prices.find(
    (price) => price.type === SubscriptionItemPriceType.Fixed
  );
  const northAmericaSMSPrice = subscriptionPlan.prices.find(
    (price) =>
      price.type === SubscriptionItemPriceType.Usage &&
      price.usageType === SubscriptionItemPriceUsageType.Sms &&
      price.smsRegion === SubscriptionItemPriceSmsRegion.NorthAmerica
  );
  const otherRegionsSMSPrice = subscriptionPlan.prices.find(
    (price) =>
      price.type === SubscriptionItemPriceType.Usage &&
      price.usageType === SubscriptionItemPriceUsageType.Sms &&
      price.smsRegion === SubscriptionItemPriceSmsRegion.OtherRegions
  );

  const previousPlanName = previousPlan(name);
  const cardTag = showRecommendedTag(name, currentPlanName) ? (
    <CardTag>
      <FormattedMessage id="SubscriptionScreen.recommended" />
    </CardTag>
  ) : null;

  return (
    <SubscriptionPlanCard
      isCurrentPlan={false}
      cardTag={cardTag}
      cardTitle={
        <CardTitle>
          <FormattedMessage id={"SubscriptionScreen.plan-name." + name} />
        </CardTitle>
      }
      cardTagline={
        <CardTagline>
          <FormattedMessage id={"SubscriptionPlanCard.plan.tagline." + name} />
        </CardTagline>
      }
      basePriceTag={
        <BasePriceTag>
          {basePrice != null ? `$${basePrice.unitAmount / 100}/mo` : "-"}
        </BasePriceTag>
      }
      mauRestriction={
        <MAURestriction>
          <FormattedMessage
            id={"SubscriptionPlanCard.plan.mau-restriction." + name}
          />
        </MAURestriction>
      }
      usagePriceTags={
        <>
          {northAmericaSMSPrice != null ? (
            <UsagePriceTag>
              <FormattedMessage
                id="SubscriptionPlanCard.sms.north-america"
                values={{
                  unitAmount: northAmericaSMSPrice.unitAmount / 100,
                }}
              />
            </UsagePriceTag>
          ) : null}
          {otherRegionsSMSPrice != null ? (
            <UsagePriceTag>
              <FormattedMessage
                id="SubscriptionPlanCard.sms.other-regions"
                values={{
                  unitAmount: otherRegionsSMSPrice.unitAmount / 100,
                }}
              />
            </UsagePriceTag>
          ) : null}
        </>
      }
      /* TODO(billing): determine the CTA */
      cta={<CTA variant="subscribe" />}
      planDetailsTitle={
        <PlanDetailsTitle>
          <FormattedMessage
            id="SubscriptionPlanCard.plan.features.title"
            values={{
              previousPlan: previousPlanName ?? "-",
            }}
          />
        </PlanDetailsTitle>
      }
      planDetailsLines={<PlanDetailsLines planName={name} />}
    />
  );
}

interface SubscriptionScreenContentProps {
  planName: string;
  subscriptionPlans: SubscriptionPlan[];
  thisMonthUsage?: SubscriptionUsage;
  previousMonthUsage?: SubscriptionUsage;
}

function getTotalCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): number | undefined {
  if (!isKnownPaidPlan(planName)) {
    return undefined;
  }

  let totalCost = 0;
  for (const item of subscriptionUsage.items) {
    totalCost += item.totalAmount ?? 0;
  }
  return totalCost;
}

interface SMSCost {
  totalCost: number;
  northAmericaCount: number;
  otherRegionsCount: number;
}

function getSMSCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): SMSCost | undefined {
  if (!isKnownPaidPlan(planName)) {
    return undefined;
  }

  const cost = {
    totalCost: 0,
    northAmericaCount: 0,
    otherRegionsCount: 0,
  };

  for (const item of subscriptionUsage.items) {
    if (
      item.type === SubscriptionItemPriceType.Usage &&
      item.usageType === SubscriptionItemPriceUsageType.Sms
    ) {
      cost.totalCost += item.totalAmount ?? 0;
      if (item.smsRegion === SubscriptionItemPriceSmsRegion.NorthAmerica) {
        cost.northAmericaCount = item.quantity;
      }
      if (item.smsRegion === SubscriptionItemPriceSmsRegion.OtherRegions) {
        cost.otherRegionsCount = item.quantity;
      }
    }
  }

  return cost;
}

function SubscriptionScreenContent(props: SubscriptionScreenContentProps) {
  const { planName, subscriptionPlans, thisMonthUsage, previousMonthUsage } =
    props;

  const totalCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getTotalCost(planName, thisMonthUsage);
  }, [planName, thisMonthUsage]);

  const smsCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getSMSCost(planName, thisMonthUsage);
  }, [planName, thisMonthUsage]);

  const baseAmount = useMemo(() => {
    if (!isKnownPaidPlan(planName)) {
      return undefined;
    }

    return (
      thisMonthUsage?.items.find(
        (a) => a.type === SubscriptionItemPriceType.Fixed
      )?.unitAmount ?? undefined
    );
  }, [planName, thisMonthUsage]);

  const mauCurrent = useMemo(() => {
    return thisMonthUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === SubscriptionItemPriceUsageType.Mau
    )?.quantity;
  }, [thisMonthUsage]);

  const mauLimit = useMemo(() => {
    if (!isKnownPlan(planName)) {
      return undefined;
    }

    return MAU_LIMIT[planName];
  }, [planName]);

  const mauPrevious = useMemo(() => {
    return previousMonthUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === SubscriptionItemPriceUsageType.Mau
    )?.quantity;
  }, [previousMonthUsage]);

  const nextBillingDate = useMemo(() => {
    if (!isKnownPaidPlan(planName)) {
      return undefined;
    }

    const nextBillingDate = thisMonthUsage?.nextBillingDate;
    if (nextBillingDate != null) {
      return new Date(nextBillingDate);
    }
    return undefined;
  }, [planName, thisMonthUsage]);

  const [hidden, setHidden] = useState(true);
  // @ts-expect-error
  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      type: DialogType.normal,
      title: <FormattedMessage id="SubscriptionScreen.enterprise.title" />,
      subText: (
        <FormattedMessage id="SubscriptionScreen.enterprise.instructions" />
      ),
    };
  }, []);
  const onClickEnterprisePlan = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setHidden(false);
  }, []);
  const onDismiss = useCallback(() => {
    setHidden(true);
  }, []);

  return (
    <>
      <Dialog
        hidden={hidden}
        onDismiss={onDismiss}
        dialogContentProps={dialogContentProps}
      >
        <DialogFooter>
          <PrimaryButton onClick={onDismiss}>
            <FormattedMessage id="SubscriptionScreen.enterprise.cta" />
          </PrimaryButton>
        </DialogFooter>
      </Dialog>

      <div className={styles.root}>
        <ScreenTitle className={styles.section}>
          <FormattedMessage id="SubscriptionScreen.title" />
        </ScreenTitle>
        <SubscriptionCurrentPlanSummary
          className={styles.section}
          planName={planName}
          baseAmount={baseAmount}
          mauCurrent={mauCurrent}
          mauLimit={mauLimit}
          mauPrevious={mauPrevious}
          nextBillingDate={nextBillingDate}
        >
          <CostItem
            title={
              <FormattedMessage id="SubscriptionCurrentPlanSummary.total-cost.title" />
            }
            kind={totalCost == null ? "non-applicable" : "billed"}
            amount={totalCost}
            tooltip={
              <FormattedMessage id="SubscriptionCurrentPlanSummary.total-cost.tooltip" />
            }
          />
          <CostItemSeparator />
          <CostItem
            title={
              <FormattedMessage id="SubscriptionCurrentPlanSummary.whatsapp.title" />
            }
            kind={
              isKnownPaidPlan(planName)
                ? "free"
                : planName === ALL_KNOWN_PLANS[0]
                ? "upgrade"
                : "non-applicable"
            }
          />
          <CostItem
            title={
              <FormattedMessage id="SubscriptionCurrentPlanSummary.sms.title" />
            }
            kind={smsCost == null ? "non-applicable" : "billed"}
            amount={smsCost == null ? undefined : smsCost.totalCost}
            tooltip={
              smsCost == null ? undefined : (
                <FormattedMessage
                  id="SubscriptionCurrentPlanSummary.sms.tooltip"
                  values={{
                    count1: smsCost.northAmericaCount,
                    count2: smsCost.otherRegionsCount,
                  }}
                />
              )
            }
          />
        </SubscriptionCurrentPlanSummary>
        <div
          className={cn(styles.section, styles.cardsContainer)}
          style={{
            boxShadow: DefaultEffects.elevation4,
          }}
        >
          <Text block={true} variant="xLarge">
            <FormattedMessage id="SubscriptionScreen.cards.title" />
          </Text>
          <div className={styles.cards}>
            {PAID_PLANS.map((planName) => {
              const plan = subscriptionPlans.find(
                (plan) => plan.name === planName
              );
              if (plan != null) {
                return (
                  <SubscriptionPlanCardRenderer
                    key={plan.name}
                    subscriptionPlan={plan}
                    currentPlanName={planName}
                  />
                );
              }
              return null;
            })}
          </div>
        </div>
        <div className={styles.footer}>
          <Text block={true}>
            <FormattedMessage
              id="SubscriptionScreen.footer.enterprise-plan"
              values={{
                onClick: onClickEnterprisePlan,
              }}
            />
          </Text>
          <Text block={true}>
            <FormattedMessage id="SubscriptionScreen.footer.pricing-details" />
          </Text>
        </div>
      </div>
    </>
  );
}

const SubscriptionScreen: React.FC = function SubscriptionScreen() {
  const now = useConst(new Date());
  const thisMonth = useMemo(() => {
    return now.toISOString();
  }, [now]);
  const previousMonth = useMemo(() => {
    return DateTime.fromJSDate(now)
      .minus({
        months: 1,
      })
      .toJSDate()
      .toISOString();
  }, [now]);

  const { appID } = useParams() as { appID: string };
  const subscriptionScreenQuery = useSubscriptionScreenQueryQuery({
    variables: {
      id: appID,
      thisMonth,
      previousMonth,
    },
  });

  if (subscriptionScreenQuery.loading) {
    return <ShowLoading />;
  }

  if (subscriptionScreenQuery.error) {
    return (
      <ShowError
        error={subscriptionScreenQuery.error}
        onRetry={() => {
          subscriptionScreenQuery.refetch().finally(() => {});
        }}
      />
    );
  }

  const planName = (subscriptionScreenQuery.data?.node as AppFragmentFragment)
    .planName;
  const subscriptionPlans =
    subscriptionScreenQuery.data?.subscriptionPlans ?? [];
  const thisMonthUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).thisMonth;
  const previousMonthUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).previousMonth;

  return (
    <SubscriptionScreenContent
      planName={planName}
      subscriptionPlans={subscriptionPlans}
      thisMonthUsage={thisMonthUsage ?? undefined}
      previousMonthUsage={previousMonthUsage ?? undefined}
    />
  );
};

export default SubscriptionScreen;
