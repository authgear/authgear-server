import React, { useContext, useMemo } from "react";
import cn from "classnames";
import {
  DefaultEffects,
  Text,
  useTheme,
  TooltipHost,
  ITooltipHostProps,
  Link,
  IButtonProps,
  ProgressIndicator,
  PartialTheme,
  ThemeProvider,
} from "@fluentui/react";
import { useId } from "@fluentui/react-hooks";
import { DateTime } from "luxon";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { formatDatetime } from "../../util/formatDatetime";
import styles from "./SubscriptionCurrentPlanSummary.module.scss";

export interface SubscriptionCurrentPlanSummaryProps {
  className?: string;
  planName: string;
  baseAmount?: number;
  mauCurrent?: number;
  mauLimit?: number;
  mauPrevious?: number;
  nextBillingDate?: Date;
  children?: React.ReactNode;
}

interface TitleProps {
  planName: string;
  baseAmount?: number;
}

function Title(props: TitleProps) {
  const { planName, baseAmount } = props;
  const { renderToString } = useContext(Context);
  const planDisplayName =
    baseAmount == null
      ? planName
      : renderToString("SubscriptionScreen.plan-name." + planName);
  return (
    <Text block={true} variant="xLarge">
      {baseAmount == null ? (
        <FormattedMessage
          id="SubscriptionCurrentPlanSummary.title.custom-plan"
          values={{
            name: planDisplayName,
          }}
        />
      ) : (
        <FormattedMessage
          id="SubscriptionCurrentPlanSummary.title.known-plan"
          values={{
            name: planDisplayName,
            amount: baseAmount / 100,
          }}
        />
      )}
    </Text>
  );
}

export function CostItemSeparator(): React.ReactElement {
  const theme = useTheme();
  return (
    <div
      className={styles.costItemSeparator}
      style={{
        backgroundColor: theme.semanticColors.bodyDivider,
      }}
    />
  );
}

export interface CostItemProps {
  title: React.ReactNode;
  kind: "free" | "upgrade" | "billed" | "non-applicable";
  tooltip?: ITooltipHostProps["content"];
  amount?: number;
  onClickUpgrade?: IButtonProps["onClick"];
}

export function CostItem(props: CostItemProps): React.ReactElement {
  const { title, kind, tooltip, amount, onClickUpgrade } = props;
  const id = useId("cost-item");
  const calloutProps = useMemo(() => {
    return {
      target: `#${id}`,
    };
  }, [id]);
  const children = (
    <>
      <Text id={id} block={true} className={styles.costItemTitle}>
        {title}
      </Text>
      <Text block={true} variant="xLarge">
        {kind === "non-applicable" ? (
          "-"
        ) : kind === "free" ? (
          <FormattedMessage id="SubscriptionCurrentPlanSummary.label.free" />
        ) : kind === "upgrade" ? (
          <Link onClick={onClickUpgrade}>
            <FormattedMessage id="SubscriptionCurrentPlanSummary.label.upgrade" />
          </Link>
        ) : (
          <>{`$${(amount ?? 0) / 100}`}</>
        )}
      </Text>
    </>
  );
  if (tooltip == null) {
    return <div>{children}</div>;
  }
  return (
    <TooltipHost content={tooltip} calloutProps={calloutProps}>
      {children}
    </TooltipHost>
  );
}

interface CostItemsProps {
  children?: React.ReactNode;
}

function CostItems(props: CostItemsProps) {
  const { children } = props;
  return <div className={styles.costItems}>{children}</div>;
}

const USAGE_METER_THEME_WARN: PartialTheme = {
  palette: {
    themePrimary: "#F9597A",
  },
};

interface UsageMeterProps {
  title: React.ReactNode;
  tooltip: ITooltipHostProps["content"];
  current?: number;
  limit?: number;
  previous?: number;
  warnPercentage: number;
  onClickUpgrade?: IButtonProps["onClick"];
}

// eslint-disable-next-line complexity
function UsageMeter(props: UsageMeterProps) {
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
    <TooltipHost content={tooltip} calloutProps={calloutProps}>
      <div className={styles.usageMeter}>
        <Text
          id={id}
          block={true}
          variant="small"
          className={styles.usageMeterTitle}
        >
          {title}
        </Text>
        <ThemeProvider theme={theme}>
          <ProgressIndicator
            className={styles.usageMeterProgressBar}
            percentComplete={percentComplete}
          />
          <Text block={true} styles={usageStyles} variant="small">
            {limit != null && current != null
              ? `${current} / ${limit}`
              : "- / -"}
            {previous != null ? (
              <FormattedMessage
                id="SubscriptionCurrentPlanSummary.mau.previous"
                values={{
                  count: previous,
                }}
              />
            ) : null}
          </Text>
          {limitReached ? (
            <Link onClick={onClickUpgrade}>
              <FormattedMessage id="SubscriptionCurrentPlanSummary.mau.limit-reached" />
            </Link>
          ) : percentComplete >= warnPercentage ? (
            <Link onClick={onClickUpgrade}>
              <FormattedMessage id="SubscriptionCurrentPlanSummary.mau.approaching-limit" />
            </Link>
          ) : (
            <Text block={true}>{"\u00a0"}</Text>
          )}
        </ThemeProvider>
      </div>
    </TooltipHost>
  );
}

interface SubscriptionManagementProps {
  nextBillingDate?: Date;
}

function SubscriptionManagement(props: SubscriptionManagementProps) {
  const { locale } = useContext(Context);
  const theme = useTheme();
  const { nextBillingDate } = props;
  const formattedDate = formatDatetime(
    locale,
    nextBillingDate ?? null,
    DateTime.DATE_SHORT
  );
  return (
    <div className={styles.subscriptionManagement}>
      {formattedDate != null ? (
        <Text
          block={true}
          styles={{
            root: {
              color: theme.palette.neutralSecondary,
            },
          }}
        >
          <FormattedMessage
            id="SubscriptionCurrentPlanSummary.next-billing-date"
            values={{
              date: formattedDate,
            }}
          />
        </Text>
      ) : null}
      <Link className={styles.subscriptionManagementLink}>
        <FormattedMessage id="SubscriptionCurrentPlanSummary.view-invoices" />
      </Link>
      <Link className={styles.subscriptionManagementLink}>
        <FormattedMessage id="SubscriptionCurrentPlanSummary.change-billing-methods" />
      </Link>
    </div>
  );
}

function SubscriptionCurrentPlanSummary(
  props: SubscriptionCurrentPlanSummaryProps
): React.ReactElement | null {
  const {
    className,
    planName,
    baseAmount,
    mauCurrent,
    mauLimit,
    mauPrevious,
    nextBillingDate,
    children,
  } = props;
  return (
    <div
      className={cn(className, styles.root)}
      style={{
        boxShadow: DefaultEffects.elevation4,
      }}
    >
      <Title planName={planName} baseAmount={baseAmount} />
      <CostItems>{children}</CostItems>
      <div className={styles.usageMeterContainer}>
        <UsageMeter
          title={
            <FormattedMessage id="SubscriptionCurrentPlanSummary.mau.title" />
          }
          current={mauCurrent}
          limit={mauLimit}
          previous={mauPrevious}
          warnPercentage={0.8}
          tooltip={
            <FormattedMessage id="SubscriptionCurrentPlanSummary.mau.tooltip" />
          }
        />
        {baseAmount != null ? (
          <SubscriptionManagement nextBillingDate={nextBillingDate} />
        ) : null}
      </div>
    </div>
  );
}

export default SubscriptionCurrentPlanSummary;
