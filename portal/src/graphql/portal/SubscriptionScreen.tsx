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
  Tabs,
  Text as RadixText,
  Dialog,
  Button,
  Spinner,
} from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  Subscription,
  SubscriptionPlan,
  SubscriptionUsage,
  Usage,
} from "./globalTypes.generated";
import { PortalAPIAppConfig } from "../../types";
import { AppFragmentFragment } from "./query/subscriptionScreenQuery.generated";
import { useSubscriptionScreenQueryQuery } from "./query/subscriptionScreenQuery";
import styles from "./SubscriptionScreen.module.css";
import { useLoading, useIsLoading } from "./../../hook/loading";
import { useSetSubscriptionCancelledStatusMutation } from "./mutations/setSubscriptionCancelledStatusMutation";
import ErrorDialog from "../../error/ErrorDialog";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { useCancelFailedSubscriptionMutation } from "./mutations/cancelFailedSubscriptionMutation";
import ExternalLink from "../../ExternalLink";
import {
  DEFAULT_FREE_PLAN,
  isStripePlan,
  Plan,
  CTAVariant,
} from "../../util/plan";
import {
  PlanCardBusiness,
  PlanCardDevelopers,
  PlanCardEnterprise,
  PlanCardFree,
} from "../../components/billing/PlanCard";
import { useCreateCheckoutSessionMutation } from "./mutations/createCheckoutSessionMutation";
import { useUpdateSubscriptionMutation } from "./mutations/updateSubscriptionMutation";
import { usePreviewUpdateSubscriptionMutation } from "./mutations/previewUpdateSubscriptionMutation";
import { formatDateOnly } from "../../util/formatDateOnly";
import { FeatureBanner } from "../../components/billing/FeatureBanner";
import { CurrentPlanCard } from "../../components/billing/CurrentPlanCard";
import { usePivotNavigation } from "../../hook/usePivot";
import { useGenerateStripeCustomerPortalSessionMutationMutation } from "./mutations/generateStripeCustomerPortalSessionMutation";
import { CancelSubscriptionReminder } from "../../components/billing/CancelSubscriptionReminder";
import { TextButton } from "../../components/v2/Button/TextButton/TextButton";
import { extractRawID } from "../../util/graphql";
import { CancelSubscriptionSurveyDialog } from "../../components/billing/CancelSubscriptionSurveyDialog";

const CHECK_IS_PROCESSING_SUBSCRIPTION_INTERVAL = 5000;

