import React from "react";
import { Icon, Text } from "@fluentui/react";
import styles from "./OverviewMetricCard.module.css";

export type MetricIconVariant = "default" | "success" | "warning" | "blocked";

const iconVariantClass: Record<MetricIconVariant, string> = {
  default: styles.metricIcon,
  success: styles.metricIconSuccess,
  warning: styles.metricIconWarning,
  blocked: styles.metricIconBlocked,
};

export interface OverviewMetricCardProps {
  iconName: string;
  iconVariant: MetricIconVariant;
  title: string;
  value: string;
}

const OverviewMetricCard: React.VFC<OverviewMetricCardProps> =
  function OverviewMetricCard(props) {
    const { iconName, iconVariant, title, value } = props;

    return (
      <div className={styles.metricCard}>
        <div className={styles.metricCardHeader}>
          <div className={iconVariantClass[iconVariant]}>
            <Icon iconName={iconName} />
          </div>
          <div className={styles.metricHeadingGroup}>
            <Text
              as="h3"
              variant="medium"
              block={true}
              className={styles.metricTitle}
            >
              {title}
            </Text>
            <Text
              as="div"
              variant="xLargePlus"
              block={true}
              className={styles.metricValue}
            >
              {value}
            </Text>
          </div>
        </div>
      </div>
    );
  };

export default OverviewMetricCard;
