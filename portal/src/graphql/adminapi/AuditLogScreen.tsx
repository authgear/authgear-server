import React, {
  useState,
  useMemo,
  useCallback,
  useContext,
  useEffect,
} from "react";
import {
  useParams,
  useSearchParams,
  URLSearchParamsInit,
} from "react-router-dom";
import {
  ICommandBarItemProps,
  addDays,
  TooltipHost,
  ITooltipHostStyles,
  ITooltipProps,
  DirectionalHint,
  IContextualMenuItem,
  Pivot,
  PivotItem,
  SearchBox,
} from "@fluentui/react";
import { useId } from "@fluentui/react-hooks";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
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
import styles from "./AuditLogScreen.module.css";
import { useAppFeatureConfigQuery } from "../portal/query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "../portal/FeatureDisabledMessageBar";
import CommandBarButton from "../../CommandBarButton";
import { useDebounced } from "../../hook/useDebounced";
import { toTypedID } from "../../util/graphql";
import { NodeType } from "./node";
import { validateEmail } from "../../util/email";
import { validatePhoneNumber } from "../../util/phone";

const pageSize = 100;

const ALL_ACTIVITY_TYPES = Object.values(AuditLogActivityType);
const ADMIN_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) =>
    activityType.startsWith("ADMIN_API") || activityType.startsWith("PROJECT")
);
const USER_ACTIVITY_TYPES = ALL_ACTIVITY_TYPES.filter(
  (activityType) => !ADMIN_ACTIVITY_TYPES.includes(activityType)
);

enum AuditLogKind {
  User = "user",
  Admin = "admin",
}
function isAuditLogKind(s: string): s is AuditLogKind {
  return Object.values(AuditLogKind).includes(s as AuditLogKind);
}

const ALL = "ALL" as const;

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
          datetime: DateTime.fromJSDate(props.lastUpdatedAt).toRelative({
            locale,
          }),
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

