import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import {
  Callout,
  Checkbox,
  ColumnActionsMode,
  ComboBox,
  DetailsListLayoutMode,
  DetailsRow,
  DirectionalHint,
  Dropdown,
  IColumn,
  IComboBox,
  IComboBoxOption,
  IDetailsList,
  IDetailsRowProps,
  IDropdownOption,
  MessageBar,
  SearchBox,
  SelectionMode,
  ShimmeredDetailsList,
} from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import ShowError from "../../ShowError";
import WidgetTitle from "../../WidgetTitle";
import PaginationWidget from "../../PaginationWidget";
import CommandBarButton from "../../CommandBarButton";
import PrimaryButton from "../../PrimaryButton";
import DateRangeDialog from "../../graphql/portal/DateRangeDialog";
import { DateRangeFilterDropdown } from "../audit-log/DateRangeFilterDropdown";
import useTransactionalState from "../../hook/useTransactionalState";
import { FraudProtectionWarningType } from "../../types";
import {
  FraudProtectionLogsQueryQuery,
  useFraudProtectionLogsQueryQuery,
} from "../../graphql/adminapi/query/fraudProtectionLogsQuery.generated";
import {
  FraudProtectionDecision,
  SortDirection,
} from "../../graphql/adminapi/globalTypes.generated";
import { encodeOffsetToCursor } from "../../util/pagination";
import { useDebounced } from "../../hook/useDebounced";
import {
  formatCustomDateRangeLabel,
  formatDatetime,
} from "../../util/formatDatetime";
import styles from "./FraudProtectionLogsTab.module.css";

const PAGE_SIZE = 20;

type ResultFilterKey = "all" | "allowed" | "flagged" | "blocked";
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

function getResultQueryVariables(resultFilter: ResultFilterKey): {
  maximumWarningCount?: number;
  minimumWarningCount?: number;
  verdicts?: FraudProtectionDecision[];
} {
  switch (resultFilter) {
    case "allowed":
      return {
        verdicts: [FraudProtectionDecision.Allowed],
        maximumWarningCount: 0,
      };
    case "flagged":
      return {
        verdicts: [FraudProtectionDecision.Allowed],
        minimumWarningCount: 1,
      };
    case "blocked":
      return { verdicts: [FraudProtectionDecision.Blocked] };
    case "all":
      return {};
  }
}

function getResultMessageID(entry: FraudProtectionLogEntry): string {
  if (entry.decision === FraudProtectionDecision.Blocked) {
    return "FraudProtectionConfigurationScreen.logs.result.blocked";
  }
  if (entry.reasonCodes.length > 0) {
    return "FraudProtectionConfigurationScreen.logs.result.flagged";
  }
  return "FraudProtectionConfigurationScreen.logs.result.allowed";
}

function getResultClassName(entry: FraudProtectionLogEntry): string {
  if (entry.decision === FraudProtectionDecision.Blocked) {
    return styles.resultBlocked;
  }
  if (entry.reasonCodes.length > 0) {
    return styles.resultFlagged;
  }
  return styles.resultAllowed;
}

type FraudProtectionLogNode = NonNullable<
  NonNullable<
    NonNullable<FraudProtectionLogsQueryQuery["fraudProtectionLogs"]>["edges"]
  >[number]
>["node"];

function mapLogNodeToEntry(
  node: NonNullable<FraudProtectionLogNode>
): FraudProtectionLogEntry {
  // actionDetail only has one variant today (SendSMS), so this check is
  // trivially true, but it guards against other FraudProtectionAction
  // variants gaining their own actionDetail type in the future.
  const phoneNumber =
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    node.actionDetail.__typename ===
    "FraudProtectionDecisionSendSMSActionDetail"
      ? node.actionDetail.recipient
      : "";
  const phoneCountryCode =
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    node.actionDetail.__typename ===
    "FraudProtectionDecisionSendSMSActionDetail"
      ? node.actionDetail.phoneNumberCountryCode ?? ""
      : "";

  return {
    id: node.id,
    createdAt: node.createdAt,
    decision: node.decision,
    reasonCodes: node.triggeredWarnings,
    ipAddress: node.ipAddress ?? "",
    geoLocationCode: node.geoLocationCode ?? "",
    userAgent: node.userAgent ?? "",
    phoneNumber,
    phoneCountryCode,
  };
}

