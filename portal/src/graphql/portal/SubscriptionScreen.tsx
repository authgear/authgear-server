/* global stripe */
import React, {
  useState,
  useCallback,
  useMemo,
  useContext,
  useEffect,
} from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { DateTime } from "luxon";
import {
  Text,
  DefaultEffects,
  Dialog,
  DialogType,
  DialogFooter,
  IDialogContentProps,
  ThemeProvider,
  PartialTheme,
  Spinner,
  SpinnerSize,
} from "@fluentui/react";
import { useConst } from "@fluentui/react-hooks";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  Subscription,
  SubscriptionItemPriceSmsRegion,
  SubscriptionItemPriceType,
  SubscriptionItemPriceUsageType,
  SubscriptionItemPriceWhatsappRegion,
  SubscriptionPlan,
  SubscriptionUsage,
} from "./globalTypes.generated";
import { PortalAPIAppConfig } from "../../types";
import { AppFragmentFragment } from "./query/subscriptionScreenQuery.generated";
import { useSubscriptionScreenQueryQuery } from "./query/subscriptionScreenQuery";
import { useGenerateStripeCustomerPortalSessionMutationMutation } from "./mutations/generateStripeCustomerPortalSessionMutation";
// import { useUpdateSubscriptionMutation } from "./mutations/updateSubscriptionMutation";
import styles from "./SubscriptionScreen.module.css";
import SubscriptionCurrentPlanSummary, {
  CostItem,
  CostItemSeparator,
} from "./SubscriptionCurrentPlanSummary";
// import SubscriptionPlanCard, {
//   CardTag,
//   CardTitle,
//   CardTagline,
//   BasePriceTag,
//   MAURestriction,
//   UsagePriceTag,
//   CTA,
//   PlanDetailsTitle,
//   PlanDetailsLine,
// } from "./SubscriptionPlanCard";
// import { useCreateCheckoutSessionMutation } from "./mutations/createCheckoutSessionMutation";
import { useLoading, useIsLoading } from "./../../hook/loading";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useSetSubscriptionCancelledStatusMutation } from "./mutations/setSubscriptionCancelledStatusMutation";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ErrorDialog from "../../error/ErrorDialog";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { useCancelFailedSubscriptionMutation } from "./mutations/cancelFailedSubscriptionMutation";
import ExternalLink from "../../ExternalLink";
import { SubscriptionScreenFooter } from "./SubscriptionScreenFooter";
import {
  isCustomPlan,
  isStripePlan,
  isFreePlan,
  isLimitedFreePlan,
  // getPreviousPlan,
  // getCTAVariant,
  getMAULimit,
  // shouldShowRecommendedTag,
  // SUBSCRIPTABLE_PLANS,
  // isPlan,
  // Plan,
} from "../../util/plan";
import {
  PlanCardBusiness,
  PlanCardDevelopers,
  PlanCardEnterprise,
  PlanCardFree,
} from "../../components/billing/PlanCard";

const CHECK_IS_PROCESSING_SUBSCRIPTION_INTERVAL = 5000;

const CONTACT_US_BUTTON_THEME: PartialTheme = {
  palette: {
    themePrimary: "#c8c8c8",
    neutralPrimary: "#c8c8c8",
  },
  semanticColors: {
    linkHovered: "#c8c8c8",
  },
};

function shouldShowFreePlanWarning(
  planName: string,
  effectiveAppConfig: PortalAPIAppConfig | undefined
): boolean {
  if (effectiveAppConfig == null) {
    return false;
  }

  if (!isLimitedFreePlan(planName)) {
    return false;
  }

  const loginIDEnabled =
    effectiveAppConfig.authentication?.identities?.includes("login_id") ??
    false;
  const phoneEnabled =
    effectiveAppConfig.identity?.login_id?.keys?.find(
      (a) => a.type === "phone"
    ) != null;
  const oobOTPSMSEnabled =
    effectiveAppConfig.authentication?.primary_authenticators?.includes(
      "oob_otp_sms"
    ) ?? false;

  return loginIDEnabled && phoneEnabled && oobOTPSMSEnabled;
}

// interface PlanDetailsLinesProps {
//   planName: string;
// }

