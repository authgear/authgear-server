import React, { useCallback, useContext, useMemo, useState } from "react";
import { useConst } from "@fluentui/react-hooks";
import { ICommandBarItemProps, Text, useTheme } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import { useAnalyticChartsQuery } from "./query/analyticChartsQuery";
import { Periodical } from "./globalTypes.generated";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import AnalyticsActivityWidget from "./AnalyticsActivityWidget";
import AnalyticsSignupConversionWidget from "./AnalyticsSignupConversionWidget";
import AnalyticsSignupMethodsWidget from "./AnalyticsSignupMethodsWidget";
import ShowError from "../../ShowError";
import CommandBarContainer from "../../CommandBarContainer";
import styles from "./AnalyticsScreen.module.css";
import useTransactionalState from "../../hook/useTransactionalState";
import DateRangeDialog from "./DateRangeDialog";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { parseDate } from "../../util/date";
import ScreenDescription from "../../ScreenDescription";
import CommandBarButton from "../../CommandBarButton";

function truncateTimeAndReplaceTimezoneToUTC(date: Date): Date {
  return new Date(
    Date.UTC(date.getFullYear(), date.getMonth(), date.getDate())
  );
}

const CommandBarLabelValue = (label: string, value: string) => {
  return (props: ICommandBarItemProps) => {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    const theme = useTheme();
    return (
      <CommandBarButton
        {...props.commandBarButtonProps}
        className={styles.commandBarButtonLabelValue}
        text={
          <>
            <span className={styles.label}>{label}</span>
            <span
              className={styles.value}
              style={{ color: theme.palette.neutralTertiary }}
            >
              {value}
            </span>
          </>
        }
        styles={{
          // https://github.com/authgear/authgear-server/issues/2348#issuecomment-1226545493
          label: {
            whiteSpace: "nowrap",
          },
        }}
      />
    );
  };
};

