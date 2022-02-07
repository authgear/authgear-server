import React, { useCallback, useContext, useMemo, useState } from "react";
import { useConst } from "@fluentui/react-hooks";
import { CommandBarButton, ICommandBarItemProps } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useAnalyticChartsQuery } from "./query/analyticChartsQuery";
import { Periodical } from "./__generated__/globalTypes";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import AnalyticsActivityWidget from "./AnalyticsActivityWidget";
import ShowError from "../../ShowError";
import CommandBarContainer from "../../CommandBarContainer";
import styles from "./AnalyticsScreen.module.scss";
import useTransactionalState from "../../hook/useTransactionalState";
import DateRangeDialog from "./DateRangeDialog";

const CommandBarLabelValue = (label: string, value: string) => {
  return (props: ICommandBarItemProps) => {
    const { commandBarButtonProps } = props;
    return (
      <CommandBarButton
        className={styles.commandBarButtonLabelValue}
        {...commandBarButtonProps}
      >
        <span className={styles.label}>{label}</span>
        <span className={styles.value}>{value}</span>
      </CommandBarButton>
    );
  };
};

const OnRenderCommandBarToLabel = () => {
  return (
    <div className={styles.commandBarButtonTo}>
      <FormattedMessage id="AnalyticsScreen.to.label"></FormattedMessage>
    </div>
  );
};

const AnalyticsScreen: React.FC = function AnalyticsScreen() {
  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);

  const today = useConst(new Date(Date.now()));
  const defaultRangeTo = useMemo(() => {
    // yesterday
    const d = new Date(
      Date.UTC(today.getUTCFullYear(), today.getUTCMonth(), today.getUTCDate())
    );
    d.setDate(d.getDate() - 1);
    return d;
  }, [today]);

  const defaultRangeFrom = useMemo(() => {
    // default 1 year range
    const d = new Date(defaultRangeTo);
    d.setFullYear(d.getFullYear() - 1);
    return d;
  }, [defaultRangeTo]);

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
    // FIXME: minDate should respect analytic epoch
    return defaultRangeFrom;
  }, [defaultRangeFrom]);

  const maxDate = useMemo(() => {
    return defaultRangeTo;
  }, [defaultRangeTo]);

  const rangeToStr = useMemo(() => {
    return rangeTo ? rangeTo.toISOString().split("T")[0] : "";
  }, [rangeTo]);

  const rangeFromStr = useMemo(() => {
    return rangeFrom ? rangeFrom.toISOString().split("T")[0] : "";
  }, [rangeFrom]);

  const [periodical, setPeriodical] = useState<Periodical>(Periodical.MONTHLY);

  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const { loading, error, refetch, activeUserChart, totalUserCountChart } =
    useAnalyticChartsQuery(appID, periodical, rangeFromStr, rangeToStr);

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
        key: "to",
        onRender: OnRenderCommandBarToLabel,
      },
      {
        key: "endDate",
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
        if (uncommittedRangeTo != null && value > uncommittedRangeTo) {
          setRangeTo(value);
          setRangeFrom(uncommittedRangeTo);
        } else {
          setRangeFrom(value);
        }
      }
    },
    [setRangeFrom, setRangeTo, uncommittedRangeTo]
  );

  const onSelectRangeTo = useCallback(
    (value: Date | null | undefined) => {
      if (value == null) {
        setRangeTo(null);
      } else {
        if (uncommittedRangeFrom != null && value < uncommittedRangeFrom) {
          setRangeFrom(value);
          setRangeTo(uncommittedRangeFrom);
        } else {
          setRangeTo(value);
        }
      }
    },
    [setRangeTo, setRangeFrom, uncommittedRangeFrom]
  );

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <>
      <CommandBarContainer isLoading={loading} primaryItems={primaryItems}>
        <ScreenContent>
          <ScreenTitle>
            <FormattedMessage id="AnalyticsScreen.title" />
          </ScreenTitle>
          <AnalyticsActivityWidget
            loading={loading}
            periodical={periodical}
            onPeriodicalChange={setPeriodical}
            activeUserChartData={activeUserChart}
            totalUserCountChartData={totalUserCountChart}
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

export default AnalyticsScreen;
