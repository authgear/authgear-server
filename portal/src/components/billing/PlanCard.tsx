import React, { useCallback, useContext, useMemo } from "react";
import { Icon, Text } from "@fluentui/react";
import styles from "./PlanCard.module.css";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import { CTAVariant, getCTAVariant } from "../../util/plan";
import Tooltip from "../../Tooltip";
import { formatDateOnly } from "../../util/formatDateOnly";

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
  applications: number | "unlimited";
  projectMembers: number | "unlimited";
  logRetentionDays: number;
  support: string;
}

interface PlanAddOns {
  additionalMAU?: {
    price: number;
    unit: number;
  };
  perEnvironment?: number;
  perApplication?: number;
  perProjectMember?: number;
}

interface AdditionalFeature {
  iconName?: string;
  message: string;
}

interface BasePlanCardProps {
  planTitle: string;
  pricePerMonth: number | "free" | "custom";
  smsPricing: PlanCardSMSPricingFixed | PlanCardSMSPricingMetered;
  actionButtonMessage: string;
  actionButtonDisabled: boolean;
  onClickActionButton?: () => void;
  features: PlanFeatures;
  additionalFeatures?: AdditionalFeature[];
  addons?: PlanAddOns;
}

function BasePlanCard({
  planTitle,
  pricePerMonth,
  smsPricing,
  actionButtonMessage,
  actionButtonDisabled,
  onClickActionButton,
  features,
  additionalFeatures,
  addons,
}: BasePlanCardProps): React.ReactElement {
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <Text variant="mediumPlus" className="font-semibold">
          {planTitle}
        </Text>
        <PlanPrice pricePerMonth={pricePerMonth} />
      </div>
      {/* 32px(gap) + 40px(height of sms price) = 72 */}
      {/* This is to prevent layout bouncing caused by text wrapping */}
      <div className="pt-[72px] relative justify-self-stretch">
        {/* Use absolute to ensure height change of this block doesn't affect layout */}
        <div className="absolute top-0 left-0 right-0 text-center">
          <PlanSMSPrice smsPricing={smsPricing} />
        </div>
        <PrimaryButton
          className="w-full"
          text={actionButtonMessage}
          disabled={actionButtonDisabled}
          onClick={onClickActionButton}
        />
      </div>
      <FeatureList {...features} />
      {additionalFeatures != null ? (
        <>
          <div className="h-px w-full bg-separator" />
          <AdditionalFeatureList features={additionalFeatures} />
        </>
      ) : null}
      {addons != null ? (
        <>
          <div className="h-px w-full bg-separator" />
          <AddOnsList {...addons} />
        </>
      ) : null}
    </div>
  );
}

function PlanPrice({
  pricePerMonth,
}: {
  pricePerMonth: number | "free" | "custom";
}) {
  const { locale } = useContext(MessageContext);

  switch (pricePerMonth) {
    case "free":
      return (
        <Text variant="xxLarge">
          <FormattedMessage id="PlanCard.price.free" />
        </Text>
      );
    case "custom":
      return (
        <Text variant="large" className="font-semibold leading-9">
          <FormattedMessage id="PlanCard.price.custom" />
        </Text>
      );
    default:
      return (
        <div className="flex items-end">
          <Text variant="xxLarge">
            <FormattedMessage
              id="PlanCard.price.monthly.value"
              values={{
                price:
                  // Number formatting {n, number, integer} in message does not work
                  // So format manually
                  pricePerMonth.toLocaleString(locale),
              }}
            />
          </Text>
          <Text className="ml-2 font-semibold" variant="large">
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
        <Text variant="medium" className="font-semibold text-center">
          <FormattedMessage
            id="PlanCard.smsPrice.fixed"
            values={{ limit: smsPricing.limit }}
          />
        </Text>
      );
    case "metered":
      return (
        <div className="text-center">
          <Text variant="medium" className="font-semibold" block={true}>
            <FormattedMessage id="PlanCard.smsPrice.metered.title" />
          </Text>
          <Text variant="medium" className="text-text-secondary" block={true}>
            <FormattedMessage
              id="PlanCard.smsPrice.metered.price"
              values={{
                northAmericaPrice: smsPricing.northAmericaPrice,
                otherRegionPrice: smsPricing.otherRegionPrice,
              }}
            />
          </Text>
        </div>
      );
  }
}

