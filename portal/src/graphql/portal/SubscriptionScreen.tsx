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
  SubscriptionPlan,
  SubscriptionUsage,
} from "./globalTypes.generated";
import { AppFragmentFragment } from "./query/subscriptionScreenQuery.generated";
import { useSubscriptionScreenQueryQuery } from "./query/subscriptionScreenQuery";
import { useGenerateStripeCustomerPortalSessionMutationMutation } from "./mutations/generateStripeCustomerPortalSessionMutation";
import { useUpdateSubscriptionMutation } from "./mutations/updateSubscriptionMutation";
import styles from "./SubscriptionScreen.module.css";
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
import { useCreateCheckoutSessionMutation } from "./mutations/createCheckoutSessionMutation";
import { useLoading, useIsLoading } from "./../../hook/loading";
import { formatDatetime } from "../../util/formatDatetime";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useSetSubscriptionCancelledStatusMutation } from "./mutations/setSubscriptionCancelledStatusMutation";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ErrorDialog from "../../error/ErrorDialog";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import LinkButton from "../../LinkButton";
import { StripeError } from "../../types";

const ALL_KNOWN_PLANS = ["free", "developers", "startups", "business"];
const PAID_PLANS = ALL_KNOWN_PLANS.slice(1);

const MAU_LIMIT: Record<string, number> = {
  free: 5000,
  developers: 1000,
  startups: 5000,
  business: 50000,
};

const CHECK_IS_PROCESSING_SUBSCRIPTION_INTERVAL = 5000;

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

function isFreePlan(planName: string): boolean {
  return planName === "free";
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
  subscriptionCancelled: boolean;
  subscriptionPlan: SubscriptionPlan;
  nextBillingDate?: Date;
}

