import React from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Text, DefaultEffects } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import {
  SubscriptionItemPriceSmsRegion,
  SubscriptionItemPriceType,
  SubscriptionItemPriceUsageType,
  SubscriptionPlan,
} from "./globalTypes.generated";
import { useSubscriptionPlansQueryQuery } from "./query/subscriptionPlansQuery";
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

const contactUsLink = "https://oursky.typeform.com/to/PecQiGfc";

const ALL_KNOWN_PLANS = ["free", "developers", "startups", "business"];
const DEFAULT_PLAN_NAME = ALL_KNOWN_PLANS[0];
const PAID_PLANS = ALL_KNOWN_PLANS.slice(0);

function previousPlan(planName: string): string | null {
  const idx = ALL_KNOWN_PLANS.indexOf(planName);
  if (idx >= 1) {
    return ALL_KNOWN_PLANS[idx - 1];
  }
  return null;
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
    (price) => price?.type === SubscriptionItemPriceType.Fixed
  );
  const northAmericaSMSPrice = subscriptionPlan.prices.find(
    (price) =>
      price?.type === SubscriptionItemPriceType.Usage &&
      price.usageType === SubscriptionItemPriceUsageType.Sms &&
      price.smsRegion === SubscriptionItemPriceSmsRegion.NorthAmerica
  );
  const otherRegionsSMSPrice = subscriptionPlan.prices.find(
    (price) =>
      price?.type === SubscriptionItemPriceType.Usage &&
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
      cta={<CTA variant="current" />}
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
}

function SubscriptionScreenContent(props: SubscriptionScreenContentProps) {
  const { planName, subscriptionPlans } = props;
  return (
    <div className={styles.root}>
      <ScreenTitle className={styles.section}>
        <FormattedMessage id="SubscriptionScreen.title" />
      </ScreenTitle>
      <SubscriptionCurrentPlanSummary
        className={styles.section}
        planName={planName}
      >
        <CostItem
          title={
            <FormattedMessage id="SubscriptionCurrentPlanSummary.total-cost.title" />
          }
          kind="non-applicable"
          tooltip={
            <FormattedMessage id="SubscriptionCurrentPlanSummary.total-cost.tooltip" />
          }
        />
        <CostItemSeparator />
        <CostItem
          title={
            <FormattedMessage id="SubscriptionCurrentPlanSummary.whatsapp.title" />
          }
          kind="non-applicable"
        />
        <CostItem
          title={
            <FormattedMessage id="SubscriptionCurrentPlanSummary.sms.title" />
          }
          kind="non-applicable"
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
              link: contactUsLink,
            }}
          />
        </Text>
        <Text block={true}>
          <FormattedMessage id="SubscriptionScreen.footer.pricing-details" />
        </Text>
      </div>
    </div>
  );
}

const SubscriptionScreen: React.FC = function SubscriptionScreen() {
  const { appID } = useParams() as { appID: string };
  const featureConfigQueryResult = useAppFeatureConfigQuery(appID);
  const subscriptionPlansQueryResult = useSubscriptionPlansQueryQuery();

  if (
    featureConfigQueryResult.loading ||
    subscriptionPlansQueryResult.loading
  ) {
    return <ShowLoading />;
  }

  if (featureConfigQueryResult.error || subscriptionPlansQueryResult.error) {
    return (
      <ShowError
        error={
          featureConfigQueryResult.error ?? subscriptionPlansQueryResult.error
        }
        onRetry={() => {
          featureConfigQueryResult.refetch().finally(() => {});
          subscriptionPlansQueryResult.refetch().finally(() => {});
        }}
      />
    );
  }

  const planName = featureConfigQueryResult.planName ?? DEFAULT_PLAN_NAME;
  const f = subscriptionPlansQueryResult.data?.subscriptionPlans ?? [];

  return (
    <SubscriptionScreenContent planName={planName} subscriptionPlans={f} />
  );
};

export default SubscriptionScreen;
