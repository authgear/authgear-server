import React, { useContext } from "react";
import { Icon, Text } from "@fluentui/react";
import styles from "./PlanCard.module.css";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import PrimaryButton from "../../PrimaryButton";
import { comparePlan, isPlan } from "../../util/plan";

interface PlanCardSMSPricingFixed {
  type: "fixed";
  limit: number;
}

interface PlanCardSMSPricingMetered {
  type: "metered";
  northAmericaPrice: number;
  otherRegionPrice: number;
}

interface PlanFeatures {
  mau: number | "unlimited" | "custom";
  applications: number;
  projectMembers: number;
  logRetentionDays: number;
  support: string;
}

interface BasePlanCardProps {
  planTitle: string;
  pricePerMonth: number | "free" | "custom";
  smsPricing: PlanCardSMSPricingFixed | PlanCardSMSPricingMetered;
  subscribeButtonMessage: string;
  subscribeButtonDisabled: boolean;
  features: PlanFeatures;
}

function BasePlanCard({
  planTitle,
  pricePerMonth,
  smsPricing,
  subscribeButtonMessage,
  subscribeButtonDisabled,
  features,
}: BasePlanCardProps): React.ReactElement {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <Text variant="mediumPlus" className="font-semibold">
          {planTitle}
        </Text>
        <PlanPrice pricePerMonth={pricePerMonth} />
      </div>
      <PlanSMSPrice smsPricing={smsPricing} />
      <PrimaryButton
        className="w-full"
        text={subscribeButtonMessage}
        disabled={subscribeButtonDisabled}
      />
      <FeatureList {...features} />
    </div>
  );
}

function PlanPrice({
  pricePerMonth,
}: {
  pricePerMonth: number | "free" | "custom";
}) {
  switch (pricePerMonth) {
    case "free":
      return (
        <Text variant="xxLarge">
          <FormattedMessage id="PlanCard.price.free" />
        </Text>
      );
    case "custom":
      return (
        <Text>
          <FormattedMessage id="PlanCard.price.custom" />
        </Text>
      );
    default:
      return (
        <div className="flex">
          <Text>
            <FormattedMessage
              id="PlanCard.price.monthly.value"
              values={{ price: pricePerMonth }}
            />
          </Text>
          <Text className="ml-2">
            <FormattedMessage id="PlanCard.price.monthly.unit" />
          </Text>
        </div>
      );
  }
}

function PlanSMSPrice({
  smsPricing,
}: {
  smsPricing: PlanCardSMSPricingFixed | PlanCardSMSPricingMetered;
}) {
  switch (smsPricing.type) {
    case "fixed":
      return (
        <Text variant="medium" className="font-semibold">
          <FormattedMessage
            id="PlanCard.smsPrice.fixed"
            values={{ limit: smsPricing.limit }}
          />
        </Text>
      );
    case "metered":
      // TODO
      return <div></div>;
  }
}

function FeatureListItem({
  iconName,
  message,
}: {
  iconName: string;
  message: React.ReactNode;
}) {
  return (
    <li className="flex items-center gap-2">
      <Icon iconName={iconName} className="text-sm text-theme-primary" />
      <Text variant="medium" className="font-semibold">
        {message}
      </Text>
    </li>
  );
}

function FeatureList({
  mau,
  applications,
  projectMembers,
  logRetentionDays,
  support,
}: PlanFeatures) {
  return (
    <ul className={styles.featureList}>
      <FeatureListItem
        iconName="Contact"
        message={
          <FormattedMessage
            id="PlanCard.plan.features.mau"
            values={{
              limit:
                typeof mau === "number"
                  ? // Number formatting {n, number, integer} in message does not work
                    // So format manually
                    Intl.NumberFormat().format(mau)
                  : mau,
            }}
          />
        }
      />
      <FeatureListItem
        iconName="OEM"
        message={
          <FormattedMessage
            id="PlanCard.plan.features.applications"
            values={{
              limit: applications.toFixed(0),
            }}
          />
        }
      />
      <FeatureListItem
        iconName="People"
        message={
          <FormattedMessage
            id="PlanCard.plan.features.projectMembers"
            values={{
              limit: projectMembers.toFixed(0),
            }}
          />
        }
      />
      <FeatureListItem
        iconName="Calendar"
        message={
          <FormattedMessage
            id="PlanCard.plan.features.logRetentionDays"
            values={{
              limit: logRetentionDays.toFixed(0),
            }}
          />
        }
      />
      <FeatureListItem iconName="Repair" message={support} />
    </ul>
  );
}

export interface PlanCardProps {
  currentPlan: string;
}

export function PlanCardFree({
  currentPlan,
}: PlanCardProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);

  const isActive = isPlan(currentPlan)
    ? comparePlan(currentPlan, "free") === 0
    : false;

  return (
    <BasePlanCard
      planTitle={renderToString("PlanCard.plan.free")}
      pricePerMonth="free"
      smsPricing={{
        type: "fixed",
        limit: 100,
      }}
      subscribeButtonMessage={
        isActive
          ? renderToString("PlanCard.plan.free.active")
          : renderToString("PlanCard.plan.free.downgrade")
      }
      subscribeButtonDisabled={isActive}
      features={{
        mau: "unlimited",
        applications: 2,
        projectMembers: 2,
        logRetentionDays: 1,
        support: renderToString("PlanCard.plan.features.support.discord"),
      }}
    />
  );
}
