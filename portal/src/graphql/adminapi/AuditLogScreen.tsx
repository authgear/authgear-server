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
import { FormattedMessage, Context } from "../../intl";
import { useQuery } from "@apollo/client";
import { DateTime } from "luxon";
import { Tabs, Text } from "@radix-ui/themes";
import AuditLogList from "./AuditLogList";
import CommandBarContainer from "../../CommandBarContainer";
import ShowError from "../../ShowError";
import AuditLogDateRangeDialog from "../../components/audit-log/AuditLogDateRangeDialog";
import { encodeOffsetToCursor } from "../../util/pagination";
import { addDays } from "../../util/date";
import useTransactionalState from "../../hook/useTransactionalState";
import {
  AuditLogListQueryQuery,
  AuditLogListQueryQueryVariables,
  AuditLogListQueryDocument,
} from "./query/auditLogListQuery.generated";
import { AuditLogActivityType, SortDirection } from "./globalTypes.generated";
import { toTypedID } from "../../util/graphql";
import { NodeType } from "./node";
import styles from "./AuditLogScreen.module.css";
import { useAppFeatureConfigQuery } from "../portal/query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "../portal/FeatureDisabledMessageBar";
import { useDebounced } from "../../hook/useDebounced";
import { parseEmail } from "../../util/email";
import { parsePhoneNumber } from "../../util/phone";
import {
  AuditLogFilter,
  AuditLogFilterBar,
  AuditLogFilterBarPropsDateRange,
  AuditLogSearchBoxProps,
} from "../../components/audit-log/AuditLogFilterBar";
import {
  parseActivityTypesFromQuery,
  serializeActivityTypesToQuery,
} from "../../components/audit-log/ActivityTypeFilterDropdown";
import {
  AuditLogDateRangePresetKey,
  detectDateRangePreset,
  getInitialAuditLogDateRange,
  getPresetDateRange,
} from "../../components/audit-log/dateRangePresets";

const pageSize = 100;

const ALL_ACTIVITY_TYPES = Object.values(AuditLogActivityType);
const ADMIN_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) =>
    activityType.startsWith("ADMIN_API") || activityType.startsWith("PROJECT")
);
// Activity types to hide from the audit log (shown elsewhere in the portal)
const HIDDEN_ACTIVITY_TYPES = [
  AuditLogActivityType.FraudProtectionDecisionRecorded,
];
const USER_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) =>
    !ADMIN_ACTIVITY_TYPES.includes(activityType) &&
    !HIDDEN_ACTIVITY_TYPES.includes(activityType)
);

function areActivityTypesEqual(
  left: AuditLogActivityType[],
  right: AuditLogActivityType[]
): boolean {
  if (left.length !== right.length) {
    return false;
  }
  return left.every((activityType) => right.includes(activityType));
}