function PlansSection({
  currentPlanName,
  subscriptionCancelled,
  nextBillingDate,
  subscriptionPlans,
  onClickContactUs,
  onClickCancelSubscription,
}: {
  currentPlanName: string;
  subscriptionCancelled: boolean;
  nextBillingDate: Date | undefined;
  subscriptionPlans: SubscriptionPlan[];
  onClickContactUs: () => void;
  onClickCancelSubscription: () => void;
}) {
  const { locale } = useContext(Context);
  const [upgradeToPlan, setUpgradeToPlan] = useState<string | null>(null);
  const [downgradeToPlan, setDowngradeToPlan] = useState<string | null>(null);
  const [isReactiveDialogHidden, setIsReactiveDialogHidden] =
    useState<boolean>(true);
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

  const [previewUpdateSubscription, { data, loading }] =
    usePreviewUpdateSubscriptionMutation();
  useLoading(loading);

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

  const onConfirmUpgrade = useCallback(() => {
    if (!upgradeToPlan) {
      console.error("upgradeToPlan should not be null");
      return;
    }
    updateSubscription({
      variables: {
        appID,
        planName: upgradeToPlan,
      },
    }).finally(() => {
      setUpgradeToPlan(null);
    });
  }, [appID, updateSubscription, upgradeToPlan]);

  const onConfirmDowngrade = useCallback(() => {
    if (!downgradeToPlan) {
      console.error("downgradeToPlan should not be null");
      return;
    }
    updateSubscription({
      variables: {
        appID,
        planName: downgradeToPlan,
      },
    }).finally(() => {
      setDowngradeToPlan(null);
    });
  }, [appID, downgradeToPlan, updateSubscription]);

  const onClickUpgrade = useCallback(
    (planName: string) => {
      previewUpdateSubscription({
        variables: {
          appID,
          planName,
        },
      }).finally(() => {});
      setUpgradeToPlan(planName);
    },
    [appID, previewUpdateSubscription]
  );

  const onClickDowngrade = useCallback(
    (planName: string) => {
      previewUpdateSubscription({
        variables: {
          appID,
          planName,
        },
      }).finally(() => {});
      setDowngradeToPlan(planName);
    },
    [appID, previewUpdateSubscription]
  );

  const onClickReactivate = useCallback(() => {
    setIsReactiveDialogHidden(false);
  }, []);

  const onClickConfirmReactivate = useCallback(async () => {
    try {
      await setSubscriptionCancelledStatus(false);
    } finally {
      setIsReactiveDialogHidden(true);
    }
  }, [setSubscriptionCancelledStatus]);

  const onPlanAction = useMemo(() => {
    const plans: Plan[] = ["enterprise"];
    if (
      subscriptionPlans.findIndex((p) => p.name === "developers2025") !== -1
    ) {
      plans.push("developers2025");
    }
    if (subscriptionPlans.findIndex((p) => p.name === "business2025") !== -1) {
      plans.push("business2025");
    }
    plans.push("enterprise");

    return plans.reduce<Partial<Record<Plan, (action: CTAVariant) => void>>>(
      (callbacks, plan) => {
        const fn = (action: CTAVariant) => {
          switch (action) {
            case "subscribe":
              onClickSubscribe(plan);
              break;
            case "upgrade":
              onClickUpgrade(plan);
              break;
            case "downgrade":
              onClickDowngrade(plan);
              break;
            case "reactivate":
              onClickReactivate();
              break;
            case "contact-us":
              onClickContactUs();
              break;
            case "current":
            case "non-applicable":
            default:
              console.error(
                `action button clicked but action:${action} should not be clickable. plan: ${plan}`
              );
              break;
          }
        };
        callbacks[plan] = fn;
        return callbacks;
      },
      {}
    );
  }, [
    onClickContactUs,
    onClickDowngrade,
    onClickReactivate,
    onClickSubscribe,
    onClickUpgrade,
    subscriptionPlans,
  ]);

  const onFreePlanAction = useCallback(
    (action: CTAVariant) => {
      switch (action) {
        case "downgrade":
          // Downgrade to the default free plan means cancel subscription.
          onClickCancelSubscription();
          break;
        // All other cases should not happen
        default:
          console.error(
            `action button clicked but action:${action} should not be clickable. plan: ${DEFAULT_FREE_PLAN}`
          );
          break;
      }
    },
    [onClickCancelSubscription]
  );

  const amountDue =
    data?.previewUpdateSubscription.amountDue != null
      ? data.previewUpdateSubscription.amountDue / 100
      : null;
  const formattedDate = formatDateOnly(locale, nextBillingDate ?? null);

  const isLoading = useIsLoading();

  const onDismissUpgradeDialog = useCallback(() => {
    setUpgradeToPlan(null);
  }, []);

  const onDismissDowngradeDialog = useCallback(() => {
    setDowngradeToPlan(null);
  }, []);

  const onDismissReactiveDialog = useCallback(() => {
    setIsReactiveDialogHidden(true);
  }, []);

  return (
    <>
      <div className="overflow-x-auto w-full">
        <div className="grid grid-flow-col grid-rows-1 auto-cols-[1fr] gap-4">
          <PlanCardFree
            currentPlan={currentPlanName}
            subscriptionCancelled={subscriptionCancelled}
            nextBillingDate={nextBillingDate}
            onAction={onFreePlanAction}
          />
          {onPlanAction.developers2025 != null ? (
            <PlanCardDevelopers
              currentPlan={currentPlanName}
              subscriptionCancelled={subscriptionCancelled}
              onAction={onPlanAction.developers2025}
            />
          ) : null}
          {onPlanAction.business2025 != null ? (
            <PlanCardBusiness
              currentPlan={currentPlanName}
              subscriptionCancelled={subscriptionCancelled}
              onAction={onPlanAction.business2025}
            />
          ) : null}
          <PlanCardEnterprise
            currentPlan={currentPlanName}
            subscriptionCancelled={subscriptionCancelled}
            onAction={onPlanAction.enterprise!}
          />
        </div>
      </div>
      <ErrorDialog
        error={reactivateSubscriptionError}
        rules={[]}
        fallbackErrorMessageID="SubscriptionScreen.reactivate.error"
        fallbackErrorMessageValues={{
          // eslint-disable-next-line react/no-unstable-nested-components
          ExternalLink: (chunks: React.ReactNode) => (
            <ExternalLink href="mailto:hello@authgear.com">
              {chunks}
            </ExternalLink>
          ),
        }}
      />
      <Dialog.Root
        open={upgradeToPlan != null}
        onOpenChange={(open) => {
          if (!open) {
            onDismissUpgradeDialog();
          }
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="SubscriptionScreen.upgrade.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            {amountDue == null ? (
              <FormattedMessage id="loading" />
            ) : (
              <FormattedMessage
                id="SubscriptionScreen.upgrade.description"
                values={{
                  amount: amountDue,
                  date: formattedDate ?? "",
                }}
              />
            )}
          </Dialog.Description>
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              onClick={onDismissUpgradeDialog}
              text={<FormattedMessage id="cancel" />}
            />
            <PrimaryButton
              size="2"
              onClick={onConfirmUpgrade}
              disabled={isLoading}
              text={<FormattedMessage id="SubscriptionScreen.label.upgrade" />}
            />
          </div>
        </Dialog.Content>
      </Dialog.Root>
      <Dialog.Root
        open={downgradeToPlan != null}
        onOpenChange={(open) => {
          if (!open) {
            onDismissDowngradeDialog();
          }
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="SubscriptionScreen.downgrade.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            {amountDue == null ? (
              <FormattedMessage id="loading" />
            ) : (
              <FormattedMessage
                id="SubscriptionScreen.downgrade.description"
                values={{
                  amount: amountDue,
                  date: formattedDate ?? "",
                }}
              />
            )}
          </Dialog.Description>
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              onClick={onDismissDowngradeDialog}
              text={<FormattedMessage id="cancel" />}
            />
            <Button
              size="2"
              variant="solid"
              color="red"
              disabled={isLoading}
              onClick={onConfirmDowngrade}
            >
              <FormattedMessage id="SubscriptionScreen.label.downgrade" />
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Root>
      <Dialog.Root
        open={!isReactiveDialogHidden}
        onOpenChange={(open) => {
          if (!open) {
            onDismissReactiveDialog();
          }
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="SubscriptionScreen.reactivate.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            <FormattedMessage id="SubscriptionScreen.reactivate.confirmation" />
          </Dialog.Description>
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              onClick={onDismissReactiveDialog}
              disabled={isReactiveDialogHidden || reactivateSubscriptionLoading}
              text={<FormattedMessage id="cancel" />}
            />
            <Button
              size="2"
              loading={reactivateSubscriptionLoading}
              disabled={isReactiveDialogHidden}
              // eslint-disable-next-line @typescript-eslint/strict-void-return
              onClick={onClickConfirmReactivate}
            >
              <FormattedMessage id="confirm" />
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Root>
    </>
  );
}

interface SubscriptionScreenContentProps {
  appID: string;
  planName: string;
  subscription?: Subscription;
  subscriptionPlans: SubscriptionPlan[];
  thisMonthUsage?: Usage;
  thisMonthSubscriptionUsage?: SubscriptionUsage;
  previousMonthSubscriptionUsage?: SubscriptionUsage;
  effectiveAppConfig?: PortalAPIAppConfig;
}

enum Tab {
  Subscription = "Subscription",
  PlanDetail = "PlanDetail",
}

function SubscriptionScreenContent(props: SubscriptionScreenContentProps) {
  const {
    appID,
    planName,
    subscription,
    subscriptionPlans,
    thisMonthUsage,
    thisMonthSubscriptionUsage,
    previousMonthSubscriptionUsage,
  } = props;

  const hasSubscription = useMemo(() => !!subscription, [subscription]);

  const subscriptionCancelled = useMemo(() => {
    return !!subscription?.endedAt;
  }, [subscription?.endedAt]);

  const nextBillingDate = useMemo(() => {
    if (!isStripePlan(planName)) {
      return undefined;
    }

    const nextBillingDate = thisMonthSubscriptionUsage?.nextBillingDate;
    if (nextBillingDate != null) {
      return new Date(nextBillingDate);
    }
    return undefined;
  }, [planName, thisMonthSubscriptionUsage]);

  const [enterpriseDialogHidden, setEnterpriseDialogHidden] = useState(true);
  const [cancelDialogHidden, setCancelDialogHidden] = useState(true);
  const [cancelSurveyDialogHidden, setCancelSurveyDialogHidden] =
    useState(true);

  const { selectedKey: selectedTab, onChangeKey: onTabChange } =
    usePivotNavigation<Tab>([Tab.Subscription, Tab.PlanDetail]);

  const onTabValueChange = useCallback(
    (value: string) => {
      if (value === Tab.Subscription || value === Tab.PlanDetail) {
        onTabChange(value);
      }
    },
    [onTabChange]
  );

  const onClickContactUs = useCallback(() => {
    setEnterpriseDialogHidden(false);
  }, []);

  const onClickCancel = useCallback((e?: React.MouseEvent) => {
    e?.preventDefault();
    e?.stopPropagation();
    setCancelDialogHidden(false);
  }, []);

  const onDismiss = useCallback(() => {
    setEnterpriseDialogHidden(true);
    setCancelDialogHidden(true);
  }, []);

  const {
    setSubscriptionCancelledStatus,
    loading: cancelSubscriptionLoading,
    error: cancelSubscriptionError,
  } = useSetSubscriptionCancelledStatusMutation(appID);
  useLoading(cancelSubscriptionLoading);

  const onClickCancelSubscriptionConfirm = useCallback(
    async (e) => {
      e.preventDefault();
      e.stopPropagation();
      await setSubscriptionCancelledStatus(true);
      setCancelDialogHidden(true);
      setCancelSurveyDialogHidden(false);
    },
    [setSubscriptionCancelledStatus]
  );

  const onConfirmCancelSurveyDialog = useCallback(() => {
    const projectID = extractRawID(appID);
    const cancelSurveyURL = `https://oursky.typeform.com/authgear-cancel#project_id=${projectID}`;
    const anchor = document.createElement("A") as HTMLAnchorElement;
    anchor.href = cancelSurveyURL;
    anchor.target = "_blank";
    anchor.click();
    anchor.remove();
    setCancelSurveyDialogHidden(true);
  }, [appID]);

  return (
    <>
      <Dialog.Root
        open={!cancelDialogHidden}
        onOpenChange={(open) => {
          if (!open) {
            onDismiss();
          }
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="SubscriptionScreen.cancel.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            {!subscription ? (
              <FormattedMessage id="SubscriptionScreen.cancel.confirmation.customPlan" />
            ) : (
              <FormattedMessage id="SubscriptionScreen.cancel.confirmation" />
            )}
          </Dialog.Description>
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              onClick={onDismiss}
              disabled={cancelSubscriptionLoading || cancelDialogHidden}
              text={<FormattedMessage id="cancel" />}
            />
            {!!subscription ? (
              <Button
                size="2"
                variant="solid"
                color="red"
                loading={cancelSubscriptionLoading}
                disabled={cancelDialogHidden}
                // eslint-disable-next-line @typescript-eslint/strict-void-return
                onClick={onClickCancelSubscriptionConfirm}
              >
                <FormattedMessage id="confirm" />
              </Button>
            ) : (
              <Button asChild={true} size="2" variant="solid" color="indigo">
                <a href="mailto:hello@authgear.com" onClick={onDismiss}>
                  <FormattedMessage id="SubscriptionScreen.cancel.confirmation.customPlan.button" />
                </a>
              </Button>
            )}
          </div>
        </Dialog.Content>
      </Dialog.Root>
      <CancelSubscriptionSurveyDialog
        isHidden={cancelSurveyDialogHidden}
        onDismiss={useCallback(() => {
          setCancelSurveyDialogHidden(true);
        }, [])}
        onConfirm={onConfirmCancelSurveyDialog}
        onCancel={useCallback(() => {
          setCancelSurveyDialogHidden(true);
        }, [])}
      />
      <ErrorDialog
        error={cancelSubscriptionError}
        rules={[]}
        fallbackErrorMessageID="SubscriptionScreen.cancel.error"
        fallbackErrorMessageValues={{
          // eslint-disable-next-line react/no-unstable-nested-components
          ExternalLink: (chunks: React.ReactNode) => (
            <ExternalLink href="mailto:hello@authgear.com">
              {chunks}
            </ExternalLink>
          ),
        }}
      />
      <Dialog.Root
        open={!enterpriseDialogHidden}
        onOpenChange={(open) => {
          if (!open) {
            onDismiss();
          }
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>
            <FormattedMessage id="SubscriptionScreen.enterprise.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            <FormattedMessage
              id="SubscriptionScreen.enterprise.instructions"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                ExternalLink: (chunks: React.ReactNode) => (
                  <ExternalLink href="mailto:hello@authgear.com">
                    {chunks}
                  </ExternalLink>
                ),
              }}
            />
          </Dialog.Description>
          <div className={styles.actions}>
            <Button asChild={true} size="2" variant="solid" color="indigo">
              <a href="mailto:hello@authgear.com" onClick={onDismiss}>
                <FormattedMessage id="SubscriptionScreen.enterprise.cta" />
              </a>
            </Button>
          </div>
        </Dialog.Content>
      </Dialog.Root>

      <ScreenContent layout="auto-rows">
        <div className={cn(styles.widget, styles.pageHeader)}>
          <RadixText as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="SubscriptionScreen.title" />
          </RadixText>
          <RadixText
            as="p"
            size="2"
            color="gray"
            className={styles.pageDescription}
          >
            <FormattedMessage id="SubscriptionScreen.description" />
          </RadixText>
        </div>
        <Tabs.Root
          className={cn(styles.widgetWide, styles.tabsRoot)}
          value={selectedTab}
          onValueChange={onTabValueChange}
        >
          <Tabs.List className={styles.tabsList}>
            <Tabs.Trigger value={Tab.Subscription}>
              <FormattedMessage id="SubscriptionScreen.tabs.subscription" />
            </Tabs.Trigger>
            <Tabs.Trigger value={Tab.PlanDetail}>
              <FormattedMessage id="SubscriptionScreen.tabs.planDetails" />
            </Tabs.Trigger>
          </Tabs.List>
          <Tabs.Content value={Tab.Subscription} className={styles.tabContent}>
            <div className="py-6 grid grid-flow-row gap-4">
              <FeatureBanner />
              <PlansSection
                currentPlanName={planName}
                subscriptionCancelled={subscriptionCancelled}
                nextBillingDate={nextBillingDate}
                subscriptionPlans={subscriptionPlans}
                onClickContactUs={onClickContactUs}
                onClickCancelSubscription={onClickCancel}
              />
              <footer className={styles.section}>
                <RadixText as="p">
                  <FormattedMessage id="SubscriptionScreen.footer.tax" />
                </RadixText>
              </footer>
            </div>
          </Tabs.Content>
          <Tabs.Content value={Tab.PlanDetail} className={styles.tabContent}>
            <div className={styles.planDetailsTabContent}>
              <PlanDetailsTab
                appID={appID}
                planName={planName}
                subscriptionCancelled={subscriptionCancelled}
                nextBillingDate={nextBillingDate}
                thisMonthUsage={thisMonthUsage}
                thisMonthSubscriptionUsage={thisMonthSubscriptionUsage}
                previousMonthSubscriptionUsage={previousMonthSubscriptionUsage}
                hasSubscription={hasSubscription}
              />
            </div>
          </Tabs.Content>
        </Tabs.Root>
      </ScreenContent>
    </>
  );
}

interface PlanDetailsTabProps {
  appID: string;
  planName: string;
  subscriptionCancelled: boolean;
  nextBillingDate: Date | undefined;
  thisMonthUsage: Usage | undefined;
  thisMonthSubscriptionUsage: SubscriptionUsage | undefined;
  previousMonthSubscriptionUsage: SubscriptionUsage | undefined;
  hasSubscription: boolean;
}

function PlanDetailsTab({
  appID,
  planName,
  subscriptionCancelled,
  nextBillingDate,
  thisMonthUsage,
  thisMonthSubscriptionUsage,
  previousMonthSubscriptionUsage,
  hasSubscription,
}: PlanDetailsTabProps) {
  const { locale } = useContext(Context);
  const formattedBillingDate = useMemo(
    () => formatDateOnly(locale, nextBillingDate ?? null),
    [locale, nextBillingDate]
  );
  const isLoading = useIsLoading();

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

  return (
    <div className={styles.planDetailsTab}>
      <div className={styles.planDetailsHeader}>
        {subscriptionCancelled && formattedBillingDate != null ? (
          <CancelSubscriptionReminder
            formattedBillingDate={formattedBillingDate}
          />
        ) : null}
        {formattedBillingDate ? (
          <RadixText as="p" size="2" color="gray">
            <FormattedMessage
              id="SubscriptionScreen.planDetails.nextBillingDate"
              values={{ date: formattedBillingDate }}
            />
          </RadixText>
        ) : null}
        <RadixText as="p" size="2" color="gray">
          <FormattedMessage id="SubscriptionScreen.planDetails.reminder" />
        </RadixText>
      </div>
      <CurrentPlanCard
        planName={planName}
        thisMonthUsage={thisMonthUsage}
        thisMonthSubscriptionUsage={thisMonthSubscriptionUsage}
        previousMonthSubscriptionUsage={previousMonthSubscriptionUsage}
        hasSubscription={hasSubscription}
      />
      {formattedBillingDate != null ? (
        <TextButton
          variant="default"
          size="3"
          loading={manageSubscriptionLoading}
          disabled={isLoading}
          text={
            <FormattedMessage id="SubscriptionScreen.footer.manageSubscription" />
          }
          onClick={onClickManageSubscription}
        />
      ) : null}
    </div>
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
      <div className={styles.processingPaymentRoot}>
        <RadixText as="p" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="SubscriptionScreen.title" />
        </RadixText>
        <div className={cn(styles.processingPaymentSection)}>
          {paymentStatus === "IsProcessing" ? (
            <>
              <div className={styles.processingPaymentSpinner}>
                <Spinner size="3" />
                <RadixText
                  as="span"
                  className={styles.processingPaymentSpinnerLabel}
                >
                  {renderToString("SubscriptionScreen.processing-payment")}
                </RadixText>
              </div>
              <Button asChild={true} size="2" variant="outline" color="gray">
                <a href="mailto:hello@authgear.com">
                  <FormattedMessage id="SubscriptionScreen.contact-us.label" />
                </a>
              </Button>
            </>
          ) : null}
          {paymentStatus === "CardDeclined" ? (
            <>
              <RadixText className={styles.processingPaymentErrorMessage}>
                <FormattedMessage id="SubscriptionScreen.payment-declined.description" />
              </RadixText>
              <div className={styles.processingPaymentButtonContainer}>
                <Button
                  size="2"
                  loading={cancelFailedSubscriptionLoading}
                  // eslint-disable-next-line @typescript-eslint/strict-void-return
                  onClick={onClickCancelFailedSubscription}
                >
                  <FormattedMessage id="SubscriptionScreen.cancel-transaction.label" />
                </Button>
              </div>
            </>
          ) : null}
          {paymentStatus === "UnknownError" ? (
            <>
              <RadixText className={styles.processingPaymentErrorMessage}>
                <FormattedMessage id="SubscriptionScreen.unknown-error.description" />
              </RadixText>
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
  const [now] = useState(() => new Date());
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
  const thisMonthSubscriptionUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).thisMonthSubscriptionUsage;
  const previousMonthSubscriptionUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).previousMonthSubscriptionUsage;
  const thisMonthUsage = (
    subscriptionScreenQuery.data?.node as AppFragmentFragment
  ).thisMonthUsage;

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
        thisMonthSubscriptionUsage={thisMonthSubscriptionUsage ?? undefined}
        previousMonthSubscriptionUsage={
          previousMonthSubscriptionUsage ?? undefined
        }
        effectiveAppConfig={effectiveAppConfig ?? undefined}
      />
    </ScreenLayoutScrollView>
  );
};

export default SubscriptionScreen;
