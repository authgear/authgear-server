import React, { useContext } from "react";
import { Heading, Text } from "@radix-ui/themes";
import { LockClosedIcon, InfoCircledIcon } from "@radix-ui/react-icons";
import { Context } from "../../intl";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import styles from "./OverviewMetricCard.module.css";

export interface OverviewEnforcementCardProps {
  title: string;
  description?: string;
  onChangeToSettings: () => void;
}

const OverviewEnforcementCard: React.VFC<OverviewEnforcementCardProps> =
  function OverviewEnforcementCard(props) {
    const { title, description, onChangeToSettings } = props;
    const { renderToString } = useContext(Context);

    return (
      <div className={styles.metricCardPrimary}>
        <div className={styles.metricCardHeader}>
          <div className={styles.metricIcon}>
            <LockClosedIcon />
          </div>
          <div className={styles.metricHeadingGroup}>
            <div className={styles.metricTitleRow}>
              <Heading
                as="h3"
                size="2"
                weight="medium"
                className={styles.metricTitle}
              >
                {title}
              </Heading>
              <Tooltip
                content={renderToString(
                  "FraudProtectionConfigurationScreen.enforcement.tooltip"
                )}
              >
                <InfoCircledIcon className={styles.metricInfoIcon} />
              </Tooltip>
            </div>
            {description != null ? (
              <Text as="div" size="2" className={styles.metricDescription}>
                {description}
              </Text>
            ) : null}
            <button
              type="button"
              className={styles.metricLink}
              onClick={onChangeToSettings}
            >
              {renderToString(
                "FraudProtectionConfigurationScreen.overview.enforcement.changeMode"
              )}
            </button>
          </div>
        </div>
      </div>
    );
  };

export default OverviewEnforcementCard;
