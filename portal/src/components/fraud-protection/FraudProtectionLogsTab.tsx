import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  ColumnActionsMode,
  ComboBox,
  DetailsListLayoutMode,
  DetailsRow,
  Dropdown,
  IColumn,
  IComboBox,
  IComboBoxOption,
  IDetailsList,
  IDetailsRowProps,
  IDropdownOption,
  IconButton,
  MessageBar,
  SearchBox,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import { DateTime } from "luxon";
import { Context, FormattedMessage } from "../../intl";
import ShowError from "../../ShowError";
import WidgetTitle from "../../WidgetTitle";
import PaginationWidget from "../../PaginationWidget";
import CommandBarButton from "../../CommandBarButton";
import DateRangeDialog from "../../graphql/portal/DateRangeDialog";
import { DateRangeFilterDropdown } from "../audit-log/DateRangeFilterDropdown";
import useTransactionalState from "../../hook/useTransactionalState";
import {
  FraudProtectionLogsQueryQuery,
  useFraudProtectionLogsQueryQuery,
} from "../../graphql/adminapi/query/fraudProtectionLogsQuery.generated";
import {
  FraudProtectionDecision,
  SortDirection,
} from "../../graphql/adminapi/globalTypes.generated";
import { encodeOffsetToCursor } from "../../util/pagination";
import { formatDatetime } from "../../util/formatDatetime";
import { useDebounced } from "../../hook/useDebounced";
import { FraudProtectionWarningType } from "../../types";
import styles from "./FraudProtectionLogsTab.module.css";

const PAGE_SIZE = 20;

type VerdictFilterKey = "all" | "allowed" | "blocked";
type ActionFilterKey = "smsotp";

const KNOWN_REASON_CODES = [
  FraudProtectionWarningType.SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED,
  FraudProtectionWarningType.SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED,
  FraudProtectionWarningType.SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED,
  FraudProtectionWarningType.SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED,
  FraudProtectionWarningType.SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED,
] as const;

interface FraudProtectionLogEntry {
  id: string;
  createdAt: string;
  decision: FraudProtectionDecision;
  reasonCodes: string[];
  ipAddress: string;
  geoLocationCode: string;
  userAgent: string;
  phoneNumber: string;
  phoneCountryCode: string;
}

interface FraudProtectionLogEntryViewModel extends FraudProtectionLogEntry {
  isExpanded: boolean;
}

interface FraudProtectionLogPrimaryRow {
  kind: "entry";
  entry: FraudProtectionLogEntryViewModel;
}

interface FraudProtectionLogDetailRow {
  kind: "details";
  id: string;
  entry: FraudProtectionLogEntryViewModel;
}

type FraudProtectionLogRowItem =
  | FraudProtectionLogPrimaryRow
  | FraudProtectionLogDetailRow;

function ensureFraudDecisionNodeID(id: string): string {
  try {
    const padding = "=".repeat((4 - (id.length % 4)) % 4);
    const decoded = atob(id.replace(/-/g, "+").replace(/_/g, "/") + padding);
    if (decoded.startsWith("FraudProtectionDecisionRecord:")) {
      return id;
    }
  } catch {
    // Not a base64url node ID.
  }
  const raw = `FraudProtectionDecisionRecord:${id}`;
  return btoa(raw).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}

function parseEntry(
  node: NonNullable<
    NonNullable<
      NonNullable<
        NonNullable<
          FraudProtectionLogsQueryQuery["fraudProtectionLogs"]
        >["edges"]
      >[number]
    >["node"]
  >,
  locale: string
): FraudProtectionLogEntry {
  const phoneNumber = node.actionDetail.recipient ?? "";
  return {
    id: node.id,
    createdAt: formatDatetime(locale, node.createdAt) ?? "",
    decision: node.decision,
    reasonCodes: node.triggeredWarnings,
    ipAddress: node.ipAddress ?? "",
    geoLocationCode: node.geoLocationCode ?? "",
    userAgent: node.userAgent ?? "",
    phoneNumber,
    phoneCountryCode: node.actionDetail.phoneNumberCountryCode ?? "",
  };
}

function getVerdictQueryVariables(
  verdictFilter: VerdictFilterKey
): FraudProtectionDecision[] | undefined {
  switch (verdictFilter) {
    case "allowed":
      return [FraudProtectionDecision.Allowed];
    case "blocked":
      return [FraudProtectionDecision.Blocked];
    case "all":
      return undefined;
  }
}