// eslint-disable-next-line complexity
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
  const queryActivityKind = searchParams.get("kind") ?? "";
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
  const [stateSelectedKey, setSelectedKey] = useState(queryActivityType ?? ALL);
  const [sortDirection, setSortDirection] =
    useState<SortDirection>(queryOrderBy);
  const [lastUpdatedAt, setLastUpdatedAt] = useState(
    queryLastUpdatedAt != null
      ? new Date(Number(queryLastUpdatedAt))
      : new Date()
  );
  const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);
  const [activityKind, setActivityKind] = useState<AuditLogKind>(() => {
    if (isAuditLogKind(queryActivityKind)) {
      return queryActivityKind;
    }
    return AuditLogKind.User;
  });
  const [searchQuery, setSearchQuery] = useState<string>(queryString);

  const {
    committedValue: rangeFrom,
    uncommittedValue: uncommittedRangeFrom,
    setValue: setRangeFrom,
    setCommittedValue: setRangeFromImmediately,
    commit: commitRangeFrom,
    rollback: rollbackRangeFrom,
  } = useTransactionalState<Date | null>(
    queryFrom != null && queryFrom !== "" ? new Date(queryFrom) : null
  );

  const {
    committedValue: rangeTo,
    uncommittedValue: uncommittedRangeTo,
    setValue: setRangeTo,
    setCommittedValue: setRangeToImmediately,
    commit: commitRangeTo,
    rollback: rollbackRangeTo,
  } = useTransactionalState<Date | null>(
    queryTo != null && queryTo !== "" ? new Date(queryTo) : null
  );

  const { appID } = useParams() as { appID: string };
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
    return lastUpdatedAt.toISOString();
  }, [rangeTo, lastUpdatedAt]);

  const isCustomDateRange = rangeFrom != null || rangeTo != null;

  const availableActivityTypes = useMemo(() => {
    return activityKind === "admin"
      ? ADMIN_ACTIVITY_TYPES
      : USER_ACTIVITY_TYPES;
  }, [activityKind]);

  const selectedKey = useMemo(() => {
    if (stateSelectedKey === ALL) {
      return stateSelectedKey;
    }
    if (
      availableActivityTypes.includes(stateSelectedKey as AuditLogActivityType)
    ) {
      return stateSelectedKey;
    }
    return ALL;
  }, [availableActivityTypes, stateSelectedKey]);

  const [debouncedSearchQuery] = useDebounced(searchQuery, 300);

  const { renderToString } = useContext(Context);

  // When the page is refreshed, and it is on the first page,
  // update last_updated_at.
  // Note that if the page is navigated from another page,
  // this effect is NOT run.
  // This is the intended behavior because we do not
  // want to change last_updated_at.
  useEffect(() => {
    if (queryPage === "1") {
      setLastUpdatedAt(new Date());
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Sync state to searchParams.
  // eslint-disable-next-line complexity
  useEffect(() => {
    const page = offset / pageSize + 1;

    const params: URLSearchParamsInit = {};

    const newQueryFrom =
      rangeFrom != null ? DateTime.fromJSDate(rangeFrom).toISODate() : "";
    const newQueryTo =
      rangeTo != null ? DateTime.fromJSDate(rangeTo).toISODate() : "";
    const newQueryOrderBy = sortDirection;
    const newQueryPage = page.toString();
    const newQueryActivityType = selectedKey;
    const newQueryLastUpdatedAt = lastUpdatedAt.getTime().toString();
    const newActivityKind = activityKind;
    const newQueryString = debouncedSearchQuery;

    params["from"] = newQueryFrom;
    params["to"] = newQueryTo;
    params["order_by"] = newQueryOrderBy;
    params["page"] = newQueryPage;
    params["activity_type"] = newQueryActivityType;
    params["last_updated_at"] = newQueryLastUpdatedAt;
    params["kind"] = newActivityKind;
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
    if (newActivityKind !== queryActivityKind) {
      callSet = true;
    }
    if (newQueryString !== queryString) {
      callSet = true;
    }

    if (callSet) {
      setSearchParams(params);
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
    selectedKey,
    lastUpdatedAt,
    setSearchParams,
    activityKind,
    queryActivityKind,
    debouncedSearchQuery,
    queryString,
  ]);

  const activityTypeOptions = useMemo(() => {
    const options = [
      {
        key: ALL as string,
        text: renderToString("AuditLogActivityType.ALL"),
      },
    ];
    for (const key of availableActivityTypes) {
      options.push({
        key: key,
        text: renderToString("AuditLogActivityType." + key),
      });
    }
    return options;
  }, [availableActivityTypes, renderToString]);

  const activityTypes: AuditLogActivityType[] | null = useMemo(() => {
    if (selectedKey === ALL) {
      return availableActivityTypes;
    }
    return [selectedKey] as AuditLogActivityType[];
  }, [availableActivityTypes, selectedKey]);

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

  const queryEmailAddresses = useMemo(() => {
    const email = validateEmail(debouncedSearchQuery);
    if (email == null) {
      return null;
    }

    switch (selectedKey) {
      // Only search email addresses if `all` or `email_sent` filter active
      case ALL:
      case AuditLogActivityType.EmailSent:
        return email ? [email] : null;
      default:
        return null;
    }
  }, [debouncedSearchQuery, selectedKey]);

  const queryPhoneNumbers = useMemo(() => {
    const phoneNumber = validatePhoneNumber(debouncedSearchQuery);
    if (phoneNumber == null) {
      return null;
    }
    switch (selectedKey) {
      // Only search phone numbers if `all` or `phone_sent` or `whatsapp_sent` filter active
      case ALL:
      case AuditLogActivityType.SmsSent:
      case AuditLogActivityType.WhatsappSent:
        return phoneNumber ? [phoneNumber] : null;
      default:
        return null;
    }
  }, [debouncedSearchQuery, selectedKey]);

  const userNodeIDs = useMemo(() => {
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
        userIDs: userNodeIDs,
        emailAddresses: queryEmailAddresses,
        phoneNumbers: queryPhoneNumbers,
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
    (
      e?: React.MouseEvent<unknown> | React.KeyboardEvent<unknown>,
      item?: IContextualMenuItem
    ) => {
      e?.stopPropagation();
      if (item != null && typeof item.key === "string") {
        setOffset(0);
        setSelectedKey(item.key);
      }
    },
    []
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

  const onChangeSearchQuery = useCallback(
    (e?: React.ChangeEvent<HTMLInputElement>) => {
      if (e === undefined) {
        return;
      }
      setSearchQuery(e.currentTarget.value);
    },
    []
  );

  const onClearSearchQuery = useCallback(() => {
    setSearchQuery("");
  }, []);

  const searchBoxPlaceholder = useMemo(() => {
    switch (selectedKey) {
      case AuditLogActivityType.EmailSent:
        return renderToString("AuditLogScreen.search-by-user-id-or-email");
      case AuditLogActivityType.SmsSent:
      case AuditLogActivityType.WhatsappSent:
        return renderToString("AuditLogScreen.search-by-user-id-or-phone");
      default:
        return renderToString("AuditLogScreen.search-by-user-id");
    }
  }, [renderToString, selectedKey]);

  const commandBarFarItems: ICommandBarItemProps[] = useMemo(() => {
    const allDateRangeLabel = renderToString("AuditLogScreen.date-range.all");
    const customDateRangeLabel = renderToString(
      "AuditLogScreen.date-range.custom"
    );
    const items: ICommandBarItemProps[] = [
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
        text: activityTypeOptions.find((option) => option.key === selectedKey)!
          .text,
        iconProps: {
          iconName: "PC1",
        },
        subMenuProps: {
          items: activityTypeOptions.map((option) => {
            return {
              key: option.key,
              text: option.text,
              onClick: onChangeSelectedKey,
            };
          }),
        },
      },
      {
        key: "search",
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: () => {
          return (
            <SearchBox
              placeholder={searchBoxPlaceholder}
              className={styles.searchBox}
              value={searchQuery}
              onChange={onChangeSearchQuery}
              onClear={onClearSearchQuery}
            />
          );
        },
      },
      {
        key: "clear",
        text: renderToString("AuditLogScreen.clear-all-filters"),
        iconProps: {
          iconName: "",
        },
        onClick: () => {
          setOffset(0);
          setRangeFromImmediately(null);
          setRangeToImmediately(null);
          setSelectedKey(ALL);
          setSearchQuery("");
        },
        buttonStyles: {
          root: {
            color: "rgba(89, 86, 83, 0.4)",
          },
        },
      },
    ];
    return items;
  }, [
    renderToString,
    isCustomDateRange,
    onClickAllDateRange,
    onClickCustomDateRange,
    activityTypeOptions,
    selectedKey,
    onChangeSelectedKey,
    searchBoxPlaceholder,
    searchQuery,
    onChangeSearchQuery,
    onClearSearchQuery,
    setRangeFromImmediately,
    setRangeToImmediately,
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

  const onTabChange = useCallback((item?: PivotItem) => {
    if (item == null) {
      return;
    }
    const { itemKey } = item.props;
    if (!itemKey || !isAuditLogKind(itemKey)) {
      return;
    }
    setActivityKind(itemKey);
    // Reset pagination on tab change
    setOffset(0);
  }, []);

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
          <Pivot
            className={styles.pivot}
            selectedKey={activityKind}
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
          </Pivot>
        </div>
        <div className={styles.listContainer}>
          <CommandBarContainer
            isLoading={loading}
            messageBar={messageBar}
            primaryItems={commandBarFarItems}
            secondaryItems={commandBarSecondaryItems}
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
