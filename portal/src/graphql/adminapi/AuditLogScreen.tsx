import React, {
  useState,
  useMemo,
  useCallback,
  useContext,
  useEffect,
  useRef,
} from "react";
import {
  useParams,
  useSearchParams,
  URLSearchParamsInit,
} from "react-router-dom";
import { addDays, PivotItem, ISearchBoxProps } from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
import { FormattedMessage, Context } from "../../intl";
import { useQuery } from "@apollo/client";
import { DateTime } from "luxon";
import NavBreadcrumb from "../../NavBreadcrumb";
import AuditLogList from "./AuditLogList";
import CommandBarContainer from "../../CommandBarContainer";
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
import {
  ADMIN_ACTIVITY_TYPES,
  AuditLogKind,
  isAuditLogKind,
  USER_ACTIVITY_TYPES,
} from "./auditLogActivityTypes";
import styles from "./AuditLogScreen.module.css";
import { useAppFeatureConfigQuery } from "../portal/query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "../portal/FeatureDisabledMessageBar";
import { useDebounced } from "../../hook/useDebounced";
import { toTypedID } from "../../util/graphql";
import { NodeType } from "./node";
import { parseEmail } from "../../util/email";
import { parsePhoneNumber } from "../../util/phone";
import {
  AuditLogFilter,
  AuditLogFilterBar,
  AuditLogFilterBarPropsDateRange,
} from "../../components/audit-log/AuditLogFilterBar";
import {
  ACTIVITY_TYPE_ALL,
  ActivityTypeFilterDropdownOptionKey,
} from "../../components/audit-log/ActivityTypeFilterDropdown";
import { formatCustomDateRangeLabel } from "../../util/formatDatetime";

const pageSize = 100;

