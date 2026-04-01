import React, { useCallback, useContext, useMemo, useState } from "react";
import { Icon, Text } from "@fluentui/react";
import { Context } from "../../intl";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./OverviewTopSourceIPs.module.css";

export interface SourceIPRow {
  ip: string;
  total: number;
  blocked: number;
  warnings: number;
}

export interface OverviewTopSourceIPsProps {
  sourceIPs: SourceIPRow[];
  maxTotal: number;
}

const OverviewTopSourceIPs: React.VFC<OverviewTopSourceIPsProps> =
  function OverviewTopSourceIPs(props) {
    const { sourceIPs, maxTotal } = props;
    const { themes } = useSystemConfig();
    const { renderToString } = useContext(Context);

    const [showAll, setShowAll] = useState(false);

    const toggleShowAll = useCallback(() => {
      setShowAll((prev) => !prev);
    }, []);

    const maxSlots = showAll ? 10 : 5;
    const displaySourceIPs = useMemo(() => {
      const list = sourceIPs.slice(0, maxSlots);
      while (list.length < maxSlots) {
        list.push({ ip: "—", total: 0, blocked: 0, warnings: 0 });
      }
      return list;
    }, [sourceIPs, maxSlots]);

    return (
      <div className={styles.topSourceSection}>
        <div className={styles.topSourceIPsHeader}>
          <div className={styles.topSourceIPsHeaderLeft}>
            <div className={styles.topSourceIPsIcon}>
              <Icon iconName="ServerEnviroment" />
            </div>
            <Text
              as="h3"
              variant="medium"
              block={true}
              className={styles.topSourceIPsTitle}
            >
              {renderToString(
                "FraudProtectionConfigurationScreen.overview.topSourceIPs.title"
              )}
            </Text>
          </div>
          <ActionButton
            theme={themes.actionButton}
            onClick={toggleShowAll}
            text={renderToString(
              showAll
                ? "FraudProtectionConfigurationScreen.overview.topSourceIPs.showLess"
                : "FraudProtectionConfigurationScreen.overview.topSourceIPs.toggle"
            )}
          />
        </div>
        <div className={styles.topSourceIPsList}>
          {displaySourceIPs.map((row, index) => (
            <div key={index} className={styles.topSourceIPRow}>
              <div className={styles.topSourceIPInfo}>
                <div className={styles.topSourceIPInfoLeft}>
                  <div className={styles.topSourceIPRank}>#{index + 1}</div>
                  <div className={styles.topSourceIPAddress}>{row.ip}</div>
                </div>
                <div className={styles.topSourceIPMetrics}>
                  <div className={styles.totalValue}>
                    {row.total || (row.ip === "—" ? "" : 0)}
                  </div>
                  {row.blocked > 0 ? (
                    <div className={styles.blockedStatus}>
                      {renderToString(
                        "FraudProtectionConfigurationScreen.overview.topSourceIPs.blockedStatus",
                        { count: row.blocked }
                      )}
                    </div>
                  ) : null}
                  {row.blocked === 0 && row.warnings > 0 ? (
                    <div className={styles.challengedStatus}>
                      {renderToString(
                        "FraudProtectionConfigurationScreen.overview.topSourceIPs.warningStatus",
                        { count: row.warnings }
                      )}
                    </div>
                  ) : null}
                </div>
              </div>
              <div className={styles.topSourceIPProgress}>
                <div
                  className={styles.progressBar}
                  style={{
                    width: `${
                      maxTotal > 0 ? (row.total / maxTotal) * 100 : 0
                    }%`,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

export default OverviewTopSourceIPs;
