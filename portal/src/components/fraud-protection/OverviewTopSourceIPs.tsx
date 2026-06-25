import React, { useContext, useMemo } from "react";
import { Icon, Text } from "@fluentui/react";
import { Context } from "../../intl";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./OverviewTopSourceIPs.module.css";

export interface SourceRow {
  label: string;
  total: number;
  blocked: number;
  flagged: number;
}

/** @deprecated Use SourceRow */
export interface SourceIPRow extends SourceRow {
  ip: string;
  geoCountryCode?: string;
}

export interface OverviewTopListProps {
  rows: SourceRow[];
  iconName: string;
  titleKey: string;
  subtitleKey?: string;
  toggleKey: string;
  showLessKey: string;
  showAll: boolean;
  onToggleShowAll: () => void;
}

const countryDisplayNames = new Intl.DisplayNames(["en"], { type: "region" });

function formatCountryLabel(code: string): string {
  if (!code) return code;
  try {
    const fullName = countryDisplayNames.of(code);
    return fullName != null && fullName !== code
      ? `${fullName} (${code})`
      : code;
  } catch {
    return code;
  }
}

function formatIPLabel(ip: string, countryCode: string): string {
  if (countryCode !== "") {
    return `${ip} (${countryCode})`;
  }
  return ip;
}