function FeatureListItem({
  iconName,
  message,
}: {
  iconName?: string;
  message: React.ReactNode;
}) {
  return (
    <li className="flex items-center gap-2">
      {iconName != null ? (
        <Icon iconName={iconName} className="text-sm text-theme-primary" />
      ) : null}
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
  const { locale } = useContext(MessageContext);
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
                    mau.toLocaleString(locale)
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
              limit:
                typeof applications === "number"
                  ? // Number formatting {n, number, integer} in message does not work
                    // So format manually
                    applications.toLocaleString(locale)
                  : applications,
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
              limit:
                typeof projectMembers === "number"
                  ? // Number formatting {n, number, integer} in message does not work
                    // So format manually
                    projectMembers.toLocaleString(locale)
                  : projectMembers,
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

function AdditionalFeatureList({
  features,
}: {
  features: AdditionalFeature[];
}) {
  return (
    <ul className={styles.featureList}>
      {features.map((feature, idx) => {
        return (
          <FeatureListItem
            key={idx}
            iconName={feature.iconName}
            message={feature.message}
          />
        );
      })}
    </ul>
  );
}

function AddonListItem({
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

function AddOnsList({
  additionalMAU,
  perApplication,
  perEnvironment,
  perProjectMember,
}: PlanAddOns) {
  return (
    <ul className={styles.addonList}>
      <li className="flex items-center">
        <Text variant="medium" className="font-semibold">
          <FormattedMessage id="PlanCard.plan.addons.title" />
        </Text>
        <Tooltip
          tooltipMessageId="PlanCard.plan.addons.hint"
          className="text-sm"
        />
      </li>
      {additionalMAU != null ? (
        <AddonListItem
          iconName="Contact"
          message={
            <FormattedMessage
              id="PlanCard.plan.addons.additionalMAU"
              values={{ price: additionalMAU.price, unit: additionalMAU.unit }}
            />
          }
        />
      ) : null}
      {perEnvironment != null ? (
        <AddonListItem
          iconName="Picture"
          message={
            <FormattedMessage
              id="PlanCard.plan.addons.environment"
              values={{ price: perEnvironment }}
            />
          }
        />
      ) : null}
      {perApplication != null ? (
        <AddonListItem
          iconName="OEM"
          message={
            <FormattedMessage
              id="PlanCard.plan.addons.application"
              values={{ price: perApplication }}
            />
          }
        />
      ) : null}
      {perProjectMember != null ? (
        <AddonListItem
          iconName="People"
          message={
            <FormattedMessage
              id="PlanCard.plan.addons.projectMember"
              values={{ price: perProjectMember }}
            />
          }
        />
      ) : null}
    </ul>
  );
}

function useSubscriptablePlanCTAButton({
  cta,
  translatedPlanName,
  nextBillingDate,
}: {
  cta: CTAVariant;
  translatedPlanName: string;
  nextBillingDate?: Date;
}) {
  const { renderToString, locale } = useContext(MessageContext);

  const formattedBillingDate = useMemo(
    () => formatDateOnly(locale, nextBillingDate ?? null),
    [locale, nextBillingDate]
  );

  const isButtonActive = (() => {
    switch (cta) {
      case "contact-us":
      case "downgrade":
      case "reactivate":
      case "subscribe":
      case "upgrade":
        return true;
      default:
        return false;
    }
  })();

  const buttonText = useMemo(() => {
    switch (cta) {
      case "contact-us":
        return renderToString("PlanCard.action.contact-us");
      case "downgrade":
        return renderToString("PlanCard.action.downgrade", {
          plan: translatedPlanName,
        });
      case "reactivate":
        return renderToString("PlanCard.action.reactivate");
      case "subscribe":
        return renderToString("PlanCard.action.subscribe", {
          plan: translatedPlanName,
        });
      case "upgrade":
        return renderToString("PlanCard.action.upgrade", {
          plan: translatedPlanName,
        });
      case "current":
        return renderToString("PlanCard.action.current", {
          plan: translatedPlanName,
        });
      case "reactivate-to-downgrade":
        return renderToString("PlanCard.action.reactivate-to-downgrade");
      case "reactivate-to-upgrade":
        return renderToString("PlanCard.action.reactivate-to-upgrade");
      case "downgrading":
        return renderToString("PlanCard.action.downgrading", {
          plan: translatedPlanName,
          date: formattedBillingDate ?? "",
        });
      case "non-applicable":
        return renderToString("PlanCard.action.non-applicable");
    }
  }, [cta, renderToString, translatedPlanName, formattedBillingDate]);

  return {
    buttonText,
    isButtonActive,
  };
}

export interface PlanCardProps {
  currentPlan: string;
  subscriptionCancelled: boolean;
  onAction: (action: CTAVariant) => void;
}

export interface FreePlanCardProps extends PlanCardProps {
  nextBillingDate: Date | undefined;
}

export function PlanCardFree({
  currentPlan,
  subscriptionCancelled,
  nextBillingDate,
  onAction,
}: FreePlanCardProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const cta = getCTAVariant({
    cardPlanName: "free",
    currentPlanName: currentPlan,
    subscriptionCancelled,
  });

  const planNameTranslated = useMemo(() => {
    return renderToString("PlanCard.plan.free");
  }, [renderToString]);

  const { buttonText, isButtonActive } = useSubscriptablePlanCTAButton({
    cta,
    translatedPlanName: planNameTranslated,
    nextBillingDate,
  });

  const onClickActionButton = useCallback(() => {
    onAction(cta);
  }, [cta, onAction]);

  return (
    <BasePlanCard
      planTitle={renderToString("PlanCard.plan.free")}
      pricePerMonth={0}
      smsPricing={{
        type: "fixed",
        limit: 100,
      }}
      actionButtonMessage={buttonText}
      actionButtonDisabled={!isButtonActive}
      onClickActionButton={onClickActionButton}
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

export function PlanCardDevelopers({
  currentPlan,
  subscriptionCancelled,
  onAction,
}: PlanCardProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const cta = getCTAVariant({
    cardPlanName: "developers2025",
    currentPlanName: currentPlan,
    subscriptionCancelled,
  });

  const planNameTranslated = useMemo(() => {
    return renderToString("PlanCard.plan.developers");
  }, [renderToString]);

  const { buttonText, isButtonActive } = useSubscriptablePlanCTAButton({
    cta,
    translatedPlanName: planNameTranslated,
  });

  const onClickActionButton = useCallback(() => {
    onAction(cta);
  }, [cta, onAction]);

  return (
    <BasePlanCard
      planTitle={planNameTranslated}
      pricePerMonth={50}
      smsPricing={{
        type: "metered",
        northAmericaPrice: 0.02,
        otherRegionPrice: 0.1,
      }}
      actionButtonMessage={buttonText}
      actionButtonDisabled={!isButtonActive}
      onClickActionButton={onClickActionButton}
      features={{
        mau: "unlimited",
        applications: 2,
        projectMembers: 2,
        logRetentionDays: 1,
        support: renderToString("PlanCard.plan.features.support.email"),
      }}
      addons={{
        perEnvironment: 100,
        perApplication: 100,
        perProjectMember: 50,
      }}
    />
  );
}

export function PlanCardBusiness({
  currentPlan,
  subscriptionCancelled,
  onAction,
}: PlanCardProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const cta = getCTAVariant({
    cardPlanName: "business2025",
    currentPlanName: currentPlan,
    subscriptionCancelled,
  });

  const planNameTranslated = useMemo(() => {
    return renderToString("PlanCard.plan.business");
  }, [renderToString]);

  const { buttonText, isButtonActive } = useSubscriptablePlanCTAButton({
    cta,
    translatedPlanName: planNameTranslated,
  });

  const onClickActionButton = useCallback(() => {
    onAction(cta);
  }, [cta, onAction]);

  return (
    <BasePlanCard
      planTitle={planNameTranslated}
      pricePerMonth={500}
      smsPricing={{
        type: "metered",
        northAmericaPrice: 0.02,
        otherRegionPrice: 0.1,
      }}
      actionButtonMessage={buttonText}
      actionButtonDisabled={!isButtonActive}
      onClickActionButton={onClickActionButton}
      features={{
        mau: 25000,
        applications: 5,
        projectMembers: 5,
        logRetentionDays: 60,
        support: renderToString("PlanCard.plan.features.support.slack"),
      }}
      additionalFeatures={[
        {
          iconName: "CheckMark",
          message: renderToString(
            "PlanCard.plan.additionalFeature.removeAuthgearBranding"
          ),
        },
        {
          iconName: "CheckMark",
          message: renderToString(
            "PlanCard.plan.additionalFeature.projectMemberRoles"
          ),
        },
      ]}
      addons={{
        additionalMAU: {
          price: 50,
          unit: 5000,
        },
        perEnvironment: 100,
        perApplication: 100,
        perProjectMember: 50,
      }}
    />
  );
}

export function PlanCardEnterprise({
  currentPlan,
  subscriptionCancelled,
  onAction,
}: PlanCardProps): React.ReactElement {
  const { renderToString } = useContext(MessageContext);
  const cta = getCTAVariant({
    cardPlanName: "enterprise",
    currentPlanName: currentPlan,
    subscriptionCancelled,
  });

  const planNameTranslated = useMemo(() => {
    return renderToString("PlanCard.plan.enterprise");
  }, [renderToString]);

  const { buttonText, isButtonActive } = useSubscriptablePlanCTAButton({
    cta,
    translatedPlanName: planNameTranslated,
  });

  const onClickActionButton = useCallback(() => {
    onAction(cta);
  }, [cta, onAction]);

  return (
    <BasePlanCard
      planTitle={planNameTranslated}
      pricePerMonth="custom"
      smsPricing={{
        type: "metered",
        northAmericaPrice: 0.02,
        otherRegionPrice: 0.1,
      }}
      actionButtonMessage={buttonText}
      actionButtonDisabled={!isButtonActive}
      onClickActionButton={onClickActionButton}
      features={{
        mau: "custom",
        applications: "unlimited",
        projectMembers: "unlimited",
        logRetentionDays: 180,
        support: renderToString(
          "PlanCard.plan.features.support.dedicatedAccountManager"
        ),
      }}
      additionalFeatures={[
        {
          message: renderToString(
            "PlanCard.plan.additionalFeature.allFeaturesInBusiness"
          ),
        },
        {
          iconName: "CheckMark",
          message: renderToString(
            "PlanCard.plan.additionalFeature.customSMSGateway"
          ),
        },
        {
          iconName: "CheckMark",
          message: renderToString("PlanCard.plan.additionalFeature.customSMTP"),
        },
        {
          iconName: "CheckMark",
          message: renderToString(
            "PlanCard.plan.additionalFeature.tailoredSLA"
          ),
        },
        {
          iconName: "CheckMark",
          message: renderToString(
            "PlanCard.plan.additionalFeature.privateCloudOption"
          ),
        },
      ]}
      addons={{
        perEnvironment: 100,
      }}
    />
  );
}