// function PlanDetailsLines(props: PlanDetailsLinesProps) {
//   const { planName } = props;
//   let length: number;
//   switch (planName) {
//     case "startups":
//       length = 4;
//       break;
//     case "business":
//       length = 6;
//       break;
//     default:
//       length = 0;
//       break;
//   }
//   const children = [];
//   for (let i = 0; i < length; i++) {
//     children.push(
//       <PlanDetailsLine key={i}>
//         <FormattedMessage
//           id={`SubscriptionPlanCard.plan.features.line.${i}.${planName}`}
//         />
//       </PlanDetailsLine>
//     );
//   }
//   return <>{children}</>;
// }

// interface SubscriptionPlanCardRenderProps {
//   planName: string;
//   currentPlanName: string;
//   subscriptionCancelled: boolean;
//   subscriptionPlan: SubscriptionPlan;
//   nextBillingDate?: Date;
// }

// function SubscriptionPlanCardRenderer(props: SubscriptionPlanCardRenderProps) {
//   const { planName, currentPlanName, subscriptionCancelled } = props;
//   const { appID } = useParams() as { appID: string };
//   const { createCheckoutSession, loading: createCheckoutSessionLoading } =
//     useCreateCheckoutSessionMutation();
//   useLoading(createCheckoutSessionLoading);
//   const [updateSubscription, { loading: updateSubscriptionLoading }] =
//     useUpdateSubscriptionMutation();
//   useLoading(updateSubscriptionLoading);
//   const {
//     setSubscriptionCancelledStatus,
//     loading: reactivateSubscriptionLoading,
//     error: reactivateSubscriptionError,
//   } = useSetSubscriptionCancelledStatusMutation(appID);
//   useLoading(reactivateSubscriptionLoading);

//   const isLoading = useIsLoading();

//   const ctaVariant = useMemo(
//     () =>
//       getCTAVariant({
//         cardPlanName: planName,
//         currentPlanName,
//         subscriptionCancelled,
//       }),
//     [planName, currentPlanName, subscriptionCancelled]
//   );

// const onClickSubscribe = useCallback(
//   (planName: string) => {
//     createCheckoutSession(appID, planName)
//       .then((url) => {
//         if (url) {
//           window.location.href = url;
//         }
//       })
//       .finally(() => {});
//   },
//   [appID, createCheckoutSession]
// );

// const onClickUpgrade = useCallback(
//   (planName: string) => {
//     updateSubscription({
//       variables: {
//         appID,
//         planName,
//       },
//     }).finally(() => {});
//   },
//   [appID, updateSubscription]
// );

// const onClickDowngrade = useCallback(
//   (planName: string) => {
//     updateSubscription({
//       variables: {
//         appID,
//         planName,
//       },
//     }).finally(() => {});
//   },
//   [appID, updateSubscription]
// );

// const onClickReactivate = useCallback(async () => {
//   await setSubscriptionCancelledStatus(false);
// }, [setSubscriptionCancelledStatus]);

// const { name } = subscriptionPlan;

// const basePrice = subscriptionPlan.prices.find(
//   (price) => price.type === SubscriptionItemPriceType.Fixed
// );
// const northAmericaSMSPrice = subscriptionPlan.prices.find(
//   (price) =>
//     price.type === SubscriptionItemPriceType.Usage &&
//     price.usageType === SubscriptionItemPriceUsageType.Sms &&
//     price.smsRegion === SubscriptionItemPriceSmsRegion.NorthAmerica
// );
// const otherRegionsSMSPrice = subscriptionPlan.prices.find(
//   (price) =>
//     price.type === SubscriptionItemPriceType.Usage &&
//     price.usageType === SubscriptionItemPriceUsageType.Sms &&
//     price.smsRegion === SubscriptionItemPriceSmsRegion.OtherRegions
// );
// const northAmericaWhatsappPrice = subscriptionPlan.prices.find(
//   (price) =>
//     price.type === SubscriptionItemPriceType.Usage &&
//     price.usageType === SubscriptionItemPriceUsageType.Whatsapp &&
//     price.whatsappRegion === SubscriptionItemPriceWhatsappRegion.NorthAmerica
// );
// const otherRegionsWhatsappPrice = subscriptionPlan.prices.find(
//   (price) =>
//     price.type === SubscriptionItemPriceType.Usage &&
//     price.usageType === SubscriptionItemPriceUsageType.Whatsapp &&
//     price.whatsappRegion === SubscriptionItemPriceWhatsappRegion.OtherRegions
// );
// const mauPrice = subscriptionPlan.prices.find(
//   (price) =>
//     price.type === SubscriptionItemPriceType.Usage &&
//     price.usageType === SubscriptionItemPriceUsageType.Mau
// );

