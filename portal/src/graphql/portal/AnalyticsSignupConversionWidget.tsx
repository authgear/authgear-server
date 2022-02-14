import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";

import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import styles from "./AnalyticsSignupConversionWidget.module.scss";

interface AnalyticsSignupConversionWidgetProps {
  className?: string;
}

const AnalyticsSignupConversionWidget: React.FC<AnalyticsSignupConversionWidgetProps> =
  function AnalyticsSignupConversionWidget(props) {
    return (
      <Widget className={props.className}>
        <WidgetTitle>
          <FormattedMessage id="AnalyticsSignupConversionWidget.title" />
        </WidgetTitle>
      </Widget>
    );
  };
export default AnalyticsSignupConversionWidget;
