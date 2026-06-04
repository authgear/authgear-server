import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  addDays,
  IColumn,
  MessageBar,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import { DateTime } from "luxon";
import { Context, FormattedMessage } from "../../intl";
import Link from "../../Link";
import CommandBarButton from "../../CommandBarButton";
import PaginationWidget from "../../PaginationWidget";
import ShowError from "../../ShowError";
import DateRangeDialog from "../portal/DateRangeDialog";
import { formatDatetime } from "../../util/formatDatetime";
import { encodeOffsetToCursor } from "../../util/pagination";
import { extractRawID } from "../../util/graphql";
import useTransactionalState from "../../hook/useTransactionalState";
import { useAppFeatureConfigQuery } from "../portal/query/appFeatureConfigQuery";
import {
  useAuditLogListQueryQuery,
  AuditLogEdgesNodeFragment,
} from "./query/auditLogListQuery.generated";
import { AuditLogActivityType, SortDirection } from "./globalTypes.generated";
import { ACTIVITY_TYPE_ALL } from "../../components/audit-log/ActivityTypeFilterDropdown";
import {
  AuditLogFilter,
  AuditLogFilterBar,
  AuditLogFilterBarPropsDateRange,
} from "../../components/audit-log/AuditLogFilterBar";
import { AuditLogKind, USER_ACTIVITY_TYPES } from "./auditLogActivityTypes";
import styles from "./UserDetailsLogs.module.css";

const LOG_PAGE_SIZE = 20;

interface UserDetailsLogsProps {
  userID: string;
}

interface LogTableItem {
  id: string;
  activityType: string;
  createdAt: string;
}

function buildAuditLogListHref(
  appID: string,
  kind: AuditLogKind,
  rawUserID: string
): string {
  const searchParams = new URLSearchParams({
    kind,
    q: rawUserID,
    page: "1",
    order_by: SortDirection.Desc,
    activity_type: ACTIVITY_TYPE_ALL,
    last_updated_at: Date.now().toString(),
    from: "",
    to: "",
  }).toString();
  return `/project/${appID}/audit-log?${searchParams}`;
}