// ---- ResultCell ----

interface ResultCellProps {
  entry: FraudProtectionLogEntry;
}

const ResultCell: React.VFC<ResultCellProps> = function ResultCell({ entry }) {
  const { renderToString } = useContext(Context);
  return (
    <span className={`${styles.resultBadge} ${getResultClassName(entry)}`}>
      {renderToString(getResultMessageID(entry))}
    </span>
  );
};

// ---- Column definitions ----

type ColumnKey =
  | "timestamp"
  | "action"
  | "result"
  | "ip"
  | "reasonCodes"
  | "ipCountry"
  | "phone"
  | "phoneCountry";

interface ColumnDef {
  key: ColumnKey;
  alwaysShown: boolean;
  defaultVisible: boolean;
  minWidth: number;
  maxWidth?: number;
  fieldName?: string;
}

const COLUMN_DEFS: ColumnDef[] = [
  {
    key: "timestamp",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 200,
    maxWidth: 240,
    fieldName: "createdAt",
  },
  {
    key: "action",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 80,
    maxWidth: 100,
  },
  {
    key: "result",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 80,
    maxWidth: 100,
  },
  {
    key: "reasonCodes",
    alwaysShown: false,
    defaultVisible: true,
    minWidth: 480,
  },
  {
    key: "ip",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 120,
    maxWidth: 150,
  },
  {
    key: "ipCountry",
    alwaysShown: false,
    defaultVisible: true,
    minWidth: 100,
    maxWidth: 130,
  },
  {
    key: "phone",
    alwaysShown: false,
    defaultVisible: false,
    minWidth: 140,
    maxWidth: 180,
  },
  {
    key: "phoneCountry",
    alwaysShown: false,
    defaultVisible: false,
    minWidth: 100,
    maxWidth: 130,
  },
];

const FRAUD_PROTECTION_LOGS_COLUMNS_STORAGE_KEY_PREFIX =
  "fraud-protection-logs-visible-columns-v2:";

const OPTIONAL_COLUMN_KEYS = new Set<ColumnKey>(
  COLUMN_DEFS.filter((c) => !c.alwaysShown).map((c) => c.key)
);

function getDefaultVisibleOptionalColumns(): Set<ColumnKey> {
  return new Set(
    COLUMN_DEFS.filter((c) => !c.alwaysShown && c.defaultVisible).map(
      (c) => c.key
    )
  );
}

function loadVisibleOptionalColumns(appID: string): Set<ColumnKey> {
  try {
    const raw = window.localStorage.getItem(
      `${FRAUD_PROTECTION_LOGS_COLUMNS_STORAGE_KEY_PREFIX}${appID}`
    );
    if (raw == null) {
      return getDefaultVisibleOptionalColumns();
    }
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) {
      return getDefaultVisibleOptionalColumns();
    }
    // Honor an explicit empty selection ([]). Only fall back to defaults when
    // the preference has never been saved (raw == null) or the value is malformed.
    const keys = parsed.filter(
      (key): key is ColumnKey =>
        typeof key === "string" && OPTIONAL_COLUMN_KEYS.has(key as ColumnKey)
    );
    return new Set(keys);
  } catch {
    return getDefaultVisibleOptionalColumns();
  }
}

function saveVisibleOptionalColumns(
  appID: string,
  columns: Set<ColumnKey>
): void {
  window.localStorage.setItem(
    `${FRAUD_PROTECTION_LOGS_COLUMNS_STORAGE_KEY_PREFIX}${appID}`,
    JSON.stringify([...columns])
  );
}

// ---- ColumnsDropdown ----

interface ColumnsDropdownProps {
  columnDefs: ColumnDef[];
  visibleOptionalColumns: Set<ColumnKey>;
  onSaveColumns: (columns: Set<ColumnKey>) => void;
}

