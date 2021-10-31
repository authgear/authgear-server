import React, { useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { PrimaryButton, Text } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./SubscriptionScreen.module.scss";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";

const contactUsLink = "https://oursky.typeform.com/to/PecQiGfc";

const paidPlansForBlocksList = ["startups", "business", "enterprise"];

function isCustomPlan(planName: string): boolean {
  return (
    ["free", "startups", "business", "enterprise"].indexOf(planName) === -1
  );
}
interface SubscriptionPlanSummaryProps {
  planName: string;
}

const SubscriptionPlanSummary: React.FC<SubscriptionPlanSummaryProps> =
  function SubscriptionPlanSummary(props) {
    const { planName } = props;
    const isCustom = useMemo(() => {
      return isCustomPlan(planName);
    }, [planName]);

    return (
      <Widget className={styles.widget}>
        <WidgetTitle>
          <FormattedMessage id="SubscriptionScreen.current-plan-summary.title" />
        </WidgetTitle>
        <div className={styles.summaryList}>
          <div className={styles.summaryItem}>
            <Text variant="medium" block={true} className={styles.label}>
              <FormattedMessage id="SubscriptionScreen.subscription.label" />
            </Text>
            <Text variant="xLarge" block={true}>
              {isCustom ? (
                <>{planName}</>
              ) : (
                <FormattedMessage
                  id={`SubscriptionScreen.plan-name.${planName}`}
                />
              )}
            </Text>
          </div>
          <div className={styles.summaryItem}>
            <Text variant="medium" block={true} className={styles.label}>
              <FormattedMessage id="SubscriptionScreen.monthly-active-users.label" />
            </Text>
            <Text variant="xLarge" block={true}>
              {isCustom ? (
                <>-</>
              ) : (
                <FormattedMessage
                  id={`SubscriptionScreen.summary-mau.${planName}`}
                />
              )}
            </Text>
          </div>
          <div className={styles.summaryItem}>
            <Text variant="medium" block={true} className={styles.label}>
              <FormattedMessage id="SubscriptionScreen.plan-cost.label" />
            </Text>
            <Text variant="xLarge">
              {isCustom ? (
                <>-</>
              ) : (
                <FormattedMessage
                  id={`SubscriptionScreen.summary-price.${planName}`}
                />
              )}
            </Text>
          </div>
        </div>
      </Widget>
    );
  };

interface SubscriptionPlanBlockProps {
  planName: string;
}

const SubscriptionPlanBlock: React.FC<SubscriptionPlanBlockProps> =
  function SubscriptionPlanBlock(props) {
    const { planName } = props;
    const { renderToString } = useContext(Context);

    const planCopy = useMemo(() => {
      return {
        title: renderToString(`SubscriptionScreen.plan-name.${planName}`),
        desc: renderToString(`SubscriptionScreen.des.${planName}`),
        price: renderToString(`SubscriptionScreen.price.${planName}`),
        mau: renderToString(`SubscriptionScreen.mau.${planName}`),
        extraCost: renderToString(`SubscriptionScreen.extra-cost.${planName}`),
        features: renderToString(`SubscriptionScreen.features.${planName}`),
      };
    }, [planName, renderToString]);

    return (
      <div className={styles.planBlockItem}>
        <Text as="h3" variant="xLarge" block={true} className={styles.title}>
          {planCopy.title}
        </Text>
        <Text variant="medium" block={true} className={styles.desc}>
          {planCopy.desc}
        </Text>
        <Text variant="xLarge" block={true} className={styles.price}>
          {planCopy.price}
        </Text>
        <Text variant="medium" block={true} className={styles.mau}>
          {planCopy.mau}
        </Text>
        {planCopy.extraCost && (
          <Text variant="medium" block={true} className={styles.extraCost}>
            {planCopy.extraCost}
          </Text>
        )}
        <div className={styles.separator} />
        <Text variant="medium" block={true} className={styles.features}>
          {planCopy.features}
        </Text>
        <PrimaryButton
          target="_blank"
          rel="noreferrer"
          href={contactUsLink}
          className={styles.contactUsLink}
        >
          <FormattedMessage id="SubscriptionScreen.contact-us.label" />
        </PrimaryButton>
      </div>
    );
  };

const SubscriptionPlanBlocks: React.FC = function SubscriptionPlanBlocks() {
  return (
    <Widget className={styles.widget}>
      <WidgetTitle>
        <FormattedMessage id="SubscriptionScreen.upgrade-your-plan.title" />
      </WidgetTitle>
      <div className={styles.planBlockList}>
        {paidPlansForBlocksList.map((planName) => (
          <SubscriptionPlanBlock key={planName} planName={planName} />
        ))}
      </div>
    </Widget>
  );
};

const SubscriptionPlanFeatures: React.FC = function SubscriptionPlanFeatures() {
  return (
    <div className={styles.widget}>
      <Text as="p" variant="medium" block={true}>
        <FormattedMessage id="SubscriptionScreen.plan-features.desc" />
      </Text>
    </div>
  );
};

const SubscriptionScreen: React.FC = function SubscriptionScreen() {
  const { appID } = useParams();
  const featureConfig = useAppFeatureConfigQuery(appID);

  const planName = useMemo(
    () => featureConfig.planName ?? "-",
    [featureConfig.planName]
  );

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

  return (
    <ScreenContent>
      <ScreenTitle className={styles.widget}>
        <FormattedMessage id="SubscriptionScreen.title" />
      </ScreenTitle>
      <SubscriptionPlanSummary planName={planName} />
      <SubscriptionPlanBlocks />
      <SubscriptionPlanFeatures />
    </ScreenContent>
  );
};

export default SubscriptionScreen;
