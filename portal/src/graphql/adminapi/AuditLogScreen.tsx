import React, { useState, useMemo, useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import {
  ICommandBarItemProps,
  IDropdownOption,
  MessageBar,
  addDays,
  TooltipHost,
  ITooltipHostStyles,
  ITooltipProps,
  CommandBarButton,
  DirectionalHint,
} from "@fluentui/react";
import { useId } from "@fluentui/react-hooks";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { useQuery } from "@apollo/client";
import { DateTime } from "luxon";
import NavBreadcrumb from "../../NavBreadcrumb";
import AuditLogList from "./AuditLogList";
import CommandBarDropdown, {
  CommandBarDropdownProps,
} from "../../CommandBarDropdown";
import CommandBarContainer from "../../CommandBarContainer";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import DateRangeDialog from "../portal/DateRangeDialog";
import { encodeOffsetToCursor } from "../../util/pagination";
import useTransactionalState from "../../hook/useTransactionalState";
import {
  AuditLogListQueryQuery,
  AuditLogListQueryQueryVariables,
  AuditLogListQueryDocument,
} from "./query/auditLogListQuery.generated";
import { AuditLogActivityType, SortDirection } from "./globalTypes.generated";
import styles from "./AuditLogScreen.module.scss";
import { useAppFeatureConfigQuery } from "../portal/query/appFeatureConfigQuery";

const pageSize = 10;

function CommandBarDropdownWrapper(props: ICommandBarItemProps) {
  const { dropdownProps } = props;
  return <CommandBarDropdown {...dropdownProps} />;
}

function RefreshButton(props: ICommandBarItemProps) {
  const tooltipStyle: Partial<ITooltipHostStyles> = {
    root: { display: "inline-block" },
  };
  const tooltipId = useId("refreshTooltip");
  const tooltipCalloutProps = {
    gapSpace: 0,
  };

  const { renderToString, locale } = useContext(Context);

  const tooltipProps: ITooltipProps = useMemo(() => {
    return {
      // eslint-disable-next-line react/no-unstable-nested-components
      onRenderContent: () => {
        const tooltipcontent = renderToString("AuditLogScreen.last-update-at", {
          datetime:
            DateTime.fromJSDate(props.lastUpdatedAt).toRelative({ locale }) ??
            "",
        });
        return <>{tooltipcontent}</>;
      },
    };
  }, [locale, props.lastUpdatedAt, renderToString]);

  return (
    <TooltipHost
      styles={tooltipStyle}
      id={tooltipId}
      calloutProps={tooltipCalloutProps}
      directionalHint={DirectionalHint.bottomCenter}
      tooltipProps={tooltipProps}
    >
      {/* @ts-expect-error */}
      <CommandBarButton {...props} />
    </TooltipHost>
  );
}

const AuditLogScreen: React.FC = function AuditLogScreen() {
  const [offset, setOffset] = useState(0);
  const [selectedKey, setSelectedKey] = useState("ALL");
  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);
  const [sortDirection, setSortDirection] = useState<SortDirection>(
    SortDirection.Desc
  );
  const [lastUpdatedAt, setLastUpdatedAt] = useState(new Date());

  const {
    committedValue: rangeFrom,
    uncommittedValue: uncommittedRangeFrom,
    setValue: setRangeFrom,
    setCommittedValue: setRangeFromImmediately,
    commit: commitRangeFrom,
    rollback: rollbackRangeFrom,
  } = useTransactionalState<Date | null>(null);

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTransactionalState<Date | null>(null);

  const { appID } = useParams();
  const featureConfig = useAppFeatureConfigQuery(appID);

  const logRetrievalDays = useMemo(() => {
    if (featureConfig.loading) {
      return -1;
    }
    return (
      featureConfig.effectiveFeatureConfig?.audit_log?.retrieval_days ?? -1
    );
  }, [
    featureConfig.loading,
    featureConfig.effectiveFeatureConfig?.audit_log?.retrieval_days,
  ]);

  const datePickerMinDate = useMemo(() => {
    if (logRetrievalDays === -1) {
      return undefined;
    }
    const minDate = addDays(lastUpdatedAt, -logRetrievalDays + 1);
    minDate.setHours(0, 0, 0, 0);
    return minDate;
  }, [lastUpdatedAt, logRetrievalDays]);

  const queryRangeFrom = useMemo(() => {
    if (rangeFrom != null) {
      return rangeFrom.toISOString();
    }
    if (datePickerMinDate != null) {
      return datePickerMinDate.toISOString();
    }
    return null;
  }, [rangeFrom, datePickerMinDate]);

  const queryRangeTo = useMemo(() => {
    if (rangeTo != null) {
      return DateTime.fromJSDate(rangeTo)
        .plus({ days: 1 })
        .toJSDate()
        .toISOString();
    }
    return lastUpdatedAt;
  }, [rangeTo, lastUpdatedAt]);

  const isCustomDateRange = rangeFrom != null || rangeTo != null;

  const { renderToString } = useContext(Context);

  const activityTypeOptions = useMemo(() => {
    const options = [
      {
        key: "ALL",
        text: renderToString("AuditLogActivityType.ALL"),
      },
    ];
    for (const key of Object.values(AuditLogActivityType)) {
      options.push({
        key: key,
        text: renderToString("AuditLogActivityType." + key),
      });
    }
    return options;
  }, [renderToString]);

  const activityTypes: AuditLogActivityType[] | null = useMemo(() => {
    if (selectedKey === "ALL") {
      return null;
    }
    return [selectedKey] as AuditLogActivityType[];
  }, [selectedKey]);

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="AuditLogScreen.title" /> }];
  }, []);

  const cursor = useMemo(() => {
    if (offset === 0) {
      return null;
    }
    return encodeOffsetToCursor(offset - 1);
  }, [offset]);

  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const {
    data: currentData,
    previousData,
    error,
    loading,
    refetch,
  } = useQuery<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>(
    AuditLogListQueryDocument,
    {
      variables: {
        pageSize,
        cursor,
        activityTypes,
        rangeFrom: queryRangeFrom,
        rangeTo: queryRangeTo,
        sortDirection,
      },
      fetchPolicy: "network-only",
      skip: featureConfig.loading,
    }
  );

  const data = currentData ?? previousData;

  const messageBar = useMemo(() => {
    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }
    if (featureConfig.error != null) {
      return (
        <ShowError
          error={featureConfig.error}
          onRetry={() => {
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }
    return null;
  }, [error, refetch, featureConfig]);

  const onChangeSelectedKey = useCallback(
    (_e: React.FormEvent<HTMLDivElement>, item?: IDropdownOption) => {
      if (item != null && typeof item.key === "string") {
        setOffset(0);
        setSelectedKey(item.key);
      }
    },
    []
  );

  const dropdownProps: CommandBarDropdownProps = useMemo(() => {
    return {
      selectedKey,
      placeholder: "",
      label: "",
      options: activityTypeOptions,
      iconProps: {
        iconName: "PC1",
      },
      onChange: onChangeSelectedKey,
      calloutProps: { directionalHintFixed: true },
    };
  }, [selectedKey, onChangeSelectedKey, activityTypeOptions]);

  const onClickAllDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setRangeFromImmediately(null);
      setRangeToImmediately(null);
    },
    [setRangeFromImmediately, setRangeToImmediately]
  );

  const onClickCustomDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(false);
    },
    []
  );

  const onClickRefresh = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setLastUpdatedAt(new Date());
      setOffset(0);
    },
    [setLastUpdatedAt, setOffset]
  );

  const commandBarFarItems: ICommandBarItemProps[] = useMemo(() => {
    const allDateRangeLabel = renderToString("AuditLogScreen.date-range.all");
    const customDateRangeLabel = renderToString(
      "AuditLogScreen.date-range.custom"
    );
    return [
      {
        key: "dateRange",
        text: isCustomDateRange ? customDateRangeLabel : allDateRangeLabel,
        iconProps: { iconName: "Calendar" },
        subMenuProps: {
          items: [
            {
              key: "allDateRange",
              text: allDateRangeLabel,
              onClick: onClickAllDateRange,
            },
            {
              key: "customDateRange",
              text: customDateRangeLabel,
              onClick: onClickCustomDateRange,
            },
          ],
        },
      },
      {
        key: "activityTypes",
        commandBarButtonAs: CommandBarDropdownWrapper,
        dropdownProps,
      },
    ];
  }, [
    dropdownProps,
    renderToString,
    isCustomDateRange,
    onClickAllDateRange,
    onClickCustomDateRange,
  ]);

  const commandBarSecondaryItems: ICommandBarItemProps[] = useMemo(() => {
    const refreshLabel = renderToString("AuditLogScreen.refresh");
    return [
      {
        key: "refresh",
        text: refreshLabel,
        iconProps: { iconName: "Sync" },
        onClick: onClickRefresh,
        commandBarButtonAs: RefreshButton,
        lastUpdatedAt,
      },
    ];
  }, [onClickRefresh, renderToString, lastUpdatedAt]);

  const onDismissDateRangeDialog = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      rollbackRangeFrom();
      rollbackRangeTo();
    },
    [rollbackRangeFrom, rollbackRangeTo]
  );

  const commitDateRange = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.preventDefault();
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      commitRangeFrom();
      commitRangeTo();
      setOffset(0);
    },
    [commitRangeFrom, commitRangeTo]
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

  const onToggleSortDirection = useCallback(() => {
    if (sortDirection === SortDirection.Desc) {
      setSortDirection(SortDirection.Asc);
    } else {
      setSortDirection(SortDirection.Desc);
    }
  }, [sortDirection]);

  return (
    <>
      <CommandBarContainer
        isLoading={loading}
        messageBar={messageBar}
        primaryItems={commandBarFarItems}
        secondaryItems={commandBarSecondaryItems}
        className={styles.root}
      >
        <ScreenContent className={styles.content} layout="list">
          <div className={styles.widget}>
            <NavBreadcrumb items={items} />
            {logRetrievalDays !== -1 && (
              <MessageBar className={styles.messageBar}>
                <FormattedMessage
                  id="FeatureConfig.audit-log.retrieval-days"
                  values={{
                    planPagePath: "../billing",
                    logRetrievalDays: logRetrievalDays,
                  }}
                />
              </MessageBar>
            )}
          </div>
          <AuditLogList
            className={styles.widget}
            loading={loading}
            auditLogs={data?.auditLogs ?? null}
            offset={offset}
            pageSize={pageSize}
            totalCount={data?.auditLogs?.totalCount ?? undefined}
            onChangeOffset={onChangeOffset}
            onToggleSortDirection={onToggleSortDirection}
            sortDirection={sortDirection}
          />
        </ScreenContent>
      </CommandBarContainer>
      <DateRangeDialog
        hidden={dateRangeDialogHidden}
        title={renderToString("AuditLogScreen.date-range.custom")}
        fromDatePickerLabel={renderToString(
          "AuditLogScreen.date-range.start-date"
        )}
        toDatePickerLabel={renderToString("AuditLogScreen.date-range.end-date")}
        rangeFrom={uncommittedRangeFrom ?? undefined}
        rangeTo={uncommittedRangeTo ?? undefined}
        fromDatePickerMinDate={datePickerMinDate}
        fromDatePickerMaxDate={lastUpdatedAt}
        toDatePickerMinDate={datePickerMinDate}
        toDatePickerMaxDate={lastUpdatedAt}
        onSelectRangeFrom={onSelectRangeFrom}
        onSelectRangeTo={onSelectRangeTo}
        onCommitDateRange={commitDateRange}
        onDismiss={onDismissDateRangeDialog}
      />
    </>
  );
};

export default AuditLogScreen;