// const previousPlanName = getPreviousPlan(name);
// const cardTag = shouldShowRecommendedTag(name, currentPlanName) ? (
//   <CardTag>
//     <FormattedMessage id="SubscriptionScreen.recommended" />
//   </CardTag>
// ) : null;

//   if (!isPlan(planName)) {
//     console.warn("unknown plan:", planName);
//     return <></>;
//   }

//   switch (planName) {
//     case "free":
//     case "free-approved":
//       return <PlanCardFree currentPlan={currentPlanName} />;

//     default:
//       console.warn("unknown plan:", planName);
//       return <></>;
//   }
// }

interface SubscriptionScreenContentProps {
  appID: string;
  planName: string;
  subscription?: Subscription;
  subscriptionPlans: SubscriptionPlan[];
  thisMonthUsage?: SubscriptionUsage;
  previousMonthUsage?: SubscriptionUsage;
  effectiveAppConfig?: PortalAPIAppConfig;
}

function getTotalCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage,
  skipFixedPriceType: boolean
): number | undefined {
  if (!isStripePlan(planName)) {
    return undefined;
  }

  let totalCost = 0;
  for (const item of subscriptionUsage.items) {
    if (skipFixedPriceType && item.type === "FIXED") {
      continue;
    }
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
  if (!isStripePlan(planName)) {
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

interface SMSCost {
  totalCost: number;
  northAmericaCount: number;
  otherRegionsCount: number;
}

function getWhatsappCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): SMSCost | undefined {
  if (!isStripePlan(planName)) {
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
      item.usageType === SubscriptionItemPriceUsageType.Whatsapp
    ) {
      cost.totalCost += item.totalAmount ?? 0;
      if (
        item.whatsappRegion === SubscriptionItemPriceWhatsappRegion.NorthAmerica
      ) {
        cost.northAmericaCount = item.quantity;
      }
      if (
        item.whatsappRegion === SubscriptionItemPriceWhatsappRegion.OtherRegions
      ) {
        cost.otherRegionsCount = item.quantity;
      }
    }
  }

  return cost;
}

interface MAUCost {
  totalCost: number;
  additionalMAU: number;
}

function getMAUCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): MAUCost | undefined {
  if (!isStripePlan(planName)) {
    return undefined;
  }

  for (const item of subscriptionUsage.items) {
    if (
      item.type === SubscriptionItemPriceType.Usage &&
      item.usageType === SubscriptionItemPriceUsageType.Mau
    ) {
      const additionalMAU = Math.max(
        0,
        item.quantity - (item.freeQuantity ?? 0)
      );
      const totalCost = item.totalAmount;
      if (totalCost != null) {
        return {
          totalCost,
          additionalMAU,
        };
      }
    }
  }

  return undefined;
}