const OverviewTopList: React.VFC<OverviewTopListProps> =
  function OverviewTopList(props) {
    const {
      rows,
      iconName,
      titleKey,
      subtitleKey,
      toggleKey,
      showLessKey,
      showAll,
      onToggleShowAll,
    } = props;
    const { themes } = useSystemConfig();
    const { renderToString } = useContext(Context);

    const maxSlots = showAll ? 10 : 5;
    const displayRows = useMemo(() => {
      const list = rows.slice(0, maxSlots);
      while (list.length < maxSlots) {
        list.push({ label: "—", total: 0, blocked: 0, flagged: 0 });
      }
      return list;
    }, [rows, maxSlots]);

    return (
      <div className={styles.topSourceSection}>
        <div className={styles.topSourceIPsHeader}>
          <div className={styles.topSourceIPsHeaderLeft}>
            <div className={styles.topSourceIPsIcon}>
              <Icon iconName={iconName} />
            </div>
            <div className={styles.topSourceIPsTitleGroup}>
              <Text
                as="h3"
                variant="medium"
                block={true}
                className={styles.topSourceIPsTitle}
              >
                {renderToString(titleKey)}
              </Text>
              {subtitleKey != null ? (
                <Text
                  as="p"
                  variant="small"
                  block={true}
                  className={styles.topSourceIPsSubtitle}
                >
                  {renderToString(subtitleKey)}
                </Text>
              ) : null}
            </div>
          </div>
          <div className={styles.headerToggle}>
            <ActionButton
              styles={{
                root: {
                  height: "auto",
                  margin: 0,
                  padding: 0,
                  minWidth: 0,
                  fontSize: 12,
                },
                label: { fontSize: 12, margin: 0 },
                flexContainer: { margin: 0, padding: 0 },
              }}
              theme={themes.actionButton}
              onClick={onToggleShowAll}
              text={renderToString(showAll ? showLessKey : toggleKey)}
            />
          </div>
        </div>

        {/* Column headers */}
        <div className={styles.columnHeaders}>
          <div className={styles.columnHeadersLabel} />
          <div className={styles.columnHeadersRight}>
            <div className={`${styles.columnHeader} ${styles.colBlocked}`}>
              {renderToString(
                "FraudProtectionConfigurationScreen.overview.list.column.blocked"
              )}
            </div>
            <div className={`${styles.columnHeader} ${styles.colFlagged}`}>
              {renderToString(
                "FraudProtectionConfigurationScreen.overview.list.column.flagged"
              )}
            </div>
            <div className={`${styles.columnHeader} ${styles.colTotal}`}>
              {renderToString(
                "FraudProtectionConfigurationScreen.overview.list.column.total"
              )}
            </div>
          </div>
        </div>

        <div className={styles.topSourceIPsList}>
          {displayRows.map((row, index) => {
            const isEmpty = row.label === "—";
            return (
              <div key={index} className={styles.topSourceIPRow}>
                <div className={styles.topSourceIPInfoLeft}>
                  <div className={styles.topSourceIPRank}>#{index + 1}</div>
                  <div className={styles.topSourceIPAddress}>{row.label}</div>
                </div>
                <div className={styles.topSourceIPMetrics}>
                  <div
                    className={`${styles.metricCol} ${styles.colBlocked} ${isEmpty ? styles.metricEmpty : ""}`}
                  >
                    {isEmpty ? "—" : row.blocked}
                  </div>
                  <div
                    className={`${styles.metricCol} ${styles.colFlagged} ${isEmpty ? styles.metricEmpty : ""}`}
                  >
                    {isEmpty ? "—" : row.flagged}
                  </div>
                  <div
                    className={`${styles.metricCol} ${styles.colTotal} ${isEmpty ? styles.metricEmpty : ""}`}
                  >
                    {isEmpty ? "—" : row.total}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

export interface OverviewTopSourceIPsProps {
  sourceIPs: SourceIPRow[];
  showAll: boolean;
  onToggleShowAll: () => void;
}

const OverviewTopSourceIPs: React.VFC<OverviewTopSourceIPsProps> =
  function OverviewTopSourceIPs(props) {
    const { sourceIPs, showAll, onToggleShowAll } = props;
    const rows: SourceRow[] = useMemo(
      () =>
        sourceIPs.map((r) => ({
          ...r,
          label: formatIPLabel(r.ip, r.geoCountryCode ?? ""),
        })),
      [sourceIPs]
    );
    return (
      <OverviewTopList
        rows={rows}
        showAll={showAll}
        onToggleShowAll={onToggleShowAll}
        iconName="ServerEnviroment"
        titleKey="FraudProtectionConfigurationScreen.overview.topSourceIPs.title"
        subtitleKey="FraudProtectionConfigurationScreen.overview.topSourceIPs.subtitle"
        toggleKey="FraudProtectionConfigurationScreen.overview.topSourceIPs.toggle"
        showLessKey="FraudProtectionConfigurationScreen.overview.topSourceIPs.showLess"
      />
    );
  };

export interface OverviewTopSMSOriginsProps {
  smsOrigins: Array<{
    phoneCountryCode: string;
    total: number;
    blocked: number;
    flagged: number;
  }>;
  showAll: boolean;
  onToggleShowAll: () => void;
}

export const OverviewTopSMSOrigins: React.VFC<OverviewTopSMSOriginsProps> =
  function OverviewTopSMSOrigins(props) {
    const { smsOrigins, showAll, onToggleShowAll } = props;
    const rows: SourceRow[] = useMemo(
      () =>
        smsOrigins.map((r) => ({
          label: formatCountryLabel(r.phoneCountryCode),
          total: r.total,
          blocked: r.blocked,
          flagged: r.flagged,
        })),
      [smsOrigins]
    );
    return (
      <OverviewTopList
        rows={rows}
        showAll={showAll}
        onToggleShowAll={onToggleShowAll}
        iconName="CellPhone"
        titleKey="FraudProtectionConfigurationScreen.overview.topSMSOrigins.title"
        subtitleKey="FraudProtectionConfigurationScreen.overview.topSMSOrigins.subtitle"
        toggleKey="FraudProtectionConfigurationScreen.overview.topSMSOrigins.toggle"
        showLessKey="FraudProtectionConfigurationScreen.overview.topSMSOrigins.showLess"
      />
    );
  };

export interface OverviewTopIPLocationsProps {
  ipLocations: Array<{
    geoCountryCode: string;
    total: number;
    blocked: number;
    flagged: number;
  }>;
  showAll: boolean;
  onToggleShowAll: () => void;
}

export const OverviewTopIPLocations: React.VFC<OverviewTopIPLocationsProps> =
  function OverviewTopIPLocations(props) {
    const { ipLocations, showAll, onToggleShowAll } = props;
    const rows: SourceRow[] = useMemo(
      () =>
        ipLocations.map((r) => ({
          label: formatCountryLabel(r.geoCountryCode),
          total: r.total,
          blocked: r.blocked,
          flagged: r.flagged,
        })),
      [ipLocations]
    );
    return (
      <OverviewTopList
        rows={rows}
        showAll={showAll}
        onToggleShowAll={onToggleShowAll}
        iconName="Globe"
        titleKey="FraudProtectionConfigurationScreen.overview.topIPLocations.title"
        subtitleKey="FraudProtectionConfigurationScreen.overview.topIPLocations.subtitle"
        toggleKey="FraudProtectionConfigurationScreen.overview.topIPLocations.toggle"
        showLessKey="FraudProtectionConfigurationScreen.overview.topIPLocations.showLess"
      />
    );
  };

export default OverviewTopSourceIPs;