const UserDetailsLogs: React.VFC<UserDetailsLogsProps> =
  function UserDetailsLogs(props) {
    const { userID } = props;
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();
    const { renderToString, locale } = useContext(Context);

    const rawUserID = useMemo(() => extractRawID(userID), [userID]);

    const [offset, setOffset] = useState(0);
    const [lastUpdatedAt, setLastUpdatedAt] = useState(() => new Date());
    const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);
    const [filters, setFilters] = useState<AuditLogFilter>({
      searchKeyword: "",
      activityType: ACTIVITY_TYPE_ALL,
    });

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

    const featureConfig = useAppFeatureConfigQuery(appID);

    const logRetrievalDays = useMemo(() => {
      if (featureConfig.isLoading) {
        return -1;
      }
      return (
        featureConfig.effectiveFeatureConfig?.audit_log?.retrieval_days ?? -1
      );
    }, [
      featureConfig.isLoading,
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
      return lastUpdatedAt.toISOString();
    }, [rangeTo, lastUpdatedAt]);

    const activityTypes: AuditLogActivityType[] = useMemo(() => {
      if (filters.activityType === ACTIVITY_TYPE_ALL) {
        return USER_ACTIVITY_TYPES;
      }
      return [filters.activityType];
    }, [filters.activityType]);

    const cursor = useMemo(() => encodeOffsetToCursor(offset), [offset]);

    const { data, error, loading, refetch } = useAuditLogListQueryQuery({
      variables: {
        pageSize: LOG_PAGE_SIZE,
        cursor,
        activityTypes,
        userIDs: [userID],
        rangeFrom: queryRangeFrom,
        rangeTo: queryRangeTo,
        sortDirection: SortDirection.Desc,
      },
      fetchPolicy: "network-only",
      skip: featureConfig.isLoading,
    });

    const isCustomDateRange = rangeFrom != null || rangeTo != null;

    const onClickAllDateRange = useCallback(
      (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
        e?.stopPropagation();
        setRangeFromImmediately(null);
        setRangeToImmediately(null);
        setOffset(0);
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

    const filtersDateRange = useMemo<AuditLogFilterBarPropsDateRange>(() => {
      return {
        value: isCustomDateRange ? "customDateRange" : "allDateRange",
        onClickAllDateRange,
        onClickCustomDateRange,
      };
    }, [isCustomDateRange, onClickAllDateRange, onClickCustomDateRange]);

    const onFilterChange = useCallback(
      (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => {
        const newFilters = fn(filters);
        if (newFilters.activityType !== filters.activityType) {
          setOffset(0);
        }
        setFilters(fn);
      },
      [filters]
    );

    const onRemoveAllFilters = useCallback(() => {
      setOffset(0);
      setRangeFromImmediately(null);
      setRangeToImmediately(null);
      setFilters({
        searchKeyword: "",
        activityType: ACTIVITY_TYPE_ALL,
      });
    }, [setRangeFromImmediately, setRangeToImmediately]);

    const onClickRefresh = useCallback(
      (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
        e?.stopPropagation();
        setLastUpdatedAt(new Date());
        setOffset(0);
      },
      []
    );

    const onSelectRangeFrom = useCallback(
      (value: Date | null | undefined) => {
        if (value == null) {
          setRangeFrom(null);
        } else if (uncommittedRangeTo != null && value > uncommittedRangeTo) {
          setRangeTo(value);
          setRangeFrom(uncommittedRangeTo);
        } else {
          setRangeFrom(value);
        }
      },
      [setRangeFrom, setRangeTo, uncommittedRangeTo]
    );

    const onSelectRangeTo = useCallback(
      (value: Date | null | undefined) => {
        if (value == null) {
          setRangeTo(null);
        } else if (
          uncommittedRangeFrom != null &&
          value < uncommittedRangeFrom
        ) {
          setRangeFrom(value);
          setRangeTo(uncommittedRangeFrom);
        } else {
          setRangeTo(value);
        }
      },
      [setRangeTo, setRangeFrom, uncommittedRangeFrom]
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

    const onDismissDateRangeDialog = useCallback(
      (e?: React.MouseEvent<unknown>) => {
        e?.stopPropagation();
        setDateRangeDialogHidden(true);
        rollbackRangeFrom();
        rollbackRangeTo();
      },
      [rollbackRangeFrom, rollbackRangeTo]
    );

    const onClickViewUserLogs = useCallback(() => {
      navigate(buildAuditLogListHref(appID, AuditLogKind.User, rawUserID));
    }, [appID, navigate, rawUserID]);

    const columns: IColumn[] = useMemo(
      () => [
        {
          key: "activityType",
          fieldName: "activityType",
          name: renderToString("UserDetails.logs.column.activity"),
          minWidth: 300,
          maxWidth: 400,
          className: styles.cell,
        },
        {
          key: "createdAt",
          fieldName: "createdAt",
          name: renderToString("UserDetails.logs.column.timestamp"),
          minWidth: 220,
          maxWidth: 220,
          className: styles.cell,
        },
      ],
      [renderToString]
    );

    const items: LogTableItem[] = useMemo(() => {
      const edges = data?.auditLogs?.edges;
      const result: LogTableItem[] = [];
      if (edges == null) {
        return result;
      }
      for (const edge of edges) {
        const node: AuditLogEdgesNodeFragment | null | undefined = edge?.node;
        if (node == null) {
          continue;
        }
        result.push({
          id: node.id,
          activityType: renderToString(
            "AuditLogActivityType." + node.activityType
          ),
          createdAt: formatDatetime(locale, node.createdAt) ?? "-",
        });
      }
      return result;
    }, [data?.auditLogs?.edges, locale, renderToString]);

    const onRenderItemColumn = useCallback(
      (item: LogTableItem, _index?: number, column?: IColumn) => {
        const text = item[column?.key as keyof LogTableItem];
        if (column?.key === "activityType") {
          return (
            <Link to={`/project/${appID}/audit-log/${item.id}/details`}>
              {text}
            </Link>
          );
        }
        return <span>{text}</span>;
      },
      [appID]
    );

    const totalCount = data?.auditLogs?.totalCount ?? undefined;
    const isEmpty = !loading && items.length === 0;

    const onChangeOffset = useCallback((offset: number) => {
      setOffset(offset);
    }, []);

    return (
      <div className={styles.root}>
        <AuditLogFilterBar
          className={styles.filterBar}
          filters={filters}
          onFilterChange={onFilterChange}
          onRemoveAllFilters={onRemoveAllFilters}
          onRefresh={onClickRefresh}
          hideSearchBox={true}
          dateRange={filtersDateRange}
          availableActivityTypes={USER_ACTIVITY_TYPES}
          lastUpdatedAt={lastUpdatedAt}
          trailingActions={
            <div className={styles.viewAllWrap}>
              <CommandBarButton
                key="viewInAuditLogs"
                iconProps={{ iconName: "ComplianceAudit" }}
                text={renderToString("UserDetails.logs.view-user-logs")}
                onClick={onClickViewUserLogs}
              />
            </div>
          }
        />
        {error != null ? (
          <ShowError error={error} onRetry={refetch} />
        ) : (
          <div className={styles.tableArea}>
            <div className={styles.listWrapper} data-is-scrollable="true">
              <ShimmeredDetailsList
                enableShimmer={loading}
                enableUpdateAnimations={false}
                selectionMode={SelectionMode.none}
                columns={columns}
                items={items}
                onRenderItemColumn={onRenderItemColumn}
              />
            </div>
            {isEmpty ? (
              <MessageBar className={styles.emptyMessageBar}>
                <FormattedMessage id="UserDetails.logs.empty" />
              </MessageBar>
            ) : (
              <PaginationWidget
                className={styles.pagination}
                offset={offset}
                pageSize={LOG_PAGE_SIZE}
                totalCount={totalCount}
                onChangeOffset={onChangeOffset}
              />
            )}
          </div>
        )}
        <DateRangeDialog
          hidden={dateRangeDialogHidden}
          title={renderToString("AuditLogScreen.date-range.custom")}
          fromDatePickerLabel={renderToString(
            "AuditLogScreen.date-range.start-date"
          )}
          toDatePickerLabel={renderToString(
            "AuditLogScreen.date-range.end-date"
          )}
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
      </div>
    );
  };

export default UserDetailsLogs;