function SubscriptionScreenContent(props: SubscriptionScreenContentProps) {
  const {
    appID,
    planName,
    subscription,
    // subscriptionPlans,
    thisMonthUsage,
    previousMonthUsage,
    effectiveAppConfig,
  } = props;
  const { themes } = useSystemConfig();

  const showFreePlanWarning = useMemo(
    () => shouldShowFreePlanWarning(planName, effectiveAppConfig),
    [planName, effectiveAppConfig]
  );

  const subscriptionEndedAt = useMemo(() => {
    if (subscription?.endedAt != null) {
      return new Date(subscription.endedAt);
    }
    return undefined;
  }, [subscription?.endedAt]);

  const subscriptionCancelled = useMemo(() => {
    return !!subscription?.endedAt;
  }, [subscription?.endedAt]);

  const totalCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    const skipFixedPriceType = subscriptionCancelled;
    return getTotalCost(planName, thisMonthUsage, skipFixedPriceType);
  }, [planName, thisMonthUsage, subscriptionCancelled]);

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

  const mauCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getMAUCost(planName, thisMonthUsage);
  }, [planName, thisMonthUsage]);

  const baseAmount = useMemo(() => {
    if (!isStripePlan(planName)) {
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
    return getMAULimit(planName);
  }, [planName]);

  const mauPrevious = useMemo(() => {
    return previousMonthUsage?.items.find(
      (a) =>
        a.type === SubscriptionItemPriceType.Usage &&
        a.usageType === SubscriptionItemPriceUsageType.Mau
    )?.quantity;
  }, [previousMonthUsage]);

  const nextBillingDate = useMemo(() => {
    if (!isStripePlan(planName)) {
      return undefined;
    }

    const nextBillingDate = thisMonthUsage?.nextBillingDate;
    if (nextBillingDate != null) {
      return new Date(nextBillingDate);
    }
    return undefined;
  }, [planName, thisMonthUsage]);

  const [enterpriseDialogHidden, setEnterpriseDialogHidden] = useState(true);
  const [cancelDialogHidden, setCancelDialogHidden] = useState(true);

  const enterpriseDialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      type: DialogType.normal,
      title: <FormattedMessage id="SubscriptionScreen.enterprise.title" />,
      // @ts-expect-error
      subText: (
        <FormattedMessage id="SubscriptionScreen.enterprise.instructions" />
      ) as IDialogContentProps["subText"],
    };
  }, []);

  const cancelDialogContentProps: IDialogContentProps = useMemo(() => {
    if (!subscription) {
      return {
        type: DialogType.normal,
        title: <FormattedMessage id="SubscriptionPlanCard.cancel.title" />,
        // @ts-expect-error
        subText: (
          <FormattedMessage id="SubscriptionPlanCard.cancel.confirmation.customPlan" />
        ) as IDialogContentProps["subText"],
      };
    }
    return {
      type: DialogType.normal,
      title: <FormattedMessage id="SubscriptionPlanCard.cancel.title" />,
      // @ts-expect-error
      subText: (
        <FormattedMessage id="SubscriptionPlanCard.cancel.confirmation" />
      ) as IDialogContentProps["subText"],
    };
  }, [subscription]);

  const onClickEnterprisePlan = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setEnterpriseDialogHidden(false);
  }, []);

  // const onClickEnterprisePlanContactUs = useCallback(() => {
  //   setEnterpriseDialogHidden(false);
  // }, []);

  const onClickCancel = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setCancelDialogHidden(false);
  }, []);

  const onDismiss = useCallback(() => {
    setEnterpriseDialogHidden(true);
    setCancelDialogHidden(true);
  }, []);

  const [generateCustomPortalSession, { loading: manageSubscriptionLoading }] =
    useGenerateStripeCustomerPortalSessionMutationMutation({
      variables: {
        appID,
      },
    });
  useLoading(manageSubscriptionLoading);

  const onClickManageSubscription = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      generateCustomPortalSession().then(
        (r) => {
          const url = r.data?.generateStripeCustomerPortalSession.url;
          if (url != null) {
            window.location.href = url;
          }
        },
        () => {}
      );
    },
    [generateCustomPortalSession]
  );

  const {
    setSubscriptionCancelledStatus,
    loading: cancelSubscriptionLoading,
    error: cancelSubscriptionError,
  } = useSetSubscriptionCancelledStatusMutation(appID);
  useLoading(cancelSubscriptionLoading);

  const onClickCancelSubscriptionConfirm = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      setSubscriptionCancelledStatus(true)
        .catch(() => {})
        .finally(() => {
          onDismiss();
        });
      setCancelDialogHidden(false);
    },
    [setSubscriptionCancelledStatus, onDismiss, setCancelDialogHidden]
  );

  const isLoading = useIsLoading();
  return (
    <>
      <Dialog
        hidden={cancelDialogHidden}
        onDismiss={onDismiss}
        dialogContentProps={cancelDialogContentProps}
      >
        <DialogFooter>
          {!!subscription ? (
            <ButtonWithLoading
              theme={themes.destructive}
              loading={cancelSubscriptionLoading}
              onClick={onClickCancelSubscriptionConfirm}
              disabled={cancelDialogHidden}
              labelId="confirm"
            />
          ) : (
            <PrimaryButton
              href="mailto:hello@authgear.com"
              onClick={onDismiss}
              disabled={cancelDialogHidden}
              text={
                <FormattedMessage id="SubscriptionPlanCard.cancel.confirmation.customPlan.button" />
              }
            />
          )}
          <DefaultButton
            onClick={onDismiss}
            disabled={cancelSubscriptionLoading || cancelDialogHidden}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
      <ErrorDialog
        error={cancelSubscriptionError}
        rules={[]}
        fallbackErrorMessageID="SubscriptionPlanCard.cancel.error"
      />
      <Dialog
        hidden={enterpriseDialogHidden}
        onDismiss={onDismiss}
        dialogContentProps={enterpriseDialogContentProps}
      >
        <DialogFooter>
          <PrimaryButton
            href="mailto:hello@authgear.com"
            onClick={onDismiss}
            text={<FormattedMessage id="SubscriptionScreen.enterprise.cta" />}
          />
        </DialogFooter>
      </Dialog>

      <div className={styles.root}>
        <ScreenTitle className={styles.section}>
          <FormattedMessage id="SubscriptionScreen.title" />
        </ScreenTitle>
        <SubscriptionCurrentPlanSummary
          className={styles.section}
          planName={planName}
          isCustomPlan={isCustomPlan(planName)}
          baseAmount={baseAmount}
          mauCurrent={mauCurrent}
          mauLimit={mauLimit}
          mauPrevious={mauPrevious}
          subscriptionEndedAt={subscriptionEndedAt}
          nextBillingDate={nextBillingDate}
          onClickManageSubscription={onClickManageSubscription}
          manageSubscriptionLoading={manageSubscriptionLoading}
          manageSubscriptionDisabled={isLoading}
          showFreePlanWarning={showFreePlanWarning}
        >
          <CostItem
            title={
              <FormattedMessage id="SubscriptionCurrentPlanSummary.total-cost.title" />
            }
            kind={
              isFreePlan(planName)
                ? "free"
                : totalCost == null
                ? "non-applicable"
                : "billed"
            }
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
            kind={whatsappCost == null ? "non-applicable" : "billed"}
            amount={whatsappCost == null ? undefined : whatsappCost.totalCost}
            tooltip={
              whatsappCost == null ? undefined : (
                <FormattedMessage
                  id="SubscriptionCurrentPlanSummary.whatsapp.tooltip"
                  values={{
                    count1: whatsappCost.northAmericaCount,
                    count2: whatsappCost.otherRegionsCount,
                  }}
                />
              )
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
          {mauCost != null ? (
            <CostItem
              title={
                <FormattedMessage id="SubscriptionCurrentPlanSummary.additional-mau.title" />
              }
              kind="billed"
              amount={mauCost.totalCost}
              tooltip={
                <FormattedMessage
                  id="SubscriptionCurrentPlanSummary.additional-mau.tooltip"
                  values={{
                    count: mauCost.additionalMAU,
                  }}
                />
              }
            />
          ) : null}
        </SubscriptionCurrentPlanSummary>
        <div className="grid grid-flow-col grid-rows-1 auto-cols-[1fr] gap-4">
          <PlanCardFree
            currentPlan={planName}
            subscriptionCancelled={subscriptionCancelled}
          />
          <PlanCardDevelopers
            currentPlan={planName}
            subscriptionCancelled={subscriptionCancelled}
          />
          <PlanCardBusiness
            currentPlan={planName}
            subscriptionCancelled={subscriptionCancelled}
          />
          <PlanCardEnterprise
            currentPlan={planName}
            subscriptionCancelled={subscriptionCancelled}
          />
        </div>
        <SubscriptionScreenFooter
          className={styles.section}
          onClickEnterprisePlan={onClickEnterprisePlan}
          onClickCancel={onClickCancel}
          subscriptionCancelled={subscriptionCancelled}
          isStripePlan={isStripePlan(planName)}
          subscriptionEndedAt={subscription?.endedAt ?? undefined}
        />
      </div>
    </>
  );
}

