import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import cn from "classnames";
import { Button, Checkbox, Popover, Select, Text } from "@radix-ui/themes";
import {
  ChevronDownIcon,
  Cross2Icon,
  MixerHorizontalIcon,
  UpdateIcon,
} from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "../../intl";
import ShowError from "../../ShowError";
import PaginationWidget from "../../PaginationWidget";
import { TextField, TextFieldIcon } from "../v2/TextField/TextField";
import FraudProtectionDateRangeDialog from "./FraudProtectionDateRangeDialog";
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
const SHIMMER_ROW_COUNT = 8;

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

const ResultCell: React.VFC<{ entry: FraudProtectionLogEntry }> =
  function ResultCell({ entry }) {
    const { renderToString } = useContext(Context);
    return (
      <span className={cn(styles.resultBadge, getResultClassName(entry))}>
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
  grow?: boolean;
  /** Caps a grow column so long content truncates instead of widening it. */
  maxWidth?: number;
}

const COLUMN_DEFS: ColumnDef[] = [
  {
    key: "timestamp",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 210,
  },
  {
    key: "action",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 96,
  },
  {
    key: "result",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 100,
  },
  {
    key: "reasonCodes",
    alwaysShown: false,
    defaultVisible: true,
    minWidth: 200,
    grow: true,
    maxWidth: 320,
  },
  {
    key: "ip",
    alwaysShown: true,
    defaultVisible: true,
    minWidth: 140,
  },
  {
    key: "ipCountry",
    alwaysShown: false,
    defaultVisible: true,
    minWidth: 120,
  },
  {
    key: "phone",
    alwaysShown: false,
    defaultVisible: false,
    minWidth: 160,
  },
  {
    key: "phoneCountry",
    alwaysShown: false,
    defaultVisible: false,
    minWidth: 120,
  },
];

function columnStyle(def: ColumnDef): React.CSSProperties {
  if (def.grow === true) {
    // maxWidth caps the flex item so long content truncates (via the cell's
    // overflow:hidden + .cellText) instead of stretching the table and
    // pushing later columns (IP, country) out of view.
    return { flex: "1 1 0", minWidth: def.minWidth, maxWidth: def.maxWidth };
  }
  return { width: def.minWidth, minWidth: def.minWidth, flexShrink: 0 };
}

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

    const alwaysShown = useMemo(
      () => columnDefs.filter((c) => c.alwaysShown),
      [columnDefs]
    );
    const optional = useMemo(
      () => columnDefs.filter((c) => !c.alwaysShown),
      [columnDefs]
    );

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (open) {
          setDraftOptionalColumns(new Set(visibleOptionalColumns));
        }
        setIsOpen(open);
      },
      [visibleOptionalColumns]
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
      setIsOpen(false);
    }, [draftOptionalColumns, onSaveColumns]);

    return (
      <Popover.Root open={isOpen} onOpenChange={onOpenChange}>
        <Popover.Trigger>
          <button type="button" className={styles.filterButton}>
            <MixerHorizontalIcon className={styles.filterButtonIcon} />
            {renderToString(
              "FraudProtectionConfigurationScreen.logs.columns.button"
            )}
          </button>
        </Popover.Trigger>
        <Popover.Content
          className={styles.columnsContent}
          sideOffset={4}
          align="end"
        >
          <Text as="p" size="1" weight="medium" className={styles.columnsLabel}>
            {renderToString(
              "FraudProtectionConfigurationScreen.logs.columns.alwaysShown"
            )}
          </Text>
          {alwaysShown.map((col) => (
            <label key={col.key} className={styles.columnsItem}>
              <Checkbox checked={true} disabled={true} />
              <Text as="span" size="2">
                {renderToString(
                  `FraudProtectionConfigurationScreen.logs.column.${col.key}`
                )}
              </Text>
            </label>
          ))}
          <Text
            as="p"
            size="1"
            weight="medium"
            className={cn(styles.columnsLabel, styles.columnsLabelOptional)}
          >
            {renderToString(
              "FraudProtectionConfigurationScreen.logs.columns.optional"
            )}
          </Text>
          {optional.map((col) => (
            <label key={col.key} className={styles.columnsItem}>
              <Checkbox
                checked={draftOptionalColumns.has(col.key)}
                onCheckedChange={() => onToggleDraftColumn(col.key)}
              />
              <Text as="span" size="2">
                {renderToString(
                  `FraudProtectionConfigurationScreen.logs.column.${col.key}`
                )}
              </Text>
            </label>
          ))}
          <div className={styles.columnsFooter}>
            <Button size="2" variant="solid" onClick={onClickSave}>
              <FormattedMessage id="save" />
            </Button>
          </div>
        </Popover.Content>
      </Popover.Root>
    );
  };

// ---- ReasonCodesDropdown ----

interface ReasonCodesDropdownProps {
  selectedReasonCodes: string[];
  onChange: (codes: string[]) => void;
}

