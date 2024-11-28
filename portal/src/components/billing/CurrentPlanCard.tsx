import React, { useMemo } from "react";
import {
  IButtonProps,
  ITooltipHostProps,
  PartialTheme,
  ProgressIndicator,
  Text,
  ThemeProvider,
  TooltipHost,
  useTheme,
} from "@fluentui/react";
import styles from "./CurrentPlanCard.module.css";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useId } from "@fluentui/react-hooks";
import LinkButton from "../../LinkButton";

interface CurrentPlanCardProps {}

export function CurrentPlanCard({}: CurrentPlanCardProps): React.ReactElement {
  return (
    <div className={styles.cardContainer}>
      <FixedCostSection />
      <MeteredCostSection />
      <MAUUsageSection />
    </div>
  );
}

function CostItemRow({
  label,
  value,
}: {
  label: React.ReactNode;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-end justify-between">
      <Text variant="medium" className="font-semibold">
        {label}
      </Text>
      <Text variant="medium">{value}</Text>
    </div>
  );
}

function FixedCostSection() {
  return (
    <section className={styles.card}>
      <div className="space-y-2">
        <Text block={true} variant="mediumPlus" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.subscriptionFee.title" />
        </Text>
        <div className="flex items-end">
          <Text variant="xxLarge">
            <FormattedMessage
              id="CurrentPlanCard.subscriptionFee.value"
              values={{ price: 50 }} // FIXME
            />
          </Text>
          <Text variant="large" className="ml-2 font-semibold">
            <FormattedMessage id="CurrentPlanCard.subscriptionFee.unit" />
          </Text>
        </div>
      </div>
      <div className="space-y-2">
        <Text block={true} variant="medium" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.subscriptionFee.include" />
        </Text>
        <CostItemRow
          label={
            <FormattedMessage
              id="CurrentPlanCard.subscriptionFee.plan"
              values={{ plan: "Developers" }} // FIXME
            />
          }
          value={
            <FormattedMessage
              id="CurrentPlanCard.subscriptionFee.planPrice"
              values={{ price: 50 }} // FIXME
            />
          }
        />
      </div>
    </section>
  );
}

function MeteredCostSection() {
  return (
    <section className={styles.card}>
      <div className="space-y-2">
        <Text block={true} variant="mediumPlus" className="font-semibold">
          <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.title" />
        </Text>
        <div className="flex items-end">
          <Text variant="xxLarge">
            <FormattedMessage
              id="CurrentPlanCard.whatsappSMSFee.value"
              values={{ price: 11 }} // FIXME
            />
          </Text>
          <Text variant="large" className="ml-2 font-semibold">
            <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.unit" />
          </Text>
        </div>
      </div>
      <div className="space-y-2">
        <CostItemRow
          label={
            <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.sms.northAmerica" />
          }
          value={
            <FormattedMessage
              id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
              values={{ unitPrice: 0.01, quantity: 10, total: 1000 }} // FIXME
            />
          }
        />
        <CostItemRow
          label={
            <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.sms.other" />
          }
          value={
            <FormattedMessage
              id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
              values={{ unitPrice: 0.01, quantity: 10, total: 1000 }} // FIXME
            />
          }
        />
        <CostItemRow
          label={
            <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.whatsapp.northAmerica" />
          }
          value={
            <FormattedMessage
              id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
              values={{ unitPrice: 0.01, quantity: 10, total: 1000 }} // FIXME
            />
          }
        />
        <CostItemRow
          label={
            <FormattedMessage id="CurrentPlanCard.whatsappSMSFee.whatsapp.other" />
          }
          value={
            <FormattedMessage
              id="CurrentPlanCard.whatsappSMSFee.whatsappSMSPrice"
              values={{ unitPrice: 0.01, quantity: 10, total: 1000 }} // FIXME
            />
          }
        />
      </div>
    </section>
  );
}

function MAUUsageSection() {
  return (
    <section className={styles.card}>
      <UsageMeter
        title={<FormattedMessage id="CurrentPlanCard.mau.title" />}
        current={523}
        limit={5000}
        previous={46}
        warnPercentage={0.8}
        tooltip={<FormattedMessage id="CurrentPlanCard.mau.tooltip" />}
      />
    </section>
  );
}

interface UsageMeterProps {
  title: React.ReactNode;
  tooltip: ITooltipHostProps["content"];
  current?: number;
  limit?: number;
  previous?: number;
  warnPercentage: number;
  onClickUpgrade?: IButtonProps["onClick"];
}

const USAGE_METER_THEME_WARN: PartialTheme = {
  palette: {
    themePrimary: "#F9597A",
  },
};

function UsageMeter(props: UsageMeterProps): React.ReactElement {
  const {
    title,
    tooltip,
    current,
    limit,
    previous,
    warnPercentage,
    onClickUpgrade,
  } = props;
  const percentComplete =
    current != null && limit != null ? current / limit : 0;
  const id = useId("usage-meter");
  const calloutProps = useMemo(() => {
    return {
      target: `#${id}`,
    };
  }, [id]);
  const currentTheme = useTheme();
  const limitReached =
    current != null && limit != null ? current >= limit : false;
  const theme = limitReached ? USAGE_METER_THEME_WARN : currentTheme;
  const usageStyles = {
    root: {
      color: limitReached
        ? USAGE_METER_THEME_WARN.palette?.themePrimary
        : currentTheme.palette.neutralSecondary,
    },
  };
  return (
    <TooltipHost
      hostClassName="col-span-2"
      content={tooltip}
      calloutProps={calloutProps}
    >
      <div className="flex flex-col">
        <Text
          id={id}
          block={true}
          variant="mediumPlus"
          className="self-start font-semibold mb-2"
        >
          {title}
        </Text>
        <ThemeProvider theme={theme}>
          <ProgressIndicator
            className="w-full"
            percentComplete={percentComplete}
          />
          <Text block={true} styles={usageStyles} variant="medium">
            {current != null ? `${current}` : "-"}
            {" / "}
            {limit != null ? `${limit}` : "-"}
            {previous != null ? (
              <FormattedMessage
                id="CurrentPlanCard.mau.previous"
                values={{
                  count: previous,
                }}
              />
            ) : null}
          </Text>
          {limitReached ? (
            <LinkButton onClick={onClickUpgrade}>
              <FormattedMessage id="CurrentPlanCard.mau.limitReached" />
            </LinkButton>
          ) : percentComplete >= warnPercentage ? (
            <LinkButton onClick={onClickUpgrade}>
              <FormattedMessage id="CurrentPlanCard.mau.approachingLimit" />
            </LinkButton>
          ) : null}
        </ThemeProvider>
      </div>
    </TooltipHost>
  );
}
