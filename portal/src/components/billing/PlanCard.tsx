import React, { useCallback, useContext, useMemo } from "react";
import {
  CalendarIcon,
  CheckIcon,
  CubeIcon,
  GearIcon,
  IdCardIcon,
  ImageIcon,
  InfoCircledIcon,
  PersonIcon,
} from "@radix-ui/react-icons";
import { Text as RadixText } from "@radix-ui/themes";
import styles from "./PlanCard.module.css";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import { CTAVariant, DEFAULT_FREE_PLAN, getCTAVariant } from "../../util/plan";
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

type PlanCardIcon = typeof PersonIcon;

const PLAN_CARD_ICONS = {
  Contact: PersonIcon,
  OEM: CubeIcon,
  People: IdCardIcon,
  Calendar: CalendarIcon,
  Repair: GearIcon,
  CheckMark: CheckIcon,
  Picture: ImageIcon,
} satisfies Record<string, PlanCardIcon>;

interface AdditionalFeature {
  icon?: PlanCardIcon;
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
        <RadixText size="3" weight="medium" className={styles.planTitle}>
          {planTitle}
        </RadixText>
        <PlanPrice pricePerMonth={pricePerMonth} />
      </div>
      {/* 32px(gap) + 40px(height of sms price) = 72 */}
      {/* This is to prevent layout bouncing caused by text wrapping */}
      <div className={styles.smsPriceSection}>
        {/* Use absolute to ensure height change of this block doesn't affect layout */}
        <div className={styles.smsPriceAbsolute}>
          <PlanSMSPrice smsPricing={smsPricing} />
        </div>
        <div className={styles.actionButton}>
          <PrimaryButton
            size="2"
            text={actionButtonMessage}
            disabled={actionButtonDisabled}
            onClick={onClickActionButton}
          />
        </div>
      </div>
      <FeatureList {...features} />
      {additionalFeatures != null ? (
        <>
          <div className={styles.separator} />
          <AdditionalFeatureList features={additionalFeatures} />
        </>
      ) : null}
      {addons != null ? (
        <>
          <div className={styles.separator} />
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
        <RadixText size="8" className={styles.priceLarge}>
          <FormattedMessage id="PlanCard.price.free" />
        </RadixText>
      );
    case "custom":
      return (
        <RadixText size="4" weight="medium" className={styles.priceCustom}>
          <FormattedMessage id="PlanCard.price.custom" />
        </RadixText>
      );
    default:
      return (
        <div className={styles.priceMonthlyRow}>
          <RadixText size="8" className={styles.priceLarge}>
            <FormattedMessage
              id="PlanCard.price.monthly.value"
              values={{
                price:
                  // Number formatting {n, number, integer} in message does not work
                  // So format manually
                  pricePerMonth.toLocaleString(locale),
              }}
            />
          </RadixText>
          <RadixText size="4" weight="medium" className={styles.priceMonthlyUnit}>
            <FormattedMessage id="PlanCard.price.monthly.unit" />
          </RadixText>
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
        <RadixText size="2" align="center">
          <FormattedMessage
            id="PlanCard.smsPrice.fixed"
            values={{ limit: smsPricing.limit }}
          />
        </RadixText>
      );
    case "metered":
      return (
        <div className={styles.smsPriceMetered}>
          <RadixText size="2" as="p" className={styles.smsPriceTitle}>
            <FormattedMessage id="PlanCard.smsPrice.metered.title" />
          </RadixText>
          <RadixText
            size="2"
            as="p"
            className={styles.smsPriceSecondary}
          >
            <FormattedMessage
              id="PlanCard.smsPrice.metered.price"
              values={{
                northAmericaPrice: smsPricing.northAmericaPrice,
                otherRegionPrice: smsPricing.otherRegionPrice,
              }}
            />
          </RadixText>
        </div>
      );
  }
}

function FeatureListItem({
  icon: Icon,
  message,
}: {
  icon?: PlanCardIcon;
  message: React.ReactNode;
}) {
  return (
    <li className={styles.featureItem}>
      {Icon != null ? (
        <Icon className={styles.featureIcon} width="1rem" height="1rem" />
      ) : null}
      <RadixText size="2">
        {message}
      </RadixText>
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
        icon={PLAN_CARD_ICONS.Contact}
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
        icon={PLAN_CARD_ICONS.OEM}
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
        icon={PLAN_CARD_ICONS.People}
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
        icon={PLAN_CARD_ICONS.Calendar}
        message={
          <FormattedMessage
            id="PlanCard.plan.features.logRetentionDays"
            values={{
              limit: logRetentionDays.toFixed(0),
            }}
          />
        }
      />
      <FeatureListItem icon={PLAN_CARD_ICONS.Repair} message={support} />
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
            icon={feature.icon}
            message={feature.message}
          />
        );
      })}
    </ul>
  );
}

function AddonListItem({
  icon: Icon,
  message,
}: {
  icon: PlanCardIcon;
  message: React.ReactNode;
}) {
  return (
    <li className={styles.featureItem}>
      <Icon className={styles.featureIcon} width="1rem" height="1rem" />
      <RadixText size="2">
        {message}
      </RadixText>
    </li>
  );
}

function AddOnsList({
  additionalMAU,
  perApplication,
  perEnvironment,
  perProjectMember,
}: PlanAddOns) {
  const { renderToString } = useContext(MessageContext);

  return (
    <ul className={styles.addonList}>
      <li className={styles.addonTitleRow}>
        <RadixText size="2">
          <FormattedMessage id="PlanCard.plan.addons.title" />
        </RadixText>
        <Tooltip
          content={
            <FormattedMessage id="PlanCard.plan.addons.hint" />
          }
        >
          <InfoCircledIcon
            className={styles.addonHintIcon}
            width="1rem"
            height="1rem"
            aria-label={renderToString("PlanCard.plan.addons.hint")}
          />
        </Tooltip>
      </li>
      {additionalMAU != null ? (
        <AddonListItem
          icon={PLAN_CARD_ICONS.Contact}
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
          icon={PLAN_CARD_ICONS.Picture}
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
          icon={PLAN_CARD_ICONS.OEM}
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
          icon={PLAN_CARD_ICONS.People}
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
    cardPlanName: DEFAULT_FREE_PLAN,
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
          icon: PLAN_CARD_ICONS.CheckMark,
          message: renderToString(
            "PlanCard.plan.additionalFeature.removeAuthgearBranding"
          ),
        },
        {
          icon: PLAN_CARD_ICONS.CheckMark,
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
          icon: PLAN_CARD_ICONS.CheckMark,
          message: renderToString(
            "PlanCard.plan.additionalFeature.customSMSGateway"
          ),
        },
        {
          icon: PLAN_CARD_ICONS.CheckMark,
          message: renderToString("PlanCard.plan.additionalFeature.customSMTP"),
        },
        {
          icon: PLAN_CARD_ICONS.CheckMark,
          message: renderToString(
            "PlanCard.plan.additionalFeature.tailoredSLA"
          ),
        },
        {
          icon: PLAN_CARD_ICONS.CheckMark,
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