const ReasonCodesDropdown: React.VFC<ReasonCodesDropdownProps> =
  function ReasonCodesDropdown({ selectedReasonCodes, onChange }) {
    const { renderToString } = useContext(Context);
    const [isOpen, setIsOpen] = useState(false);

    const selectedSet = useMemo(
      () => new Set(selectedReasonCodes),
      [selectedReasonCodes]
    );

    const onToggle = useCallback(
      (code: string, checked: boolean) => {
        if (checked) {
          onChange([...selectedReasonCodes, code]);
        } else {
          onChange(selectedReasonCodes.filter((c) => c !== code));
        }
      },
      [onChange, selectedReasonCodes]
    );

    const onClear = useCallback(
      (e: React.MouseEvent | React.KeyboardEvent) => {
        e.preventDefault();
        e.stopPropagation();
        onChange([]);
      },
      [onChange]
    );

    const hasSelection = selectedReasonCodes.length > 0;
    const triggerLabel = hasSelection
      ? renderToString(
          "FraudProtectionConfigurationScreen.logs.reasonCodes.selected",
          { count: selectedReasonCodes.length }
        )
      : renderToString(
          "FraudProtectionConfigurationScreen.logs.reasonCodes.placeholder"
        );

    return (
      <Popover.Root open={isOpen} onOpenChange={setIsOpen}>
        <Popover.Trigger>
          <button
            type="button"
            className={cn(
              styles.filterSelectTrigger,
              styles.reasonCodesTrigger
            )}
          >
            <span
              className={cn(
                styles.filterSelectValue,
                !hasSelection && styles.filterSelectPlaceholder
              )}
            >
              {triggerLabel}
            </span>
            {hasSelection ? (
              <span
                role="button"
                tabIndex={0}
                className={styles.filterClearButton}
                aria-label={renderToString(
                  "FraudProtectionConfigurationScreen.logs.reasonCodes.clear"
                )}
                onPointerDown={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                }}
                onClick={onClear}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    onClear(e);
                  }
                }}
              >
                <Cross2Icon className={styles.filterSelectIcon} />
              </span>
            ) : (
              <ChevronDownIcon className={styles.filterSelectIcon} />
            )}
          </button>
        </Popover.Trigger>
        <Popover.Content
          className={styles.reasonCodesContent}
          sideOffset={4}
          align="start"
        >
          {KNOWN_REASON_CODES.map((code) => (
            <label key={code} className={styles.columnsItem}>
              <Checkbox
                checked={selectedSet.has(code)}
                onCheckedChange={(checked) => onToggle(code, checked === true)}
              />
              <Text as="span" size="2">
                {renderToString(
                  `FraudProtectionConfigurationScreen.logs.reasonCode.${code}`
                )}
              </Text>
            </label>
          ))}
        </Popover.Content>
      </Popover.Root>
    );
  };

// ---- Cell rendering ----