interface SubscriptionProcessingPaymentScreenProps {
  stripeError?: stripe.Error;
}

const SubscriptionProcessingPaymentScreen: React.VFC<SubscriptionProcessingPaymentScreenProps> =
  function SubscriptionProcessingPaymentScreen(
    props: SubscriptionProcessingPaymentScreenProps
  ) {
    const { stripeError } = props;
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const {
      cancelFailedSubscription,
      loading: cancelFailedSubscriptionLoading,
      error: cancelFailedSubscriptionError,
    } = useCancelFailedSubscriptionMutation(appID);
    useLoading(cancelFailedSubscriptionLoading);

    const paymentStatus = useMemo(() => {
      if (stripeError == null) {
        return "IsProcessing";
      }
      // https://stripe.com/docs/error-codes
      if (stripeError.code === "card_declined") {
        return "CardDeclined";
      }
      return "UnknownError";
    }, [stripeError]);

    const onClickCancelFailedSubscription = useCallback(async () => {
      await cancelFailedSubscription();
    }, [cancelFailedSubscription]);

    return (
      <div className={styles.root}>
        <ScreenTitle className={styles.section}>
          <FormattedMessage id="SubscriptionScreen.title" />
        </ScreenTitle>
        <div
          className={cn(styles.processingPaymentSection)}
          style={{
            boxShadow: DefaultEffects.elevation4,
          }}
        >
          {paymentStatus === "IsProcessing" ? (
            <>
              <Spinner
                className={styles.processingPaymentSpinner}
                labelPosition="right"
                label={renderToString("SubscriptionScreen.processing-payment")}
                size={SpinnerSize.large}
                styles={{
                  label: {
                    whiteSpace: "pre-line",
                    textAlign: "left",
                    marginLeft: "16px",
                    fontSize: "14px",
                    lineHeight: "20px",
                  },
                }}
              />
              <ThemeProvider theme={CONTACT_US_BUTTON_THEME}>
                <ExternalLink href={"mailto:hello@authgear.com"}>
                  <Text>
                    <FormattedMessage id="SubscriptionScreen.contact-us.label" />
                  </Text>
                </ExternalLink>
              </ThemeProvider>
            </>
          ) : null}
          {paymentStatus === "CardDeclined" ? (
            <>
              <Text className={styles.processingPaymentErrorMessage}>
                <FormattedMessage id="SubscriptionScreen.payment-declined.description" />
              </Text>
              <div className={styles.processingPaymentButtonContainer}>
                <ButtonWithLoading
                  loading={cancelFailedSubscriptionLoading}
                  onClick={onClickCancelFailedSubscription}
                  labelId="SubscriptionScreen.cancel-transaction.label"
                />
              </div>
            </>
          ) : null}
          {paymentStatus === "UnknownError" ? (
            <>
              <Text className={styles.processingPaymentErrorMessage}>
                <FormattedMessage id="SubscriptionScreen.unknown-error.description" />
              </Text>
            </>
          ) : null}
          <ErrorDialog
            error={cancelFailedSubscriptionError}
            rules={[]}
            fallbackErrorMessageID="SubscriptionScreen.cancel-transaction-error.description"
          />
        </div>
      </div>
    );
  };