const ColumnsDropdown: React.VFC<ColumnsDropdownProps> =
  function ColumnsDropdown({
    columnDefs,
    visibleOptionalColumns,
    onSaveColumns,
  }) {
    const { renderToString } = useContext(Context);
    const [isOpen, setIsOpen] = useState(false);
    const [draftOptionalColumns, setDraftOptionalColumns] = useState<
      Set<ColumnKey>
    >(() => new Set(visibleOptionalColumns));
    const buttonRef = useRef<HTMLButtonElement | null>(null);

    const onOpen = useCallback(() => {
      setDraftOptionalColumns(new Set(visibleOptionalColumns));
      setIsOpen(true);
    }, [visibleOptionalColumns]);

    const onClose = useCallback(() => {
      setIsOpen(false);
    }, []);

    const alwaysShown = useMemo(
      () => columnDefs.filter((c) => c.alwaysShown),
      [columnDefs]
    );
    const optional = useMemo(
      () => columnDefs.filter((c) => !c.alwaysShown),
      [columnDefs]
    );

    const onToggleDraftColumn = useCallback((key: ColumnKey) => {
      setDraftOptionalColumns((prev) => {
        const next = new Set(prev);
        if (next.has(key)) {
          next.delete(key);
        } else {
          next.add(key);
        }
        return next;
      });
    }, []);

    const onClickSave = useCallback(() => {
      onSaveColumns(draftOptionalColumns);
      onClose();
    }, [draftOptionalColumns, onSaveColumns, onClose]);

    return (
      <div className={styles.columnsDropdownWrapper}>
        <button
          ref={buttonRef}
          type="button"
          className={styles.columnsButton}
          onClick={isOpen ? onClose : onOpen}
        >
          <span className={styles.columnsButtonIcon}>
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <rect
                x="1"
                y="2"
                width="3"
                height="10"
                rx="0.5"
                fill="currentColor"
                opacity="0.7"
              />
              <rect
                x="5.5"
                y="2"
                width="3"
                height="10"
                rx="0.5"
                fill="currentColor"
              />
              <rect
                x="10"
                y="2"
                width="3"
                height="10"
                rx="0.5"
                fill="currentColor"
                opacity="0.7"
              />
            </svg>
          </span>
          {renderToString(
            "FraudProtectionConfigurationScreen.logs.columns.button"
          )}
        </button>
        {isOpen ? (
          <Callout
            target={buttonRef}
            onDismiss={onClose}
            directionalHint={DirectionalHint.bottomLeftEdge}
            gapSpace={4}
            isBeakVisible={false}
            doNotLayer={true}
          >
            <div className={styles.columnsCallout}>
              <div className={styles.columnsSectionLabel}>
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.columns.alwaysShown"
                )}
              </div>
              {alwaysShown.map((col) => (
                <Checkbox
                  key={col.key}
                  className={styles.columnsCheckboxAlways}
                  checked={true}
                  disabled={true}
                  label={renderToString(
                    `FraudProtectionConfigurationScreen.logs.column.${col.key}`
                  )}
                />
              ))}
              <div
                className={`${styles.columnsSectionLabel} ${styles.columnsSectionLabelOptional}`}
              >
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.columns.optional"
                )}
              </div>
              {optional.map((col) => (
                <Checkbox
                  key={col.key}
                  className={styles.columnsCheckbox}
                  checked={draftOptionalColumns.has(col.key)}
                  onChange={() => onToggleDraftColumn(col.key)}
                  label={renderToString(
                    `FraudProtectionConfigurationScreen.logs.column.${col.key}`
                  )}
                />
              ))}
              <div className={styles.columnsCalloutFooter}>
                <PrimaryButton
                  text={<FormattedMessage id="save" />}
                  onClick={onClickSave}
                />
              </div>
            </div>
          </Callout>
        ) : null}
      </div>
    );
  };

// ---- FraudProtectionLogsTab ----