// eslint-disable-next-line complexity
function SubscriptionPlanCardRenderer(props: SubscriptionPlanCardRenderProps) {
  const {
    currentPlanName,
    subscriptionCancelled,
    subscriptionPlan,
    nextBillingDate,
  } = props;
  const { appID } = useParams() as { appID: string };
  const { createCheckoutSession, loading: createCheckoutSessionLoading } =
    useCreateCheckoutSessionMutation();
  useLoading(createCheckoutSessionLoading);
  const [updateSubscription, { loading: updateSubscriptionLoading }] =
    useUpdateSubscriptionMutation();
  useLoading(updateSubscriptionLoading);
  const {
    setSubscriptionCancelledStatus,
    loading: reactivateSubscriptionLoading,
    error: reactivateSubscriptionError,
  } = useSetSubscriptionCancelledStatusMutation(appID);
  useLoading(reactivateSubscriptionLoading);

  const isLoading = useIsLoading();

  const ctaVariant = useMemo(() => {
    if (!isKnownPlan(currentPlanName)) {
      return "non-applicable";
    }
    if (!isKnownPaidPlan(currentPlanName)) {
      return "subscribe";
    }
    const targetPlan = subscriptionPlan.name;
    const currentPlanIdx = ALL_KNOWN_PLANS.indexOf(currentPlanName);
    const targetPlanIdx = ALL_KNOWN_PLANS.indexOf(targetPlan);
    if (subscriptionCancelled) {
      if (currentPlanIdx === targetPlanIdx) {
        return "reactivate";
      }
      return "non-applicable";
    }
    if (currentPlanIdx > targetPlanIdx) {
      return "downgrade";
    } else if (currentPlanIdx < targetPlanIdx) {
      return "upgrade";
    }
    return "current";
  }, [currentPlanName, subscriptionPlan.name, subscriptionCancelled]);

  const onClickSubscribe = useCallback(
    (planName: string) => {
      createCheckoutSession(appID, planName)
        .then((url) => {
          if (url) {
            window.location.href = url;
          }
        })
        .finally(() => {});
    },
    [appID, createCheckoutSession]
  );

  const onClickUpgrade = useCallback(
    (planName: string) => {
      updateSubscription({
        variables: {
          appID,
          planName,
        },
      }).finally(() => {});
    },
    [appID, updateSubscription]
  );

  const onClickDowngrade = useCallback(
    (planName: string) => {
      updateSubscription({
        variables: {
          appID,
          planName,
        },
      }).finally(() => {});
    },
    [appID, updateSubscription]
  );

  const onClickReactivate = useCallback(async () => {
    await setSubscriptionCancelledStatus(false);
  }, [setSubscriptionCancelledStatus]);

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
  const mauPrice = subscriptionPlan.prices.find(
    (price) =>
      price.type === SubscriptionItemPriceType.Usage &&
      price.usageType === SubscriptionItemPriceUsageType.Mau
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
          {basePrice != null
            ? `$${basePrice.unitAmount / 100}${mauPrice == null ? "" : "+"}/mo`
            : "-"}
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
          {mauPrice != null ? (
            <UsagePriceTag>
              <FormattedMessage
                id="SubscriptionPlanCard.mau"
                values={{
                  unitAmount: mauPrice.unitAmount / 100,
                  divisor: mauPrice.transformQuantityDivideBy ?? 1,
                }}
              />
            </UsagePriceTag>
          ) : null}
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
      cta={
        <CTA
          appID={appID}
          planName={subscriptionPlan.name}
          variant={ctaVariant}
          disabled={isLoading}
          onClickSubscribe={onClickSubscribe}
          onClickUpgrade={onClickUpgrade}
          onClickDowngrade={onClickDowngrade}
          onClickReactivate={onClickReactivate}
          reactivateError={reactivateSubscriptionError}
          reactivateLoading={reactivateSubscriptionLoading}
          nextBillingDate={nextBillingDate}
        />
      }
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
  appID: string;
  planName: string;
  subscription?: Subscription;
  subscriptionPlans: SubscriptionPlan[];
  thisMonthUsage?: SubscriptionUsage;
  previousMonthUsage?: SubscriptionUsage;
}

function getTotalCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage,
  skipFixedPriceType: boolean
): number | undefined {
  if (!isKnownPaidPlan(planName)) {
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

interface MAUCost {
  totalCost: number;
  additionalMAU: number;
}

function getMAUCost(
  planName: string,
  subscriptionUsage: SubscriptionUsage
): MAUCost | undefined {
  if (!isKnownPaidPlan(planName)) {
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

const CANCEL_THEME: PartialTheme = {
  palette: {
    themePrimary: "#c8c8c8",
    neutralPrimary: "#c8c8c8",
  },
  semanticColors: {
    linkHovered: "#c8c8c8",
  },
};

// eslint-disable-next-line complexity
function SubscriptionScreenContent(props: SubscriptionScreenContentProps) {
  const { locale } = useContext(Context);
  const {
    appID,
    planName,
    subscription,
    subscriptionPlans,
    thisMonthUsage,
    previousMonthUsage,
  } = props;
  const { themes } = useSystemConfig();

  const hasSubscription = useMemo(() => !!subscription, [subscription]);

  const formattedSubscriptionEndedAt = useMemo(() => {
    return subscription?.endedAt
      ? formatDatetime(locale, subscription.endedAt, DateTime.DATETIME_SHORT)
      : null;
  }, [subscription?.endedAt, locale]);

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

  const mauCost = useMemo(() => {
    if (thisMonthUsage == null) {
      return undefined;
    }
    return getMAUCost(planName, thisMonthUsage);
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

  const [enterpriseDialogHidden, setEnterpriseDialogHidden] = useState(true);
  const [cancelDialogHidden, setCancelDialogHidden] = useState(true);

  // @ts-expect-error
  const enterpriseDialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      type: DialogType.normal,
      title: <FormattedMessage id="SubscriptionScreen.enterprise.title" />,
      subText: (
        <FormattedMessage id="SubscriptionScreen.enterprise.instructions" />
      ),
    };
  }, []);

  // @ts-expect-error
  const cancelDialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      type: DialogType.normal,
      title: <FormattedMessage id="SubscriptionPlanCard.cancel.title" />,
      subText: (
        <FormattedMessage id="SubscriptionPlanCard.cancel.confirmation" />
      ),
    };
  }, []);

  const onClickEnterprisePlan = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setEnterpriseDialogHidden(false);
  }, []);

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
          <ButtonWithLoading
            theme={themes.destructive}
            loading={cancelSubscriptionLoading}
            onClick={onClickCancelSubscriptionConfirm}
            disabled={cancelDialogHidden}
            labelId="confirm"
          />
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
          baseAmount={baseAmount}
          mauCurrent={mauCurrent}
          mauLimit={mauLimit}
          mauPrevious={mauPrevious}
          subscriptionEndedAt={subscriptionEndedAt}
          nextBillingDate={nextBillingDate}
          onClickManageSubscription={onClickManageSubscription}
          manageSubscriptionLoading={manageSubscriptionLoading}
          manageSubscriptionDisabled={isLoading}
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
            kind={isKnownPaidPlan(planName) ? "free" : "non-applicable"}
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
            {PAID_PLANS.map((paidPlanName) => {
              const plan = subscriptionPlans.find(
                (plan) => plan.name === paidPlanName
              );
              if (plan != null) {
                return (
                  <SubscriptionPlanCardRenderer
                    key={plan.name}
                    subscriptionCancelled={subscriptionCancelled}
                    subscriptionPlan={plan}
                    currentPlanName={planName}
                    nextBillingDate={nextBillingDate}
                  />
                );
              }
              return null;
            })}
          </div>
        </div>
        <div className={styles.footer}>
          <Text block={true}>
            <FormattedMessage id="SubscriptionScreen.footer.tax" />
          </Text>
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
          {isKnownPaidPlan(planName) ? (
            <>
              <Text block={true}>
                <FormattedMessage id="SubscriptionScreen.footer.usage-delay-disclaimer" />
              </Text>
              {hasSubscription ? (
                subscriptionCancelled ? (
                  <Text block={true}>
                    <FormattedMessage
                      id="SubscriptionScreen.footer.expire"
                      values={{
                        date: formattedSubscriptionEndedAt ?? "",
                      }}
                    />
                  </Text>
                ) : (
                  <ThemeProvider theme={CANCEL_THEME}>
                    <LinkButton onClick={onClickCancel}>
                      <Text>
                        <FormattedMessage id="SubscriptionScreen.footer.cancel" />
                      </Text>
                    </LinkButton>
                  </ThemeProvider>
                )
              ) : null}
            </>
          ) : null}
        </div>
      </div>
    </>
  );
}

