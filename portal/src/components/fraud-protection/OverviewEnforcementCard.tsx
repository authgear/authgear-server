import React, { useContext } from "react";
import { Icon, Text } from "@fluentui/react";
import { Context } from "../../intl";
import Tooltip from "../../Tooltip";
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
            <Icon iconName="Shield" />
          </div>
          <div className={styles.metricHeadingGroup}>
            <div className={styles.metricTitleRow}>
              <Text
                as="h3"
                variant="medium"
                block={true}
                className={styles.metricTitle}
              >
                {title}
              </Text>
              <Tooltip
                tooltipMessageId="FraudProtectionConfigurationScreen.enforcement.tooltip"
                className={styles.metricInfoTooltip}
              >
                <Icon iconName="Info" className={styles.metricInfoIcon} />
              </Tooltip>
            </div>
            {description != null ? (
              <Text
                as="div"
                variant="medium"
                block={true}
                className={styles.metricDescription}
              >
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
