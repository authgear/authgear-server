import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { addDays } from "@fluentui/react";
import {
  Callout as RadixCallout,
  Spinner,
  Table,
  Text,
} from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { DateTime } from "luxon";
import { Context, FormattedMessage } from "../../intl";
import Link from "../../Link";
import PaginationWidget from "../../PaginationWidget";
import ShowError from "../../ShowError";
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
import {
  AuditLogFilter,
  AuditLogFilterBar,
  AuditLogFilterBarPropsDateRange,
} from "../../components/audit-log/AuditLogFilterBar";
import AuditLogDateRangeDialog from "../../components/audit-log/AuditLogDateRangeDialog";
import {
  AuditLogDateRangePresetKey,
  getPresetDateRange,
} from "../../components/audit-log/dateRangePresets";
import { serializeActivityTypesToQuery } from "../../components/audit-log/ActivityTypeFilterDropdown";
import { AuditLogKind, USER_ACTIVITY_TYPES } from "./auditLogActivityTypes";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
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
    activity_type: serializeActivityTypesToQuery([]),
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
    const [dateRangePreset, setDateRangePreset] =
      useState<AuditLogDateRangePresetKey>("last30Days");
    const presetBeforeCustomDialogRef =
      useRef<AuditLogDateRangePresetKey>("last30Days");
    const [filters, setFilters] = useState<AuditLogFilter>({
      searchKeyword: "",
      activityTypes: [],
    });

    const initialDateRange = useMemo(
      () => getPresetDateRange("last30Days", lastUpdatedAt),
      // Only compute the first default range.
      // eslint-disable-next-line react-hooks/exhaustive-deps
      []
    );

    const {
      committedValue: rangeFrom,
      uncommittedValue: uncommittedRangeFrom,
      setValue: setRangeFrom,
      setCommittedValue: setRangeFromImmediately,
      commit: commitRangeFrom,
      rollback: rollbackRangeFrom,
    } = useTransactionalState<Date | null>(initialDateRange.from);

    const {
      committedValue: rangeTo,
      uncommittedValue: uncommittedRangeTo,
      setValue: setRangeTo,
      setCommittedValue: setRangeToImmediately,
      commit: commitRangeTo,
      rollback: rollbackRangeTo,
    } = useTransactionalState<Date | null>(initialDateRange.to);

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

    useEffect(() => {
      if (dateRangePreset === "custom") {
        return;
      }
      const range = getPresetDateRange(
        dateRangePreset,
        lastUpdatedAt,
        datePickerMinDate
      );
      setRangeFromImmediately(range.from);
      setRangeToImmediately(range.to);
    }, [
      dateRangePreset,
      lastUpdatedAt,
      datePickerMinDate,
      setRangeFromImmediately,
      setRangeToImmediately,
    ]);

    const activityTypes: AuditLogActivityType[] = useMemo(() => {
      if (filters.activityTypes.length === 0) {
        return USER_ACTIVITY_TYPES;
      }
      return filters.activityTypes;
    }, [filters.activityTypes]);

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

    const onOpenCustomDateRangeDialog = useCallback(() => {
      presetBeforeCustomDialogRef.current = dateRangePreset;
      setDateRangeDialogHidden(false);
    }, [dateRangePreset]);

    const onChangeDateRangePreset = useCallback(
      (preset: AuditLogDateRangePresetKey) => {
        if (preset === "custom") {
          presetBeforeCustomDialogRef.current = dateRangePreset;
          setDateRangePreset(preset);
          setDateRangeDialogHidden(false);
          return;
        }
        setDateRangePreset(preset);
        setOffset(0);
      },
      [dateRangePreset]
    );

    const filtersDateRange = useMemo<AuditLogFilterBarPropsDateRange>(() => {
      return {
        value: dateRangePreset,
        onChange: onChangeDateRangePreset,
        rangeFrom,
        rangeTo,
        onOpenCustomDateRangeDialog,
      };
    }, [
      dateRangePreset,
      onChangeDateRangePreset,
      rangeFrom,
      rangeTo,
      onOpenCustomDateRangeDialog,
    ]);

    const onFilterChange = useCallback(
      (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => {
        setFilters((prev) => {
          const next = fn(prev);
          if (
            next.activityTypes.length !== prev.activityTypes.length ||
            next.activityTypes.some(
              (activityType, index) =>
                activityType !== prev.activityTypes[index]
            )
          ) {
            setOffset(0);
          }
          return next;
        });
      },
      []
    );

    const onClickRefresh = useCallback(() => {
      setLastUpdatedAt(new Date());
      setOffset(0);
    }, []);

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
        setDateRangePreset(presetBeforeCustomDialogRef.current);
      },
      [rollbackRangeFrom, rollbackRangeTo]
    );

    const onClickViewUserLogs = useCallback(() => {
      navigate(buildAuditLogListHref(appID, AuditLogKind.User, rawUserID));
    }, [appID, navigate, rawUserID]);

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

    const totalCount = data?.auditLogs?.totalCount ?? undefined;
    const isEmpty = !loading && items.length === 0;

    const onChangeOffset = useCallback((nextOffset: number) => {
      setOffset(nextOffset);
    }, []);

    return (
      <div className={styles.root}>
        <AuditLogFilterBar
          className={styles.filterBar}
          filters={filters}
          onFilterChange={onFilterChange}
          onRefresh={onClickRefresh}
          hideSearchBox={true}
          dateRange={filtersDateRange}
          availableActivityTypes={USER_ACTIVITY_TYPES}
          lastUpdatedAt={lastUpdatedAt}
          trailingActions={
            <SecondaryButton
              size="2"
              text={<FormattedMessage id="UserDetails.logs.view-user-logs" />}
              onClick={onClickViewUserLogs}
            />
          }
        />
        {error != null ? (
          // eslint-disable-next-line @typescript-eslint/strict-void-return
          <ShowError error={error} onRetry={refetch} />
        ) : (
          <div className={styles.tableArea}>
            {loading ? (
              <div className={styles.loading}>
                <Spinner />
              </div>
            ) : isEmpty ? (
              <RadixCallout.Root color="gray" size="2" variant="surface">
                <RadixCallout.Icon>
                  <InfoCircledIcon width="1rem" height="1rem" />
                </RadixCallout.Icon>
                <RadixCallout.Text>
                  <FormattedMessage id="UserDetails.logs.empty" />
                </RadixCallout.Text>
              </RadixCallout.Root>
            ) : (
              <>
                <div className={styles.listWrapper} data-is-scrollable="true">
                  <Table.Root className={styles.table} variant="surface">
                    <Table.Header>
                      <Table.Row>
                        <Table.ColumnHeaderCell
                          className={styles.activityColumn}
                        >
                          <FormattedMessage id="UserDetails.logs.column.activity" />
                        </Table.ColumnHeaderCell>
                        <Table.ColumnHeaderCell
                          className={styles.timestampColumn}
                        >
                          <FormattedMessage id="UserDetails.logs.column.timestamp" />
                        </Table.ColumnHeaderCell>
                      </Table.Row>
                    </Table.Header>
                    <Table.Body>
                      {items.map((item) => (
                        <Table.Row key={item.id}>
                          <Table.Cell className={styles.activityColumn}>
                            <Link
                              to={`/project/${appID}/audit-log/${item.id}/details`}
                            >
                              {item.activityType}
                            </Link>
                          </Table.Cell>
                          <Table.Cell className={styles.timestampColumn}>
                            <Text as="span" size="2" color="gray">
                              {item.createdAt}
                            </Text>
                          </Table.Cell>
                        </Table.Row>
                      ))}
                    </Table.Body>
                  </Table.Root>
                </div>
                <PaginationWidget
                  className={styles.pagination}
                  offset={offset}
                  pageSize={LOG_PAGE_SIZE}
                  totalCount={totalCount}
                  onChangeOffset={onChangeOffset}
                />
              </>
            )}
          </div>
        )}
        <AuditLogDateRangeDialog
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
