import React from "react";
import { useParams } from "react-router-dom";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./SubscriptionScreen.module.scss";
import SubscriptionCurrentPlanSummary, {
  CostItem,
  CostItemSeparator,
} from "./SubscriptionCurrentPlanSummary";

const DEFAULT_PLAN_NAME = "free";

const contactUsLink = "https://oursky.typeform.com/to/PecQiGfc";

const SubscriptionScreen: React.FC = function SubscriptionScreen() {
  const { appID } = useParams() as { appID: string };
  const featureConfig = useAppFeatureConfigQuery(appID);

  if (featureConfig.loading) {
    return <ShowLoading />;
  }

  if (featureConfig.error) {
    return (
      <ShowError
        error={featureConfig.error}
        onRetry={() => {
          featureConfig.refetch().finally(() => {});
        }}
      />
    );
  }

  const planName = featureConfig.planName ?? DEFAULT_PLAN_NAME;

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
};

export default SubscriptionScreen;
