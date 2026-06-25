import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import WidgetTitle from "../../WidgetTitle";
import { FraudProtectionDecisionAction } from "../../types";
import { useFraudProtectionOverviewQueryQuery } from "../../graphql/adminapi/query/fraudProtectionOverviewQuery.generated";
import OverviewMetricCard from "./OverviewMetricCard";
import OverviewEnforcementCard from "./OverviewEnforcementCard";
import OverviewTopSourceIPs, {
  OverviewTopIPLocations,
  OverviewTopSMSOrigins,
  SourceIPRow,
} from "./OverviewTopSourceIPs";
import OverviewRequestsChart from "./OverviewRequestsChart";
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
    const [showAllTopLists, setShowAllTopLists] = useState(false);

    const toggleShowAllTopLists = useCallback(() => {
      setShowAllTopLists((prev) => !prev);
    }, []);

    const timeRangeVars = useMemo(() => {
      const now = new Date();
      if (overviewTimeRange === "24h") {
        const startOfToday = new Date(now);
        startOfToday.setHours(0, 0, 0, 0);
        const endOfToday = new Date(now);
        endOfToday.setHours(23, 59, 59, 999);
        return {
          rangeFrom: startOfToday.toISOString(),
          rangeTo: endOfToday.toISOString(),
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
    const overviewHasError = overviewError != null;
    const onRetryOverview = useCallback(() => {
      void refetchOverview();
    }, [refetchOverview]);

    const sendSMS = overviewData?.fraudProtectionOverview?.sendSMS;

    const sourceIPs = useMemo<SourceIPRow[]>(() => {
      return (sendSMS?.topSourceIPs ?? []).map((ip) => ({
        ip: ip.ipAddress,
        geoCountryCode: ip.geoCountryCode,
        label: ip.ipAddress,
        total: ip.total,
        blocked: ip.blocked,
        flagged: ip.flagged,
      }));
    }, [sendSMS?.topSourceIPs]);

    const smsOrigins = useMemo(() => {
      return sendSMS?.topSMSOrigins ?? [];
    }, [sendSMS?.topSMSOrigins]);

    const ipLocations = useMemo(() => {
      return sendSMS?.topIPLocations ?? [];
    }, [sendSMS?.topIPLocations]);

    const timeBuckets = useMemo(() => {
      return sendSMS?.timeBuckets ?? [];
    }, [sendSMS?.timeBuckets]);

    const displayMetrics = useMemo(() => {
      return {
        total: sendSMS?.total ?? 0,
        blocked: sendSMS?.blocked ?? 0,
        flagged: sendSMS?.flagged ?? 0,
      };
    }, [sendSMS?.blocked, sendSMS?.flagged, sendSMS?.total]);

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
            <div className={styles.overviewCardsRow}>
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
                value={formatCount(displayMetrics.total)}
              />
              <OverviewMetricCard
                iconName="Warning"
                iconVariant="warning"
                title={renderToString(
                  "FraudProtectionConfigurationScreen.overview.flagged.title"
                )}
                value={formatCount(displayMetrics.flagged)}
              />
              <OverviewMetricCard
                iconName="BlockContact"
                iconVariant="blocked"
                title={renderToString(
                  "FraudProtectionConfigurationScreen.overview.blocked.title"
                )}
                value={formatCount(displayMetrics.blocked)}
              />
            </div>
            <OverviewRequestsChart
              timeBuckets={timeBuckets}
              timeRange={overviewTimeRange}
              rangeFrom={timeRangeVars.rangeFrom}
              rangeTo={timeRangeVars.rangeTo}
            />
            <div className={styles.overviewBottomRow}>
              <OverviewTopSMSOrigins
                smsOrigins={smsOrigins}
                showAll={showAllTopLists}
                onToggleShowAll={toggleShowAllTopLists}
              />
              <OverviewTopIPLocations
                ipLocations={ipLocations}
                showAll={showAllTopLists}
                onToggleShowAll={toggleShowAllTopLists}
              />
              <OverviewTopSourceIPs
                sourceIPs={sourceIPs}
                showAll={showAllTopLists}
                onToggleShowAll={toggleShowAllTopLists}
              />
            </div>
          </div>
        )}
      </section>
    );
  };

export default FraudProtectionOverviewTab;