const FraudProtectionLogsTab: React.VFC = function FraudProtectionLogsTab() {
  const { renderToString, locale } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const [offset, setOffset] = useState(0);
  const [sortDirection] = useState(SortDirection.Desc);
  const [actionFilter] = useState<ActionFilterKey>("smsotp");
  const [resultFilter, setResultFilter] = useState<ResultFilterKey>("all");
  const [selectedReasonCodes, setSelectedReasonCodes] = useState<string[]>([]);
  const [searchText, setSearchText] = useState("");
  const detailsListRef = useRef<IDetailsList | null>(null);
  const [visibleOptionalColumns, setVisibleOptionalColumns] = useState<
    Set<ColumnKey>
  >(() => loadVisibleOptionalColumns(appID));

  const onSaveColumns = useCallback(
    (columns: Set<ColumnKey>) => {
      setVisibleOptionalColumns(columns);
      saveVisibleOptionalColumns(appID, columns);
    },
    [appID]
  );
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
      return rangeTo.toISOString();
    }
    return lastUpdatedAt.toISOString();
  }, [rangeTo, lastUpdatedAt]);

  const isCustomDateRange = rangeFrom != null || rangeTo != null;

  const customDateRangeLabel = useMemo(
    () =>
      isCustomDateRange
        ? formatCustomDateRangeLabel(rangeFrom, rangeTo)
        : undefined,
    [isCustomDateRange, rangeFrom, rangeTo]
  );

  const cursor = useMemo(() => encodeOffsetToCursor(offset), [offset]);

  const [debouncedSearch] = useDebounced(searchText, 300);
  const resultQueryVariables = useMemo(
    () => getResultQueryVariables(resultFilter),
    [resultFilter]
  );

  const { data, loading, error, refetch } = useFraudProtectionLogsQueryQuery({
    variables: {
      pageSize: PAGE_SIZE,
      cursor,
      rangeFrom: queryRangeFrom,
      rangeTo: queryRangeTo,
      sortDirection,
      verdicts: resultQueryVariables.verdicts,
      reasonCodes:
        selectedReasonCodes.length > 0 ? selectedReasonCodes : undefined,
      maximumWarningCount: resultQueryVariables.maximumWarningCount,
      minimumWarningCount: resultQueryVariables.minimumWarningCount,
      search:
        debouncedSearch.trim() !== "" ? debouncedSearch.trim() : undefined,
    },
    fetchPolicy: "network-only",
  });

  const allEntries = useMemo<FraudProtectionLogEntry[]>(() => {
    const edges = data?.fraudProtectionLogs?.edges ?? [];
    return edges
      .map((edge) => edge?.node)
      .filter(
        (node): node is NonNullable<FraudProtectionLogNode> => node != null
      )
      .map(mapLogNodeToEntry);
  }, [data]);

  const entries = useMemo<FraudProtectionLogEntryViewModel[]>(
    () =>
      allEntries.map((entry) => ({
        ...entry,
        isExpanded: false,
      })),
    [allEntries]
  );

  const rowItems = useMemo<FraudProtectionLogRowItem[]>(
    () => entries.map((entry) => ({ kind: "entry" as const, entry })),
    [entries]
  );

  // Reset offset when filters change
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setOffset(0);
  }, [
    resultFilter,
    selectedReasonCodes,
    debouncedSearch,
    queryRangeFrom,
    queryRangeTo,
  ]);

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
      } else if (uncommittedRangeFrom != null && value < uncommittedRangeFrom) {
        setRangeFrom(value);
        setRangeTo(uncommittedRangeFrom);
      } else {
        setRangeTo(value);
      }
    },
    [setRangeTo, setRangeFrom, uncommittedRangeFrom]
  );

  const resultOptions = useMemo<IDropdownOption[]>(
    () => [
      {
        key: "all",
        text: renderToString(
          "FraudProtectionConfigurationScreen.logs.result.all"
        ),
      },
      {
        key: "allowed",
        text: renderToString(
          "FraudProtectionConfigurationScreen.logs.result.allowed"
        ),
      },
      {
        key: "flagged",
        text: renderToString(
          "FraudProtectionConfigurationScreen.logs.result.flagged"
        ),
      },
      {
        key: "blocked",
        text: renderToString(
          "FraudProtectionConfigurationScreen.logs.result.blocked"
        ),
      },
    ],
    [renderToString]
  );

  const onChangeResult = useCallback(
    (_e: unknown, option?: IDropdownOption) => {
      if (option?.key != null) {
        setResultFilter(option.key as ResultFilterKey);
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
      if (item.kind === "details") return null;
      const entry = item.entry;
      switch (column?.key) {
        case "columnSettings":
          return null;
        case "timestamp":
          return <span>{formatDatetime(locale, entry.createdAt) ?? "—"}</span>;
        case "action":
          return (
            <span className={styles.actionCell}>
              <FormattedMessage id="FraudProtectionConfigurationScreen.logs.action.smsotp" />
            </span>
          );
        case "result":
          return <ResultCell entry={entry} />;
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
        case "ipCountry":
          return <span>{entry.geoLocationCode || "—"}</span>;
        case "phoneCountry":
          return <span>{entry.phoneCountryCode || "—"}</span>;
        case "phone":
          return <span>{entry.phoneNumber || "—"}</span>;
        default: {
          const fieldName = column?.fieldName as
            | keyof FraudProtectionLogEntryViewModel
            | undefined;
          const value = fieldName != null ? (entry[fieldName] as string) : "";
          return <span>{value}</span>;
        }
      }
    },
    [locale]
  );

  const onRenderRow = useCallback(
    (rowProps?: IDetailsRowProps) => {
      if (rowProps == null) return null;
      const item = rowProps.item as FraudProtectionLogRowItem | undefined;
      if (item == null) {
        return <DetailsRow {...rowProps} />;
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
          <DetailsRow {...rowProps} />
        </div>
      );
    },
    [onClickRow]
  );

  const onChangeOffset = useCallback((newOffset: number) => {
    setOffset(newOffset);
  }, []);

  const localizedColumns = useMemo<IColumn[]>(() => {
    return COLUMN_DEFS.filter(
      (def) => def.alwaysShown || visibleOptionalColumns.has(def.key)
    ).map((def) => ({
      key: def.key,
      name: renderToString(
        `FraudProtectionConfigurationScreen.logs.column.${def.key}`
      ),
      fieldName: def.fieldName,
      minWidth: def.minWidth,
      maxWidth: def.maxWidth,
      columnActionsMode: ColumnActionsMode.disabled,
    }));
  }, [renderToString, visibleOptionalColumns]);

  const totalCount = data?.fraudProtectionLogs?.totalCount ?? 0;
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
              className={cn(
                styles.dateRangeFilter,
                isCustomDateRange && styles.dateRangeFilterCustom
              )}
              value={isCustomDateRange ? "customDateRange" : "allDateRange"}
              customRangeLabel={customDateRangeLabel}
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
              className={styles.resultDropdown}
              selectedKey={resultFilter}
              options={resultOptions}
              onChange={onChangeResult}
            />
          </div>
          <div className={styles.bottomRow}>
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
              className={styles.searchBoxFilter}
              value={searchText}
              onChange={onChangeSearch}
              onClear={onClearSearch}
              placeholder={renderToString(
                "FraudProtectionConfigurationScreen.logs.search.placeholder"
              )}
            />
            <div className={styles.filterActions}>
              <ColumnsDropdown
                columnDefs={COLUMN_DEFS}
                visibleOptionalColumns={visibleOptionalColumns}
                onSaveColumns={onSaveColumns}
              />
              <CommandBarButton
                className={styles.refreshButton}
                iconProps={{ iconName: "Sync" }}
                text={renderToString(
                  "FraudProtectionConfigurationScreen.logs.refresh"
                )}
                onClick={onClickRefresh}
              />
            </div>
          </div>
        </div>

        {!isEmpty ? (
          <>
            <div className={styles.tableWrapper}>
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
                  return `${item.entry.id}-row`;
                }}
                selectionMode={SelectionMode.none}
                layoutMode={DetailsListLayoutMode.justified}
                onRenderRow={onRenderRow}
                onRenderItemColumn={onRenderItemColumn}
                className={styles.list}
              />
            </div>

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
        showTimePicker={true}
      />
    </section>
  );
};

export default FraudProtectionLogsTab;
