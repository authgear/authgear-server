import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import WidgetTitle from "../../WidgetTitle";
import { FraudProtectionDecisionAction } from "../../types";
import { useFraudProtectionOverviewQueryQuery } from "../../graphql/adminapi/query/fraudProtectionOverviewQuery.generated";
import OverviewMetricCard from "./OverviewMetricCard";
import OverviewEnforcementCard from "./OverviewEnforcementCard";
import OverviewTopSourceIPs, { SourceIPRow } from "./OverviewTopSourceIPs";
import styles from "./FraudProtectionOverviewTab.module.css";

type OverviewTimeRange = "24h" | "7d";

export interface FraudProtectionOverviewTabProps {
  enabled: boolean;
  enforcementMode: FraudProtectionDecisionAction;
  onChangeToSettings: () => void;
}

const FraudProtectionOverviewTab: React.VFC<FraudProtectionOverviewTabProps> =
  function FraudProtectionOverviewTab(props) {
    const { enabled, enforcementMode, onChangeToSettings } = props;
    const { renderToString } = useContext(Context);
    const isObserveMode = enforcementMode === "record_only";

    const [overviewTimeRange, setOverviewTimeRange] =
      useState<OverviewTimeRange>("24h");

    const timeRangeVars = useMemo(() => {
      const now = new Date();
      if (overviewTimeRange === "24h") {
        return {
          rangeFrom: new Date(
            now.getTime() - 24 * 60 * 60 * 1000
          ).toISOString(),
          rangeTo: now.toISOString(),
        };
      }
      return {
        rangeFrom: new Date(
          now.getTime() - 7 * 24 * 60 * 60 * 1000
        ).toISOString(),
        rangeTo: now.toISOString(),
      };
    }, [overviewTimeRange]);

    const {
      data: overviewData,
      loading: overviewIsLoading,
      error: overviewError,
      refetch: refetchOverview,
    } = useFraudProtectionOverviewQueryQuery({
      skip: !enabled,
      variables: {
        rangeFrom: timeRangeVars.rangeFrom,
        rangeTo: timeRangeVars.rangeTo,
      },
    });
    const overview = overviewData?.fraudProtectionOverview ?? null;
    const overviewHasError = overviewError != null;
    const onRetryOverview = useCallback(() => {
      void refetchOverview();
    }, [refetchOverview]);

    const formatCount = useCallback((value: number | undefined): string => {
      return value == null ? "—" : String(value);
    }, []);

    const enforcementTitle = renderToString(
      enabled
        ? "FraudProtectionConfigurationScreen.overview.enforcement.enabled.title"
        : "FraudProtectionConfigurationScreen.overview.enforcement.disabled.title"
    );
    const enforcementDescription = renderToString(
      isObserveMode
        ? "FraudProtectionConfigurationScreen.overview.enforcement.observe.description"
        : "FraudProtectionConfigurationScreen.overview.enforcement.protect.description"
    );

    const sourceIPs = useMemo<SourceIPRow[]>(() => {
      return (
        overview?.topSourceIPs.map((sourceIP) => ({
          ip: sourceIP.ipAddress,
          total: sourceIP.totalActions,
          blocked: sourceIP.blockedActions,
          flagged: sourceIP.warnedActions,
        })) ?? []
      );
    }, [overview]);

    const maxTotal = useMemo(() => {
      if (sourceIPs.length === 0) {
        return 0;
      }
      return Math.max(...sourceIPs.map((row) => row.total));
    }, [sourceIPs]);

    const overviewTimeRangeOptions = useMemo(
      () => [
        {
          key: "24h" as const,
          label: renderToString(
            "FraudProtectionConfigurationScreen.overview.timeRange.last24Hours"
          ),
        },
        {
          key: "7d" as const,
          label: renderToString(
            "FraudProtectionConfigurationScreen.overview.timeRange.last7Days"
          ),
        },
      ],
      [renderToString]
    );

    return (
      <section className={styles.section}>
        {overviewHasError ? (
          <ShowError error={overviewError} onRetry={onRetryOverview} />
        ) : null}
        <div className={styles.overviewHeaderRow}>
          <WidgetTitle>
            <FormattedMessage id="FraudProtectionConfigurationScreen.tab.overview.title" />
          </WidgetTitle>
          <div className={styles.overviewTimeRange}>
            <span>
              <FormattedMessage id="FraudProtectionConfigurationScreen.overview.timeRange.label" />
            </span>
            <div className={styles.overviewTimeRangeGroup}>
              {overviewTimeRangeOptions.map((option) => (
                <button
                  key={option.key}
                  type="button"
                  className={`${styles.overviewTimeRangeButton} ${
                    overviewTimeRange === option.key
                      ? styles.overviewTimeRangeButtonSelected
                      : ""
                  }`}
                  onClick={() => setOverviewTimeRange(option.key)}
                >
                  {option.label}
                </button>
              ))}
            </div>
          </div>
        </div>
        {overviewIsLoading ? (
          <ShowLoading />
        ) : (
          <div className={styles.overviewLayout}>
            <div className={styles.overviewMain}>
              <div className={styles.overviewMainRow1}>
                <OverviewEnforcementCard
                  title={enforcementTitle}
                  description={enforcementDescription}
                  onChangeToSettings={onChangeToSettings}
                />
                <OverviewMetricCard
                  iconName="Message"
                  iconVariant="default"
                  title={renderToString(
                    "FraudProtectionConfigurationScreen.overview.total.title"
                  )}
                  value={formatCount(overview?.totalActions)}
                />
              </div>
              <OverviewMetricCard
                iconName="Warning"
                iconVariant="warning"
                title={renderToString(
                  "FraudProtectionConfigurationScreen.overview.flagged.title"
                )}
                value={formatCount(overview?.warnedActions)}
              />
              <OverviewMetricCard
                iconName="BlockContact"
                iconVariant="blocked"
                title={renderToString(
                  "FraudProtectionConfigurationScreen.overview.blocked.title"
                )}
                value={formatCount(overview?.blockedActions)}
              />
            </div>
            <div className={styles.overviewSide}>
              <OverviewTopSourceIPs sourceIPs={sourceIPs} maxTotal={maxTotal} />
            </div>
          </div>
        )}
      </section>
    );
  };

export default FraudProtectionOverviewTab;