interface SubscriptionProcessingPaymentScreenProps {
  stripeError?: StripeError;
}

const SubscriptionProcessingPaymentScreen: React.VFC<SubscriptionProcessingPaymentScreenProps> =
  function SubscriptionProcessingPaymentScreen(
    props: SubscriptionProcessingPaymentScreenProps
  ) {
    const { stripeError } = props;
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const paymentStatus = useMemo(() => {
      if (stripeError == null) {
        return "IsProcessing";
      }
      if (stripeError.code === "card_declined") {
        return "CardDeclined";
      }
      return "UnknownError";
    }, [stripeError]);

    const onClickCancelFailedSubscription = useCallback(() => {}, [appID]);

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
                },
              }}
            />
          ) : null}
          {paymentStatus === "CardDeclined" ? (
            <>
              <Text className={styles.processingPaymentErrorMessage}>
                <FormattedMessage id="SubscriptionScreen.payment-declined.description" />
              </Text>
              <div className={styles.processingPaymentButtonContainer}>
                <PrimaryButton
                  onClick={onClickCancelFailedSubscription}
                  text={
                    <FormattedMessage id="SubscriptionScreen.cancel-transaction.label" />
                  }
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

  return (
    <ScreenLayoutScrollView>
      <SubscriptionScreenContent
        appID={appID}
        planName={planName}
        subscription={subscription ?? undefined}
        subscriptionPlans={subscriptionPlans}
        thisMonthUsage={thisMonthUsage ?? undefined}
        previousMonthUsage={previousMonthUsage ?? undefined}
      />
    </ScreenLayoutScrollView>
  );
};

export default SubscriptionScreen;