enum AuditLogKind {
  User = "user",
  Admin = "admin",
}
function isAuditLogKind(s: string): s is AuditLogKind {
  return Object.values(AuditLogKind).includes(s as AuditLogKind);
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
  // When the page is refreshed, and it is on the first page,
  // update last_updated_at.
  // Note that if the page is navigated from another page,
  // this initializer is NOT run again.
  // This is the intended behavior because we do not
  // want to change last_updated_at.
  const [lastUpdatedAt, setLastUpdatedAt] = useState(() => {
    if (queryPage === "1") {
      return new Date();
    }
    return queryLastUpdatedAt != null
      ? new Date(Number(queryLastUpdatedAt))
      : new Date();
  });
  const initialDateRange = useMemo(
    () => getInitialAuditLogDateRange(queryFrom, queryTo, queryLastUpdatedAt),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    []
  );

  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);
  const presetBeforeCustomDialogRef =
    useRef<AuditLogDateRangePresetKey>("today");
  const [dateRangePreset, setDateRangePreset] =
    useState<AuditLogDateRangePresetKey>(initialDateRange.preset);
  const [auditLogKind, setAuditLogKind] = useState<AuditLogKind>(() => {
    if (isAuditLogKind(queryAuditLogKind)) {
      return queryAuditLogKind;
    }
    return AuditLogKind.User;
  });

  const availableActivityTypes = useMemo(() => {
    return auditLogKind === "admin"
      ? ADMIN_ACTIVITY_TYPES
      : USER_ACTIVITY_TYPES;
  }, [auditLogKind]);

  const defaultActivityTypes = useMemo<AuditLogActivityType[]>(() => {
    return parseActivityTypesFromQuery(
      queryActivityType,
      availableActivityTypes
    );
  }, [availableActivityTypes, queryActivityType]);

  const [filters, setFilters] = useState<AuditLogFilter>({
    searchKeyword: queryString,
    activityTypes: defaultActivityTypes,
  });

  const {
    committedValue: rangeFrom,
    uncommittedValue: uncommittedRangeFrom,
    setValue: setRangeFrom,
    setCommittedValue: setRangeFromImmediately,
    commit: commitRangeFrom,
    rollback: rollbackRangeFrom,
  } = useTransactionalState<Date | null>(initialDateRange.rangeFrom);

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTransactionalState<Date | null>(initialDateRange.rangeTo);

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
    [dateRangePreset, setOffset]
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

  const [debouncedSearchQuery] = useDebounced(filters.searchKeyword, 300);

  // Keep local state in sync when the URL changes (e.g. browser back/forward,
  // or navigating from User Details "View in Audit Logs").
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setFilters((prev) => {
      const next = {
        searchKeyword: queryString,
        activityTypes: defaultActivityTypes,
      };
      if (
        prev.searchKeyword === next.searchKeyword &&
        prev.activityTypes.length === next.activityTypes.length &&
        prev.activityTypes.every(
          (activityType, index) => activityType === next.activityTypes[index]
        )
      ) {
        return prev;
      }
      return next;
    });
  }, [queryString, defaultActivityTypes]);

  // Reset page to zero on search.
  // This adjusts state during render instead of in an effect,
  // so the reset applies in the same render pass as the search change.
  const [prevDebouncedSearchQuery, setPrevDebouncedSearchQuery] = useState<
    string | null
  >(null);
  if (prevDebouncedSearchQuery !== debouncedSearchQuery) {
    setPrevDebouncedSearchQuery(debouncedSearchQuery);
    setOffset(0);
  }

  const { renderToString } = useContext(Context);

  // Sync state to searchParams.
  // The searchParams are a mirror of the state, so they must be replaced
  // instead of pushed. Otherwise pressing back restores an out-of-sync URL,
  // which this effect immediately pushes forward again, trapping the user
  // on this screen.

  useEffect(() => {
    const page = offset / pageSize + 1;

    const params: URLSearchParamsInit = {};

    const newQueryFrom =
      rangeFrom != null ? DateTime.fromJSDate(rangeFrom).toISODate() : "";
    const newQueryTo =
      rangeTo != null ? DateTime.fromJSDate(rangeTo).toISODate() : "";
    const newQueryOrderBy = sortDirection;
    const newQueryPage = page.toString();
    const newQueryActivityType = serializeActivityTypesToQuery(
      filters.activityTypes
    );
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
      setSearchParams(params, { replace: true });
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
    filters.activityTypes,
    lastUpdatedAt,
    setSearchParams,
    auditLogKind,
    queryAuditLogKind,
    debouncedSearchQuery,
    queryString,
  ]);

  const activityTypes: AuditLogActivityType[] | null = useMemo(() => {
    if (filters.activityTypes.length === 0) {
      return availableActivityTypes;
    }
    return filters.activityTypes;
  }, [availableActivityTypes, filters.activityTypes]);

  const cursor = useMemo(() => {
    return encodeOffsetToCursor(offset);
  }, [offset]);

  const onChangeOffset = useCallback(
    (offset) => {
      setOffset(offset);
    },
    [setOffset]
  );

  // Derive from the effective activityTypes (which accounts for the
  // user/admin tab) so that query routing and the search box placeholder
  // never diverge.
  const searchIncludesEmail = useMemo(() => {
    return activityTypes.includes(AuditLogActivityType.EmailSent);
  }, [activityTypes]);

  const searchIncludesPhone = useMemo(() => {
    return (
      activityTypes.includes(AuditLogActivityType.SmsSent) ||
      activityTypes.includes(AuditLogActivityType.WhatsappSent)
    );
  }, [activityTypes]);

  const queryEmailAddresses = useMemo(() => {
    const email = parseEmail(debouncedSearchQuery);
    if (email == null) {
      return null;
    }
    if (searchIncludesEmail) {
      return [email];
    }
    return null;
  }, [debouncedSearchQuery, searchIncludesEmail]);

  const queryPhoneNumbers = useMemo(() => {
    const phoneNumber = parsePhoneNumber(debouncedSearchQuery);
    if (phoneNumber == null) {
      return null;
    }
    if (searchIncludesPhone) {
      return [phoneNumber];
    }
    return null;
  }, [debouncedSearchQuery, searchIncludesPhone]);

  const queryUserIDs = useMemo(() => {
    if (queryEmailAddresses != null || queryPhoneNumbers != null) {
      return null;
    }
    // only search by userIDs if query notLikeEmail & notLikePhoneNumber
    const trimmed = debouncedSearchQuery.trim();
    return trimmed ? [toTypedID(NodeType.User, trimmed)] : null;
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
  const auditLogs = data?.auditLogs ?? null;
  const listLoading = loading || featureConfig.isLoading;

  const messageBar = useMemo(() => {
    if (error != null) {
      return (
        <ShowError
          error={error}
          onRetry={() => {
            refetch().finally(() => {});
          }}
        />
      );
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
    (fn: (prevValue: AuditLogFilter) => AuditLogFilter) => {
      const newFilters = fn(filters);

      if (
        !areActivityTypesEqual(newFilters.activityTypes, filters.activityTypes)
      ) {
        setOffset(0);
      }
      setFilters(fn);
    },
    [filters, setOffset]
  );

  const onClickRefresh = useCallback(
    (e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>) => {
      e?.stopPropagation();
      setLastUpdatedAt(new Date());
      setOffset(0);
    },
    [setLastUpdatedAt, setOffset]
  );

  const searchBoxPlaceholder = useMemo(() => {
    if (searchIncludesEmail && searchIncludesPhone) {
      return renderToString(
        "AuditLogScreen.search-by-user-id-or-email-or-phone"
      );
    }
    if (searchIncludesEmail) {
      return renderToString("AuditLogScreen.search-by-user-id-or-email");
    }
    if (searchIncludesPhone) {
      return renderToString("AuditLogScreen.search-by-user-id-or-phone");
    }
    return renderToString("AuditLogScreen.search-by-user-id");
  }, [searchIncludesEmail, searchIncludesPhone, renderToString]);

  const searchBoxProps = useMemo<AuditLogSearchBoxProps>(() => {
    return {
      placeholder: searchBoxPlaceholder,
    };
  }, [searchBoxPlaceholder]);

  const onDismissDateRangeDialog = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      rollbackRangeFrom();
      rollbackRangeTo();
      if (presetBeforeCustomDialogRef.current === "custom") {
        setDateRangePreset("custom");
        return;
      }
      setDateRangePreset(
        detectDateRangePreset(
          rangeFrom,
          rangeTo,
          lastUpdatedAt,
          datePickerMinDate
        )
      );
    },
    [
      rollbackRangeFrom,
      rollbackRangeTo,
      rangeFrom,
      rangeTo,
      lastUpdatedAt,
      datePickerMinDate,
    ]
  );

  const commitDateRange = useCallback(
    (e?: React.MouseEvent<unknown>) => {
      e?.preventDefault();
      e?.stopPropagation();
      setDateRangeDialogHidden(true);
      commitRangeFrom();
      commitRangeTo();
      setDateRangePreset("custom");
      setOffset(0);
    },
    [commitRangeFrom, commitRangeTo, setOffset]
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
  }, [sortDirection, setSortDirection]);

  const onTabChange = useCallback(
    (value: string) => {
      if (!isAuditLogKind(value) || value === auditLogKind) {
        return;
      }
      setAuditLogKind(value);
      setOffset(0);
      setFilters({
        searchKeyword: "",
        activityTypes: [],
      });
    },
    [auditLogKind, setOffset]
  );
  return (
    <>
      <div className={styles.root}>
        <div className={styles.header}>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="AuditLogScreen.title" />
          </Text>
          {logRetrievalDays !== -1 ? (
            <FeatureDisabledMessageBar
              className={styles.messageBar}
              messageID="FeatureConfig.audit-log.retrieval-days"
              messageValues={{ logRetrievalDays: logRetrievalDays }}
            />
          ) : null}
          <Tabs.Root
            value={auditLogKind}
            onValueChange={onTabChange}
            className={styles.pivot}
          >
            <Tabs.List>
              <Tabs.Trigger value={AuditLogKind.User}>
                {renderToString("AuditLogScreen.acitity-kind.user")}
              </Tabs.Trigger>
              <Tabs.Trigger value={AuditLogKind.Admin}>
                {renderToString("AuditLogScreen.acitity-kind.admin")}
              </Tabs.Trigger>
            </Tabs.List>
          </Tabs.Root>
        </div>
        <AuditLogFilterBar
          filters={filters}
          onFilterChange={onFilterChange}
          searchBoxProps={searchBoxProps}
          dateRange={filtersDateRange}
          availableActivityTypes={availableActivityTypes}
          onRefresh={onClickRefresh}
          lastUpdatedAt={lastUpdatedAt}
          wideActivityTypeDropdown={auditLogKind === AuditLogKind.Admin}
        />
        <div className={styles.listContainer}>
          <CommandBarContainer
            messageBar={messageBar}
            className={styles.commandBarContainerContent}
            headerPosition="static"
          >
            <AuditLogList
              className={styles.list}
              loading={listLoading}
              auditLogs={auditLogs}
              searchParams={searchParams.toString()}
              offset={offset}
              pageSize={pageSize}
              totalCount={auditLogs?.totalCount ?? undefined}
              onChangeOffset={onChangeOffset}
              onToggleSortDirection={onToggleSortDirection}
              sortDirection={sortDirection}
            />
          </CommandBarContainer>
        </div>
      </div>
      <AuditLogDateRangeDialog
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