function parseDateRangeSearchParam(
  value: string | null,
  bound: "from" | "to"
): Date | null {
  if (value == null || value === "") {
    return null;
  }
  // Legacy bookmarks used yyyy-MM-dd. Backend rangeTo is exclusive
  // (created_at < rangeTo), so `from` is start of day and `to` is end of day
  // to keep the selected calendar day inclusive.
  if (/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    const dt = DateTime.fromISO(value);
    if (!dt.isValid) {
      return null;
    }
    return bound === "to"
      ? dt.endOf("day").toJSDate()
      : dt.startOf("day").toJSDate();
  }
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function formatDateRangeSearchParam(date: Date | null): string {
  if (date == null) {
    return "";
  }
  return DateTime.fromJSDate(date).toISO() ?? "";
}

function isBareAuditLogListURL(
  queryAuditLogKind: string,
  queryString: string,
  queryLastUpdatedAt: string | null,
  queryPage: string | null
): boolean {
  return (
    queryAuditLogKind === "" &&
    queryString === "" &&
    (queryLastUpdatedAt == null || queryLastUpdatedAt === "") &&
    (queryPage == null || queryPage === "")
  );
}

const AuditLogScreen: React.VFC = function AuditLogScreen() {
  const [searchParams, setSearchParams] = useSearchParams();

  const queryFrom = searchParams.get("from");
  const queryTo = searchParams.get("to");
  const queryOrderBy =
    searchParams.get("order_by") === SortDirection.Asc
      ? SortDirection.Asc
      : SortDirection.Desc;
  const queryPage = searchParams.get("page");
  const queryActivityType = searchParams.get("activity_type");
  const queryLastUpdatedAt = searchParams.get("last_updated_at");
  const queryAuditLogKind = searchParams.get("kind") ?? "";
  const queryString = searchParams.get("q") ?? "";

  const initialOffset = useMemo(() => {
    if (queryPage != null) {
      const page = parseInt(queryPage, 10);
      if (page >= 1) {
        return (page - 1) * pageSize;
      }
    }
    return 0;
  }, [queryPage]);

  const [offset, setOffset] = useState(initialOffset);
  const [sortDirection, setSortDirection] =
    useState<SortDirection>(queryOrderBy);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(
    queryLastUpdatedAt != null
      ? new Date(Number(queryLastUpdatedAt))
      : new Date()
  );
  const lastUpdatedAtRef = useRef(lastUpdatedAt);
  useEffect(() => {
    lastUpdatedAtRef.current = lastUpdatedAt;
  });
  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);
  const auditLogKind: AuditLogKind = isAuditLogKind(queryAuditLogKind)
    ? queryAuditLogKind
    : AuditLogKind.User;

  const availableActivityTypes = useMemo(() => {
    return auditLogKind === "admin"
      ? ADMIN_ACTIVITY_TYPES
      : USER_ACTIVITY_TYPES;
  }, [auditLogKind]);

  const defaultActivityType =
    useMemo<ActivityTypeFilterDropdownOptionKey>(() => {
      if (queryActivityType == null) {
        return ACTIVITY_TYPE_ALL;
      }
      const queryActivityTypeKey = queryActivityType as AuditLogActivityType;
      if (availableActivityTypes.includes(queryActivityTypeKey)) {
        return queryActivityTypeKey;
      }
      return ACTIVITY_TYPE_ALL;
    }, [availableActivityTypes, queryActivityType]);

  const [filters, setFilters] = useState<AuditLogFilter>({
    searchKeyword: queryString,
    activityType: defaultActivityType,
  });

  const {
    committedValue: rangeFrom,
    uncommittedValue: uncommittedRangeFrom,
    setValue: setRangeFrom,
    setCommittedValue: setRangeFromImmediately,
    commit: commitRangeFrom,
    rollback: rollbackRangeFrom,
  } = useTransactionalState<Date | null>(
    parseDateRangeSearchParam(queryFrom, "from")
  );

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTransactionalState<Date | null>(
    parseDateRangeSearchParam(queryTo, "to")
  );

  const { appID } = useParams() as { appID: string };
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
      return rangeTo.toISOString();
    }
    return lastUpdatedAt.toISOString();
  }, [rangeTo, lastUpdatedAt]);

  const isCustomDateRange = rangeFrom != null || rangeTo != null;

  const { renderToString } = useContext(Context);

  const customDateRangeLabel = useMemo(
    () =>
      isCustomDateRange
        ? formatCustomDateRangeLabel(rangeFrom, rangeTo)
        : undefined,
    [isCustomDateRange, rangeFrom, rangeTo]
  );

  const onClickAllDateRange = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setRangeFromImmediately(null);
      setRangeToImmediately(null);
    },
    [setRangeFromImmediately, setRangeToImmediately]
  );

  const onClickCustomDateRange = useCallback(
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(false);
    },
    []
  );

  const filtersDateRange = useMemo<AuditLogFilterBarPropsDateRange>(() => {
    return {
      value: isCustomDateRange ? "customDateRange" : "allDateRange",
      customRangeLabel: customDateRangeLabel,
      onClickAllDateRange,
      onClickCustomDateRange,
    };
  }, [
    isCustomDateRange,
    customDateRangeLabel,
    onClickAllDateRange,
    onClickCustomDateRange,
  ]);

  const [debouncedSearchQuery] = useDebounced(filters.searchKeyword, 300);

  // Keep local state in sync when the URL changes (e.g. browser back/forward).
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setFilters((prev) => {
      const next = {
        searchKeyword: queryString,
        activityType: defaultActivityType,
      };
      if (
        prev.searchKeyword === next.searchKeyword &&
        prev.activityType === next.activityType
      ) {
        return prev;
      }
      return next;
    });
  }, [queryString, defaultActivityType]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setOffset(initialOffset);
  }, [initialOffset]);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setSortDirection(queryOrderBy);
  }, [queryOrderBy]);

  useEffect(() => {
    if (queryLastUpdatedAt == null || queryLastUpdatedAt === "") {
      return;
    }
    const next = new Date(Number(queryLastUpdatedAt));
    if (next.getTime() === lastUpdatedAtRef.current.getTime()) {
      return;
    }
    setLastUpdatedAt(next);
  }, [queryLastUpdatedAt]);

  // Reset page to zero on search
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setOffset(0);
  }, [debouncedSearchQuery]);

  // On first page load without a timestamp in the URL (e.g. sidebar nav),
  // set last_updated_at. When arriving via a link that already includes
  // last_updated_at (e.g. User Details "View all"), keep the URL value so we
  // do not rewrite search params and break the browser back button.
  useEffect(() => {
    if (queryPage !== "1") {
      return;
    }
    if (queryLastUpdatedAt != null && queryLastUpdatedAt !== "") {
      return;
    }
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLastUpdatedAt(new Date());
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Sync state to searchParams.

  useEffect(() => {
    const page = offset / pageSize + 1;

    const params: URLSearchParamsInit = {};

    const newQueryFrom = formatDateRangeSearchParam(rangeFrom);
    const newQueryTo = formatDateRangeSearchParam(rangeTo);
    const newQueryOrderBy = sortDirection;
    const newQueryPage = page.toString();
    const newQueryActivityType = filters.activityType;
    const newQueryLastUpdatedAt = lastUpdatedAt.getTime().toString();
    const newAuditLogKind = auditLogKind;
    const newQueryString = debouncedSearchQuery;

    params["from"] = newQueryFrom;
    params["to"] = newQueryTo;
    params["order_by"] = newQueryOrderBy;
    params["page"] = newQueryPage;
    params["activity_type"] = newQueryActivityType;
    params["last_updated_at"] = newQueryLastUpdatedAt;
    params["kind"] = newAuditLogKind;
    params["q"] = newQueryString;

    let callSet = false;
    if (newQueryFrom !== queryFrom) {
      callSet = true;
    }
    if (newQueryTo !== queryTo) {
      callSet = true;
    }
    if (newQueryOrderBy !== queryOrderBy) {
      callSet = true;
    }
    if (newQueryPage !== queryPage) {
      callSet = true;
    }
    if (newQueryActivityType !== queryActivityType) {
      callSet = true;
    }
    if (newQueryLastUpdatedAt !== queryLastUpdatedAt) {
      callSet = true;
    }
    if (newAuditLogKind !== queryAuditLogKind) {
      callSet = true;
    }
    if (newQueryString !== queryString) {
      callSet = true;
    }

    if (callSet) {
      const replace = isBareAuditLogListURL(
        queryAuditLogKind,
        queryString,
        queryLastUpdatedAt,
        queryPage
      );
      setSearchParams(params, { replace });
    }
  }, [
    queryFrom,
    queryTo,
    queryOrderBy,
    queryPage,
    queryActivityType,
    queryLastUpdatedAt,
    rangeFrom,
    rangeTo,
    sortDirection,
    offset,
    filters.activityType,
    lastUpdatedAt,
    setSearchParams,
    auditLogKind,
    queryAuditLogKind,
    debouncedSearchQuery,
    queryString,
  ]);

  const activityTypes: AuditLogActivityType[] | null = useMemo(() => {
    if (filters.activityType === ACTIVITY_TYPE_ALL) {
      return availableActivityTypes;
    }
    return [filters.activityType];
  }, [availableActivityTypes, filters.activityType]);

  const items = useMemo(() => {
    return [{ to: ".", label: <FormattedMessage id="AuditLogScreen.title" /> }];
  }, []);

  const cursor = useMemo(() => {
    return encodeOffsetToCursor(offset);
  }, [offset]);

  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const onChangeOffset = useCallback((offset) => {
    setOffset(offset);
  }, []);

  const queryEmailAddresses = useMemo(() => {
    const email = parseEmail(debouncedSearchQuery);
    if (email == null) {
      return null;
    }

    switch (filters.activityType) {
      // Only search email addresses if `all` or `email_sent` filter active
      case ACTIVITY_TYPE_ALL:
      case AuditLogActivityType.EmailSent:
        return email ? [email] : null;
      default:
        return null;
    }
  }, [debouncedSearchQuery, filters.activityType]);

  const queryPhoneNumbers = useMemo(() => {
    const phoneNumber = parsePhoneNumber(debouncedSearchQuery);
    if (phoneNumber == null) {
      return null;
    }
    switch (filters.activityType) {
      // Only search phone numbers if `all` or `phone_sent` or `whatsapp_sent` filter active
      case ACTIVITY_TYPE_ALL:
      case AuditLogActivityType.SmsSent:
      case AuditLogActivityType.WhatsappSent:
        return [phoneNumber];
      default:
        return null;
    }
  }, [debouncedSearchQuery, filters.activityType]);

  const queryUserIDs = useMemo(() => {
    if (queryEmailAddresses != null || queryPhoneNumbers != null) {
      return null;
    }
    // only search by userIDs if query notLikeEmail & notLikePhoneNumber
    return debouncedSearchQuery
      ? [toTypedID(NodeType.User, debouncedSearchQuery)]
      : null;
  }, [debouncedSearchQuery, queryEmailAddresses, queryPhoneNumbers]);

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
        userIDs: queryUserIDs,
        emailAddresses: queryEmailAddresses,
        phoneNumbers: queryPhoneNumbers,
        rangeFrom: queryRangeFrom,
        rangeTo: queryRangeTo,
        sortDirection,
      },
      fetchPolicy: "network-only",
      skip: featureConfig.isLoading,
    }
  );

  const data = currentData ?? previousData;

  const messageBar = useMemo(() => {
    if (error != null) {
      // eslint-disable-next-line @typescript-eslint/strict-void-return
      return <ShowError error={error} onRetry={refetch} />;
    }
    if (featureConfig.loadError != null) {
      return (
        <ShowError
          error={featureConfig.loadError}
          onRetry={() => {
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }
    return null;
  }, [error, refetch, featureConfig]);

  const onFilterChange = useCallback(
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
    (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => {
      const newFilters = fn(filters);

      if (newFilters.activityType !== filters.activityType) {
        setOffset(0);
      }
      setFilters(fn);
    },
    [filters]
  );

  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const onRemoveAllFilters = useCallback(() => {
    setOffset(0);
    setRangeFromImmediately(null);
    setRangeToImmediately(null);
    onFilterChange(() => ({
      searchKeyword: "",
      activityType: ACTIVITY_TYPE_ALL,
    }));
  }, [onFilterChange, setRangeFromImmediately, setRangeToImmediately]);

  const onClickRefresh = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setLastUpdatedAt(new Date());
      setOffset(0);
    },
    [setLastUpdatedAt, setOffset]
  );

  const searchBoxPlaceholder = useMemo(() => {
    switch (filters.activityType) {
      case AuditLogActivityType.EmailSent:
        return renderToString("AuditLogScreen.search-by-user-id-or-email");
      case AuditLogActivityType.SmsSent:
      case AuditLogActivityType.WhatsappSent:
        return renderToString("AuditLogScreen.search-by-user-id-or-phone");
      default:
        return renderToString("AuditLogScreen.search-by-user-id");
    }
  }, [filters.activityType, renderToString]);

  const searchBoxProps = useMemo<ISearchBoxProps>(() => {
    return {
      placeholder: searchBoxPlaceholder,
    };
  }, [searchBoxPlaceholder]);

  const onDismissDateRangeDialog = useCallback(
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
    (e?: React.MouseEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      rollbackRangeFrom();
      rollbackRangeTo();
    },
    [rollbackRangeFrom, rollbackRangeTo]
  );

  const commitDateRange = useCallback(
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
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

  // eslint-disable-next-line react-hooks/preserve-manual-memoization
  const onToggleSortDirection = useCallback(() => {
    if (sortDirection === SortDirection.Desc) {
      setSortDirection(SortDirection.Asc);
    } else {
      setSortDirection(SortDirection.Desc);
    }
  }, [sortDirection]);

  const onTabChange = useCallback(
    // eslint-disable-next-line react-hooks/preserve-manual-memoization
    (item?: PivotItem) => {
      if (item == null) {
        return;
      }
      const { itemKey } = item.props;
      if (!itemKey || !isAuditLogKind(itemKey)) {
        return;
      }
      setOffset(0);
      setFilters((prev) => ({
        ...prev,
        activityType: ACTIVITY_TYPE_ALL,
      }));
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev);
          params.set("kind", itemKey);
          params.set("page", "1");
          params.set("activity_type", ACTIVITY_TYPE_ALL);
          return params;
        },
        { replace: false }
      );
    },
    [setSearchParams]
  );
  return (
    <>
      <div className={styles.root}>
        <div className={styles.header}>
          <NavBreadcrumb className="" items={items} />
          {logRetrievalDays !== -1 ? (
            <FeatureDisabledMessageBar
              className={styles.messageBar}
              messageID="FeatureConfig.audit-log.retrieval-days"
              messageValues={{ logRetrievalDays: logRetrievalDays }}
            />
          ) : null}
          <AGPivot
            className={styles.pivot}
            selectedKey={auditLogKind}
            onLinkClick={onTabChange}
          >
            <PivotItem
              itemKey={AuditLogKind.User}
              headerText={renderToString("AuditLogScreen.acitity-kind.user")}
            />
            <PivotItem
              itemKey={AuditLogKind.Admin}
              headerText={renderToString("AuditLogScreen.acitity-kind.admin")}
            />
          </AGPivot>
        </div>
        <AuditLogFilterBar
          filters={filters}
          onFilterChange={onFilterChange}
          searchBoxProps={searchBoxProps}
          dateRange={filtersDateRange}
          availableActivityTypes={availableActivityTypes}
          onRemoveAllFilters={onRemoveAllFilters}
          onRefresh={onClickRefresh}
          lastUpdatedAt={lastUpdatedAt}
        />
        <div className={styles.listContainer}>
          <CommandBarContainer
            isLoading={loading}
            messageBar={messageBar}
            hideCommandBar={true}
            className={styles.commandBarContainerContent}
            headerPosition="static"
          >
            <AuditLogList
              className={styles.list}
              loading={loading}
              auditLogs={data?.auditLogs ?? null}
              searchParams={searchParams.toString()}
              offset={offset}
              pageSize={pageSize}
              totalCount={data?.auditLogs?.totalCount ?? undefined}
              onChangeOffset={onChangeOffset}
              onToggleSortDirection={onToggleSortDirection}
              sortDirection={sortDirection}
            />
          </CommandBarContainer>
        </div>
      </div>
      <DateRangeDialog
        hidden={dateRangeDialogHidden}
        title={renderToString("AuditLogScreen.date-range.custom")}
        fromDatePickerLabel={renderToString(
          "AuditLogScreen.date-range.start-date"
        )}
        toDatePickerLabel={renderToString("AuditLogScreen.date-range.end-date")}
        rangeFrom={uncommittedRangeFrom ?? undefined}
        rangeTo={uncommittedRangeTo ?? undefined}
        fromDatePickerMaxDate={lastUpdatedAt}
        toDatePickerMaxDate={lastUpdatedAt}
        onSelectRangeFrom={onSelectRangeFrom}
        onSelectRangeTo={onSelectRangeTo}
        onCommitDateRange={commitDateRange}
        onDismiss={onDismissDateRangeDialog}
        showTimePicker={true}
      />
    </>
  );
};

export default AuditLogScreen;
