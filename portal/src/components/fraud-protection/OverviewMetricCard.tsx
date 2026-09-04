import React from "react";
import { Heading, Text } from "@radix-ui/themes";
import styles from "./OverviewMetricCard.module.css";

export type MetricIconVariant = "default" | "success" | "warning" | "blocked";

const iconVariantClass: Record<MetricIconVariant, string> = {
  default: styles.metricIcon,
  success: styles.metricIconSuccess,
  warning: styles.metricIconWarning,
  blocked: styles.metricIconBlocked,
};

export interface OverviewMetricCardProps {
  icon: React.ReactNode;
  iconVariant: MetricIconVariant;
  title: string;
  value: string;
}

const OverviewMetricCard: React.VFC<OverviewMetricCardProps> =
  function OverviewMetricCard(props) {
    const { icon, iconVariant, title, value } = props;

    return (
      <div className={styles.metricCard}>
        <div className={styles.metricCardHeader}>
          <div className={iconVariantClass[iconVariant]}>{icon}</div>
          <div className={styles.metricHeadingGroup}>
            <Heading
              as="h3"
              size="2"
              weight="medium"
              className={styles.metricTitle}
            >
              {title}
            </Heading>
            <Text
              as="div"
              size="6"
              weight="bold"
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