const AnalyticsScreenContent: React.VFC = function AnalyticsScreenContent() {
  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);

  const { analyticEpoch: analyticEpochStr } = useSystemConfig();
  const analyticEpochDate = useMemo(() => {
    if (analyticEpochStr === "") {
      return undefined;
    }
    return parseDate(analyticEpochStr);
  }, [analyticEpochStr]);

  const today = useConst(new Date(Date.now()));
  const yesterday = useMemo(() => {
    // yesterday
    const d = new Date(
      Date.UTC(
        today.getUTCFullYear(),
        today.getUTCMonth(),
        today.getUTCDate() - 1
      )
    );
    return d;
  }, [today]);

  const defaultRangeTo = useMemo(() => yesterday, [yesterday]);
  const defaultRangeFrom = useMemo(() => {
    // default 1 year range
    const d = new Date(
      Date.UTC(
        defaultRangeTo.getUTCFullYear() - 1,
        defaultRangeTo.getUTCMonth(),
        defaultRangeTo.getUTCDate()
      )
    );
    if (analyticEpochDate && analyticEpochDate > d) {
      return analyticEpochDate;
    }
    return d;
  }, [defaultRangeTo, analyticEpochDate]);

  const {
    committedValue: rangeFrom,
    uncommittedValue: uncommittedRangeFrom,
    setValue: setRangeFrom,
    setCommittedValue: setRangeFromImmediately,
    commit: commitRangeFrom,
    rollback: rollbackRangeFrom,
  } = useTransactionalState<Date | null>(defaultRangeFrom);

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTransactionalState<Date | null>(defaultRangeTo);

  const minDate = useMemo(() => {
    return analyticEpochDate;
  }, [analyticEpochDate]);

  const maxDate = useMemo(() => {
    return yesterday;
  }, [yesterday]);

  const rangeToStr = useMemo(() => {
    return rangeTo ? rangeTo.toISOString().split("T")[0] : "";
  }, [rangeTo]);

  const rangeFromStr = useMemo(() => {
    return rangeFrom ? rangeFrom.toISOString().split("T")[0] : "";
  }, [rangeFrom]);

  const [periodical, setPeriodical] = useState<Periodical>(Periodical.Monthly);

  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const {
    loading,
    error,
    refetch,
    activeUserChart,
    totalUserCountChart,
    signupConversionRate,
    signupByMethodsChart,
  } = useAnalyticChartsQuery(appID, periodical, rangeFromStr, rangeToStr);

  const onClickDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(false);
    },
    []
  );

  const onClickResetDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setRangeFromImmediately(defaultRangeFrom);
      setRangeToImmediately(defaultRangeTo);
    },
    [
      setRangeFromImmediately,
      setRangeToImmediately,
      defaultRangeFrom,
      defaultRangeTo,
    ]
  );

  const primaryItems: ICommandBarItemProps[] = useMemo(() => {
    return [
      {
        key: "startDate",
        text: `${renderToString(
          "AnalyticsScreen.start-date.label"
        )} ${rangeFromStr}`,
        iconProps: { iconName: "Calendar" },
        commandBarButtonAs: CommandBarLabelValue(
          renderToString("AnalyticsScreen.start-date.label"),
          rangeFromStr
        ),
        commandBarButtonProps: {
          iconProps: { iconName: "Calendar" },
          onClick: onClickDateRange,
        },
      },
      {
        key: "endDate",
        text: `${renderToString(
          "AnalyticsScreen.end-date.label"
        )} ${rangeToStr}`,
        iconProps: { iconName: "Calendar" },
        commandBarButtonAs: CommandBarLabelValue(
          renderToString("AnalyticsScreen.end-date.label"),
          rangeToStr
        ),
        commandBarButtonProps: {
          iconProps: { iconName: "Calendar" },
          onClick: onClickDateRange,
        },
      },
      {
        key: "reset",
        text: renderToString("AnalyticsScreen.clear-date-range.label"),
        iconProps: { iconName: "ClearFilter" },
        onClick: onClickResetDateRange,
      },
    ];
  }, [
    renderToString,
    rangeFromStr,
    rangeToStr,
    onClickDateRange,
    onClickResetDateRange,
  ]);

  const onDismissDateRangeDialog = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      rollbackRangeFrom();
      rollbackRangeTo();
    },
    [setDateRangeDialogHidden, rollbackRangeFrom, rollbackRangeTo]
  );

  const commitDateRange = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.preventDefault();
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      commitRangeFrom();
      commitRangeTo();
    },
    [setDateRangeDialogHidden, commitRangeFrom, commitRangeTo]
  );

  const onSelectRangeFrom = useCallback(
    (value: Date | null | undefined) => {
      if (value == null) {
        setRangeFrom(null);
      } else {
        // the date from the date picker is 0:0:0 in the user timezone
        // in analytics page context, all data are in UTC
        // so we set the time of the date object to UTC 0:0:0
        value = truncateTimeAndReplaceTimezoneToUTC(value);
        if (uncommittedRangeTo == null) {
          setRangeFrom(value);
        } else if (value > uncommittedRangeTo) {
          setRangeTo(value);
          setRangeFrom(uncommittedRangeTo);
        } else {
          setRangeFrom(value);
          // bound date range within 1 year
          let limitRangeTo = new Date(
            Date.UTC(
              value.getUTCFullYear() + 1,
              value.getUTCMonth(),
              value.getUTCDate()
            )
          );
          if (limitRangeTo > yesterday) {
            limitRangeTo = yesterday;
          }
          if (uncommittedRangeTo > limitRangeTo) {
            setRangeTo(limitRangeTo);
          }
        }
      }
    },
    [setRangeFrom, setRangeTo, uncommittedRangeTo, yesterday]
  );

  const onSelectRangeTo = useCallback(
    (value: Date | null | undefined) => {
      if (value == null) {
        setRangeTo(null);
      } else {
        // the date from the date picker is 0:0:0 in the user timezone
        // in analytics page context, all data are in UTC
        // so we set the time of the date object to UTC 0:0:0
        value = truncateTimeAndReplaceTimezoneToUTC(value);
        if (uncommittedRangeFrom == null) {
          setRangeTo(value);
        } else if (value < uncommittedRangeFrom) {
          setRangeFrom(value);
          setRangeTo(uncommittedRangeFrom);
        } else {
          setRangeTo(value);

          // bound date range within 1 year and before epoch date
          let limitRangeFrom = new Date(
            Date.UTC(
              value.getUTCFullYear() - 1,
              value.getUTCMonth(),
              value.getUTCDate()
            )
          );
          if (analyticEpochDate != null && limitRangeFrom < analyticEpochDate) {
            limitRangeFrom = analyticEpochDate;
          }
          if (uncommittedRangeFrom < limitRangeFrom) {
            setRangeFrom(limitRangeFrom);
          }
        }
      }
    },
    [setRangeTo, setRangeFrom, uncommittedRangeFrom, analyticEpochDate]
  );

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <>
      <CommandBarContainer isLoading={loading} primaryItems={primaryItems}>
        <ScreenContent layout="list">
          <ScreenTitle className={styles.widget}>
            <FormattedMessage id="AnalyticsScreen.title" />
          </ScreenTitle>
          <ScreenDescription className={styles.widget}>
            <FormattedMessage id="AnalyticsScreen.description" />
          </ScreenDescription>
          <AnalyticsActivityWidget
            className={styles.activityWidget}
            loading={loading}
            periodical={periodical}
            onPeriodicalChange={setPeriodical}
            activeUserChartData={activeUserChart}
            totalUserCountChartData={totalUserCountChart}
          />
          <AnalyticsSignupConversionWidget
            className={styles.signupConversionWidget}
            loading={loading}
            signupConversionRate={signupConversionRate}
          />
          <AnalyticsSignupMethodsWidget
            className={styles.signupMethodsWidget}
            loading={loading}
            signupByMethodsChart={signupByMethodsChart}
          />
        </ScreenContent>
      </CommandBarContainer>
      <DateRangeDialog
        hidden={dateRangeDialogHidden}
        title={renderToString("AnalyticsScreen.date-range.dialog-title")}
        fromDatePickerLabel={renderToString(
          "AuditLogScreen.date-range.start-date"
        )}
        toDatePickerLabel={renderToString("AuditLogScreen.date-range.end-date")}
        rangeFrom={uncommittedRangeFrom ?? undefined}
        rangeTo={uncommittedRangeTo ?? undefined}
        fromDatePickerMinDate={minDate}
        fromDatePickerMaxDate={maxDate}
        toDatePickerMinDate={minDate}
        toDatePickerMaxDate={maxDate}
        onSelectRangeFrom={onSelectRangeFrom}
        onSelectRangeTo={onSelectRangeTo}
        onCommitDateRange={commitDateRange}
        onDismiss={onDismissDateRangeDialog}
      />
    </>
  );
};

const AnalyticsScreen: React.VFC = function AnalyticsScreen() {
  const { analyticEnabled } = useSystemConfig();

  if (!analyticEnabled) {
    return <Text>Analytics page is disabled.</Text>;
  }

  return <AnalyticsScreenContent />;
};

export default AnalyticsScreen;