const SubscriptionScreen: React.VFC = function SubscriptionScreen() {
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

  const isProcessingSubscription =
    !!subscriptionScreenQuery.data &&
    (subscriptionScreenQuery.data.node as AppFragmentFragment)
      .isProcessingSubscription;

  const lastStripeError = useMemo(() => {
    return (
      !!subscriptionScreenQuery.data &&
      (subscriptionScreenQuery.data.node as AppFragmentFragment).lastStripeError
    );
  }, [subscriptionScreenQuery]);

  // if isProcessingSubscription is true
  // refetch in every few seconds and wait until it changes to false
  useEffect(() => {
    if (subscriptionScreenQuery.loading) {
      return () => {};
    }
    if (!isProcessingSubscription) {
      return () => {};
    }
    const interval = setInterval(() => {
      subscriptionScreenQuery.refetch().finally(() => {});
    }, CHECK_IS_PROCESSING_SUBSCRIPTION_INTERVAL);
    return () => {
      clearInterval(interval);
    };
  }, [
    subscriptionScreenQuery.loading,
    isProcessingSubscription,
    subscriptionScreenQuery,
  ]);

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

  if (isProcessingSubscription) {
    return (
      <SubscriptionProcessingPaymentScreen stripeError={lastStripeError} />
    );
  }

  const planName = (subscriptionScreenQuery.data?.node as AppFragmentFragment)
    .planName;
  const subscription = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).subscription;
  const subscriptionPlans =
    subscriptionScreenQuery.data?.subscriptionPlans ?? [];
  const thisMonthUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).thisMonth;
  const previousMonthUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).previousMonth;

  const effectiveAppConfig = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).effectiveAppConfig as PortalAPIAppConfig | null | undefined;

  return (
    <ScreenLayoutScrollView>
      <SubscriptionScreenContent
        appID={appID}
        planName={planName}
        subscription={subscription ?? undefined}
        subscriptionPlans={subscriptionPlans}
        thisMonthUsage={thisMonthUsage ?? undefined}
        previousMonthUsage={previousMonthUsage ?? undefined}
        effectiveAppConfig={effectiveAppConfig ?? undefined}
      />
    </ScreenLayoutScrollView>
  );
};

export default SubscriptionScreen;