function renderCellContent(
  columnKey: ColumnKey,
  entry: FraudProtectionLogEntry,
  locale: string
): React.ReactNode {
  switch (columnKey) {
    case "timestamp":
      return formatDatetime(locale, entry.createdAt) ?? "—";
    case "action":
      return (
        <span className={styles.actionCell}>
          <FormattedMessage id="FraudProtectionConfigurationScreen.logs.action.smsotp" />
        </span>
      );
    case "result":
      return <ResultCell entry={entry} />;
    case "reasonCodes":
      return entry.reasonCodes.length > 0 ? entry.reasonCodes.join(", ") : "—";
    case "ip":
      return entry.ipAddress || "—";
    case "ipCountry":
      return entry.geoLocationCode || "—";
    case "phone":
      return entry.phoneNumber || "—";
    case "phoneCountry":
      return entry.phoneCountryCode || "—";
    default:
      return "—";
  }
}

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
        ? formatCustomDateRangeLabel(locale, rangeFrom, rangeTo)
        : undefined,
    [isCustomDateRange, locale, rangeFrom, rangeTo]
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

  const entries = useMemo<FraudProtectionLogEntry[]>(() => {
    const edges = data?.fraudProtectionLogs?.edges ?? [];
    return edges
      .map((edge) => edge?.node)
      .filter(
        (node): node is NonNullable<FraudProtectionLogNode> => node != null
      )
      .map(mapLogNodeToEntry);
  }, [data]);

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

  const onChangeResult = useCallback((value: string) => {
    setResultFilter(value as ResultFilterKey);
  }, []);

  const onChangeSearch = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchText(e.currentTarget.value);
    },
    []
  );

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

  const onChangeOffset = useCallback((newOffset: number) => {
    setOffset(newOffset);
  }, []);

  const visibleColumns = useMemo<ColumnDef[]>(
    () =>
      COLUMN_DEFS.filter(
        (def) => def.alwaysShown || visibleOptionalColumns.has(def.key)
      ),
    [visibleOptionalColumns]
  );

  const totalCount = data?.fraudProtectionLogs?.totalCount ?? 0;
  const isEmpty = !loading && entries.length === 0;

  if (error != null) {
    return <ShowError error={error} onRetry={onClickRefresh} />;
  }

  return (
    <section className={styles.section}>
      <Text as="p" size="4" weight="bold" className={styles.title}>
        <FormattedMessage id="FraudProtectionConfigurationScreen.tab.logs.title" />
      </Text>

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
          <Select.Root value={actionFilter} disabled={true} size="2">
            <Select.Trigger className={styles.filterSelect} />
            <Select.Content position="popper">
              <Select.Item value="smsotp">
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.action.smsotp"
                )}
              </Select.Item>
            </Select.Content>
          </Select.Root>
          <Select.Root
            value={resultFilter}
            onValueChange={onChangeResult}
            size="2"
          >
            <Select.Trigger className={styles.filterSelect} />
            <Select.Content position="popper">
              <Select.Item value="all">
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.result.all"
                )}
              </Select.Item>
              <Select.Item value="allowed">
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.result.allowed"
                )}
              </Select.Item>
              <Select.Item value="flagged">
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.result.flagged"
                )}
              </Select.Item>
              <Select.Item value="blocked">
                {renderToString(
                  "FraudProtectionConfigurationScreen.logs.result.blocked"
                )}
              </Select.Item>
            </Select.Content>
          </Select.Root>
        </div>
        <div className={styles.bottomRow}>
          <ReasonCodesDropdown
            selectedReasonCodes={selectedReasonCodes}
            onChange={setSelectedReasonCodes}
          />
          <div className={styles.searchBoxFilter}>
            <TextField
              size="2"
              type="search"
              value={searchText}
              placeholder={renderToString(
                "FraudProtectionConfigurationScreen.logs.search.placeholder"
              )}
              iconStart={TextFieldIcon.MagnifyingGlass}
              onChange={onChangeSearch}
              suffixPlain={true}
              suffix={
                searchText !== "" ? (
                  <button
                    type="button"
                    className={styles.searchClearButton}
                    aria-label={renderToString(
                      "FraudProtectionConfigurationScreen.logs.search.clear"
                    )}
                    onClick={onClearSearch}
                  >
                    <Cross2Icon className={styles.searchClearIcon} />
                  </button>
                ) : undefined
              }
            />
          </div>
          <div className={styles.filterActions}>
            <ColumnsDropdown
              columnDefs={COLUMN_DEFS}
              visibleOptionalColumns={visibleOptionalColumns}
              onSaveColumns={onSaveColumns}
            />
            <Button
              type="button"
              size="2"
              variant="ghost"
              color="gray"
              highContrast={true}
              className={styles.refreshButton}
              onClick={onClickRefresh}
            >
              <UpdateIcon />
              {renderToString(
                "FraudProtectionConfigurationScreen.logs.refresh"
              )}
            </Button>
          </div>
        </div>
      </div>

      <div className={styles.tableWrapper}>
        <div className={styles.table}>
          <div className={styles.tableHeader}>
            {visibleColumns.map((def) => (
              <div
                key={def.key}
                className={styles.headerCell}
                style={columnStyle(def)}
              >
                {renderToString(
                  `FraudProtectionConfigurationScreen.logs.column.${def.key}`
                )}
              </div>
            ))}
          </div>
          {loading ? (
            Array.from({ length: SHIMMER_ROW_COUNT }).map((_, i) => (
              <div key={`shimmer-${i}`} className={styles.shimmerRow}>
                {visibleColumns.map((def) => (
                  <div
                    key={def.key}
                    className={styles.cell}
                    style={columnStyle(def)}
                  >
                    <div className={styles.shimmerCell} />
                  </div>
                ))}
              </div>
            ))
          ) : isEmpty ? (
            <div className={styles.emptyRow}>
              <Text size="2" className={styles.emptyText}>
                <FormattedMessage id="FraudProtectionConfigurationScreen.logs.empty" />
              </Text>
            </div>
          ) : (
            entries.map((entry) => (
              <div
                key={entry.id}
                className={styles.tableRow}
                role="button"
                tabIndex={0}
                onClick={() => onClickRow(entry.id)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.preventDefault();
                    onClickRow(entry.id);
                  }
                }}
              >
                {visibleColumns.map((def) => (
                  <div
                    key={def.key}
                    className={styles.cell}
                    style={columnStyle(def)}
                  >
                    <span className={styles.cellText}>
                      {renderCellContent(def.key, entry, locale)}
                    </span>
                  </div>
                ))}
              </div>
            ))
          )}
        </div>
      </div>

      {!isEmpty ? (
        <PaginationWidget
          className={styles.pagination}
          offset={offset}
          pageSize={PAGE_SIZE}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
      ) : null}

      <FraudProtectionDateRangeDialog
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