// ---- VerdictCell ----

interface VerdictCellProps {
  entry: FraudProtectionLogEntry;
}

const VerdictCell: React.VFC<VerdictCellProps> = function VerdictCell({
  entry,
}) {
  const { renderToString } = useContext(Context);
  if (entry.decision === FraudProtectionDecision.Blocked) {
    return (
      <span className={`${styles.verdictBadge} ${styles.verdictBlocked}`}>
        {renderToString(
          "FraudProtectionConfigurationScreen.logs.verdict.blocked"
        )}
      </span>
    );
  }
  return (
    <span className={`${styles.verdictBadge} ${styles.verdictAllowed}`}>
      {renderToString(
        "FraudProtectionConfigurationScreen.logs.verdict.allowed"
      )}
    </span>
  );
};

// ---- LogRowDetails ----

interface LogRowDetailsProps {
  entry: FraudProtectionLogEntryViewModel;
}

const LogRowDetails: React.VFC<LogRowDetailsProps> = function LogRowDetails({
  entry,
}) {
  const hasPhoneCountryCode = entry.phoneCountryCode !== "";

  return (
    <div className={styles.rowDetails}>
      <div className={styles.detailsGrid}>
        <div className={styles.detailsSection}>
          <div className={styles.detailsSectionTitle}>
            <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.ipDetails" />
          </div>
          <div className={styles.detailsField}>
            <span className={styles.detailsLabel}>
              <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.ip" />
            </span>
            <span className={styles.detailsValue}>
              {entry.ipAddress || "—"}
            </span>
          </div>
          <div className={styles.detailsField}>
            <span className={styles.detailsLabel}>
              <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.geoLocation" />
            </span>
            <span className={styles.detailsValue}>
              {entry.geoLocationCode || "—"}
            </span>
          </div>
        </div>

        <div className={styles.detailsSection}>
          <div className={styles.detailsSectionTitle}>
            <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.targetInfo" />
          </div>
          <div className={styles.detailsField}>
            <span className={styles.detailsLabel}>
              <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.phone" />
            </span>
            <span className={styles.detailsValue}>
              {entry.phoneNumber || "—"}
            </span>
          </div>
          {hasPhoneCountryCode ? (
            <div className={styles.detailsField}>
              <span className={styles.detailsLabel}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.phoneCountryCode" />
              </span>
              <span className={styles.detailsValue}>
                {entry.phoneCountryCode}
              </span>
            </div>
          ) : null}
        </div>

        <div className={styles.detailsSection}>
          <div className={styles.detailsSectionTitle}>
            <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.riskAssessment" />
          </div>
          <div className={styles.detailsField}>
            <span className={styles.detailsLabel}>
              <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.reasonCodes" />
            </span>
            {entry.reasonCodes.length > 0 ? (
              <div className={styles.reasonCodeTags}>
                {entry.reasonCodes.map((code) => (
                  <span key={code} className={styles.reasonCodeTag}>
                    {code}
                  </span>
                ))}
              </div>
            ) : (
              <span className={styles.detailsValue}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.details.none" />
              </span>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

// ---- FraudProtectionLogsTab ----

export interface FraudProtectionLogsTabProps {}

const columns: IColumn[] = [
  {
    key: "expand",
    name: "",
    minWidth: 32,
    maxWidth: 32,
    columnActionsMode: ColumnActionsMode.disabled,
  },
  {
    key: "timestamp",
    name: "Timestamp",
    fieldName: "createdAt",
    minWidth: 200,
    maxWidth: 240,
    columnActionsMode: ColumnActionsMode.disabled,
  },
  {
    key: "action",
    name: "Action",
    minWidth: 80,
    maxWidth: 100,
    columnActionsMode: ColumnActionsMode.disabled,
  },
  {
    key: "verdict",
    name: "Verdict",
    minWidth: 80,
    maxWidth: 100,
    columnActionsMode: ColumnActionsMode.disabled,
  },
  {
    key: "reasonCodes",
    name: "Reason codes",
    minWidth: 200,
    columnActionsMode: ColumnActionsMode.disabled,
  },
  {
    key: "ip",
    name: "IP",
    minWidth: 120,
    maxWidth: 150,
    columnActionsMode: ColumnActionsMode.disabled,
  },
];

const FraudProtectionLogsTab: React.VFC<FraudProtectionLogsTabProps> =
  function FraudProtectionLogsTab() {
    const { renderToString, locale } = useContext(Context);
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();

    const [offset, setOffset] = useState(0);
    const [sortDirection] = useState(SortDirection.Desc);
    const [actionFilter] = useState<ActionFilterKey>("smsotp");
    const [verdictFilter, setVerdictFilter] = useState<VerdictFilterKey>("all");
    const [selectedReasonCodes, setSelectedReasonCodes] = useState<string[]>(
      []
    );
    const [searchText, setSearchText] = useState("");
    const [expandedRowId, setExpandedRowId] = useState<string | null>(null);
    const detailsListRef = useRef<IDetailsList | null>(null);
    const [dateRangeDialogHidden, setDateRangeDialogHidden] = useState(true);
    const [lastUpdatedAt, setLastUpdatedAt] = useState(() => new Date());

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

    const queryRangeFrom = useMemo(
      () => (rangeFrom != null ? rangeFrom.toISOString() : null),
      [rangeFrom]
    );

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

    const cursor = useMemo(() => encodeOffsetToCursor(offset), [offset]);

    const [debouncedSearch] = useDebounced(searchText, 300);

    const { data, previousData, loading, error, refetch } =
      useFraudProtectionLogsQueryQuery({
        variables: {
          pageSize: PAGE_SIZE,
          cursor,
          rangeFrom: queryRangeFrom,
          rangeTo: queryRangeTo,
          sortDirection,
          verdicts: getVerdictQueryVariables(verdictFilter),
          reasonCodes:
            selectedReasonCodes.length > 0 ? selectedReasonCodes : undefined,
          search:
            debouncedSearch.trim() !== "" ? debouncedSearch.trim() : undefined,
        },
        fetchPolicy: "network-only",
      });

    const currentData = data ?? previousData;

    const allEntries = useMemo<FraudProtectionLogEntry[]>(() => {
      const edges = currentData?.fraudProtectionLogs?.edges ?? [];
      return edges
        .map((edge) => {
          const node = edge?.node;
          if (node == null) return null;
          return parseEntry(node, locale);
        })
        .filter((e): e is FraudProtectionLogEntry => e != null);
    }, [currentData, locale]);

    const entries = useMemo<FraudProtectionLogEntryViewModel[]>(
      () =>
        allEntries.map((entry) => ({
          ...entry,
          isExpanded: expandedRowId === entry.id,
        })),
      [allEntries, expandedRowId]
    );

    const rowItems = useMemo<FraudProtectionLogRowItem[]>(() => {
      const items: FraudProtectionLogRowItem[] = [];
      for (const entry of entries) {
        items.push({ kind: "entry", entry });
        if (entry.isExpanded) {
          items.push({
            kind: "details",
            id: `${entry.id}::details`,
            entry,
          });
        }
      }
      return items;
    }, [entries]);

    useEffect(() => {
      setOffset(0);
    }, [verdictFilter, selectedReasonCodes, debouncedSearch]);

    const onClickRefresh = useCallback(() => {
      setLastUpdatedAt(new Date());
      setOffset(0);
      void refetch();
    }, [refetch]);

    const onClickAllDateRange = useCallback(() => {
      setRangeFromImmediately(null);
      setRangeToImmediately(null);
    }, [setRangeFromImmediately, setRangeToImmediately]);

    const onClickCustomDateRange = useCallback(() => {
      setDateRangeDialogHidden(false);
    }, []);

    const onDismissDateRangeDialog = useCallback(() => {
      setDateRangeDialogHidden(true);
      rollbackRangeFrom();
      rollbackRangeTo();
    }, [rollbackRangeFrom, rollbackRangeTo]);

    const commitDateRange = useCallback(() => {
      setDateRangeDialogHidden(true);
      commitRangeFrom();
      commitRangeTo();
      setOffset(0);
    }, [commitRangeFrom, commitRangeTo]);

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

    const verdictOptions = useMemo<IDropdownOption[]>(
      () => [
        {
          key: "all",
          text: renderToString(
            "FraudProtectionConfigurationScreen.logs.verdict.all"
          ),
        },
        {
          key: "allowed",
          text: renderToString(
            "FraudProtectionConfigurationScreen.logs.verdict.allowed"
          ),
        },
        {
          key: "blocked",
          text: renderToString(
            "FraudProtectionConfigurationScreen.logs.verdict.blocked"
          ),
        },
      ],
      [renderToString]
    );

    const onChangeVerdict = useCallback(
      (_e: unknown, option?: IDropdownOption) => {
        if (option?.key != null) {
          setVerdictFilter(option.key as VerdictFilterKey);
        }
      },
      []
    );

    const actionOptions = useMemo<IDropdownOption[]>(
      () => [
        {
          key: "smsotp",
          text: renderToString(
            "FraudProtectionConfigurationScreen.logs.action.smsotp"
          ),
        },
      ],
      [renderToString]
    );

    const reasonCodeOptions = useMemo<IComboBoxOption[]>(
      () =>
        KNOWN_REASON_CODES.map((code) => ({
          key: code,
          text: renderToString(
            `FraudProtectionConfigurationScreen.logs.reasonCode.${code}`
          ),
        })),
      [renderToString]
    );

    const onChangeReasonCodes = useCallback(
      (_ev: React.FormEvent<IComboBox>, option?: IComboBoxOption) => {
        if (option == null) return;
        const key = option.key as string;
        setSelectedReasonCodes((prev) =>
          option.selected ? [...prev, key] : prev.filter((k) => k !== key)
        );
      },
      []
    );

    const onChangeSearch = useCallback((_: unknown, newValue?: string) => {
      setSearchText(newValue ?? "");
    }, []);

    const onClearSearch = useCallback(() => {
      setSearchText("");
    }, []);

    const onToggleRow = useCallback((id: string) => {
      setExpandedRowId((prev) => (prev === id ? null : id));
    }, []);

    const onClickRow = useCallback(
      (id: string) => {
        const nodeID = ensureFraudDecisionNodeID(id);
        navigate(
          `/project/${appID}/attack-protection/fraud-protection/logs/${nodeID}`
        );
      },
      [appID, navigate]
    );

    const onRenderItemColumn = useCallback(
      (item?: FraudProtectionLogRowItem, _index?: number, column?: IColumn) => {
        if (item == null) return null;
        if (item.kind === "details") {
          if (column?.key !== "details") return null;
          return <LogRowDetails entry={item.entry} />;
        }
        const entry = item.entry;
        switch (column?.key) {
          case "expand":
            return (
              <IconButton
                className={styles.expandButton}
                iconProps={{
                  iconName: entry.isExpanded ? "ChevronDown" : "ChevronRight",
                }}
                onClick={(e) => {
                  e.stopPropagation();
                  onToggleRow(entry.id);
                }}
              />
            );
          case "action":
            return (
              <span className={styles.actionCell}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.action.smsotp" />
              </span>
            );
          case "verdict":
            return <VerdictCell entry={entry} />;
          case "reasonCodes":
            return (
              <span>
                {entry.reasonCodes.length > 0
                  ? entry.reasonCodes.join(", ")
                  : "—"}
              </span>
            );
          case "ip":
            return <span>{entry.ipAddress || "—"}</span>;
          default: {
            const fieldName = column?.fieldName as
              | keyof FraudProtectionLogEntryViewModel
              | undefined;
            const value = fieldName != null ? (entry[fieldName] as string) : "";
            return <span>{value}</span>;
          }
        }
      },
      [onToggleRow]
    );

    const onRenderRow = useCallback(
      (rowProps?: IDetailsRowProps) => {
        if (rowProps == null) return null;
        const item = rowProps.item as FraudProtectionLogRowItem | undefined;
        if (item == null) {
          return <DetailsRow {...rowProps} />;
        }
        if (item.kind === "details") {
          return <LogRowDetails entry={item.entry} />;
        }
        return (
          <div
            className={styles.logRow}
            onClick={() => onClickRow(item.entry.id)}
            role="button"
            tabIndex={0}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                onClickRow(item.entry.id);
              }
            }}
          >
            <DetailsRow
              {...rowProps}
              className={item.entry.isExpanded ? styles.expandedRow : undefined}
            />
          </div>
        );
      },
      [onClickRow]
    );

    const onChangeOffset = useCallback((newOffset: number) => {
      setOffset(newOffset);
    }, []);

    const localizedColumns = useMemo<IColumn[]>(
      () =>
        columns.map((col) => ({
          ...col,
          name:
            col.key === "expand"
              ? ""
              : renderToString(
                  `FraudProtectionConfigurationScreen.logs.column.${col.key}`
                ),
        })),
      [renderToString]
    );

    const totalCount =
      currentData?.fraudProtectionLogs?.totalCount ?? undefined;
    const isEmpty = !loading && rowItems.length === 0;

    if (error != null) {
      return <ShowError error={error} onRetry={onClickRefresh} />;
    }

    return (
      <section className={styles.section}>
        <WidgetTitle>
          <FormattedMessage id="FraudProtectionConfigurationScreen.tab.logs.title" />
        </WidgetTitle>

        <div>
          <div className={styles.filterRow}>
            <div className={styles.filterGroup}>
              <DateRangeFilterDropdown
                value={isCustomDateRange ? "customDateRange" : "allDateRange"}
                onClickAllDateRange={onClickAllDateRange}
                onClickCustomDateRange={onClickCustomDateRange}
              />
              <Dropdown
                className={styles.actionDropdown}
                selectedKey={actionFilter}
                options={actionOptions}
                disabled={true}
              />
              <Dropdown
                className={styles.verdictDropdown}
                selectedKey={verdictFilter}
                options={verdictOptions}
                onChange={onChangeVerdict}
              />
              <ComboBox
                className={styles.reasonCodeComboBox}
                multiSelect={true}
                options={reasonCodeOptions}
                selectedKey={selectedReasonCodes}
                onChange={onChangeReasonCodes}
                placeholder={renderToString(
                  "FraudProtectionConfigurationScreen.logs.reasonCodes.placeholder"
                )}
                allowFreeInput={false}
              />
              <SearchBox
                className={styles.searchBox}
                value={searchText}
                onChange={onChangeSearch}
                onClear={onClearSearch}
                placeholder={renderToString(
                  "FraudProtectionConfigurationScreen.logs.search.placeholder"
                )}
              />
            </div>
            <div className={styles.filterActions}>
              <CommandBarButton
                iconProps={{ iconName: "Sync" }}
                text={renderToString(
                  "FraudProtectionConfigurationScreen.logs.refresh"
                )}
                onClick={onClickRefresh}
              />
            </div>
          </div>

          {!isEmpty ? (
            <>
              <ShimmeredDetailsList
                componentRef={detailsListRef}
                enableShimmer={loading}
                enableUpdateAnimations={false}
                items={rowItems}
                columns={localizedColumns}
                getKey={(item?: FraudProtectionLogRowItem, index?: number) => {
                  if (item == null) {
                    return String(index ?? "");
                  }
                  return item.kind === "details"
                    ? item.id
                    : `${item.entry.id}-detail`;
                }}
                selectionMode={SelectionMode.none}
                layoutMode={DetailsListLayoutMode.justified}
                onRenderRow={onRenderRow}
                onRenderItemColumn={onRenderItemColumn}
                className={styles.list}
              />

              <PaginationWidget
                className={styles.pagination}
                offset={offset}
                pageSize={PAGE_SIZE}
                totalCount={totalCount}
                onChangeOffset={onChangeOffset}
              />
            </>
          ) : (
            <MessageBar>
              <FormattedMessage id="FraudProtectionConfigurationScreen.logs.empty" />
            </MessageBar>
          )}
        </div>

        <DateRangeDialog
          hidden={dateRangeDialogHidden}
          title={renderToString(
            "FraudProtectionConfigurationScreen.logs.dateRange.dialog.title"
          )}
          fromDatePickerLabel={renderToString(
            "FraudProtectionConfigurationScreen.logs.dateRange.dialog.from"
          )}
          toDatePickerLabel={renderToString(
            "FraudProtectionConfigurationScreen.logs.dateRange.dialog.to"
          )}
          rangeFrom={uncommittedRangeFrom ?? undefined}
          rangeTo={uncommittedRangeTo ?? undefined}
          fromDatePickerMaxDate={lastUpdatedAt}
          toDatePickerMaxDate={lastUpdatedAt}
          onSelectRangeFrom={onSelectRangeFrom}
          onSelectRangeTo={onSelectRangeTo}
          onCommitDateRange={commitDateRange}
          onDismiss={onDismissDateRangeDialog}
        />
      </section>
    );
  };

export default FraudProtectionLogsTab;
