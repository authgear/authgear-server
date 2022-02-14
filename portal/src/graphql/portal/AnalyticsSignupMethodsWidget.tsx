import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";

import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import styles from "./AnalyticsSignupMethodsWidget.module.scss";

interface AnalyticsSignupMethodsWidgetProps {
  className?: string;
}

const AnalyticsSignupMethodsWidget: React.FC<AnalyticsSignupMethodsWidgetProps> =
  function AnalyticsSignupMethodsWidget(props) {
    return (
      <Widget className={props.className}>
        <WidgetTitle>
          <FormattedMessage id="AnalyticsSignupMethodsWidget.title" />
        </WidgetTitle>
      </Widget>
    );
  };
export default AnalyticsSignupMethodsWidget;
