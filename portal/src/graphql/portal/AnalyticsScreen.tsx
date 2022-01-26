import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import AnalyticsActivityWidget from "./AnalyticsActivityWidget";
import styles from "./AnalyticsScreen.module.scss";

const AnalyticsScreen: React.FC = function AnalyticsScreen() {
  return (
    <ScreenContent>
      <ScreenTitle>
        <FormattedMessage id="AnalyticsScreen.title" />
      </ScreenTitle>
      <AnalyticsActivityWidget />
    </ScreenContent>
  );
};

export default AnalyticsScreen;
