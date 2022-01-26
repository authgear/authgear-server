import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import styles from "./AnalyticsActivityWidget.module.scss";

const AnalyticsActivityWidget: React.FC = function AnalyticsActivityWidget() {
  return (
    <Widget className={styles.widget}>
      <WidgetTitle>
        <FormattedMessage id="AnalyticsActivityWidget.title" />
      </WidgetTitle>
    </Widget>
  );
};

export default AnalyticsActivityWidget;
