import { FormattedMessage } from "@oursky/react-messageformat";
import React from "react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import styles from "./BotProtectionConfigurationScreen.module.css";

export interface BotProtectionConfigurationContentProps { }

const BotProtectionConfigurationContent: React.VFC<BotProtectionConfigurationContentProps> =
  function BotProtectionConfigurationContent() {
    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="BotProtectionConfigurationScreen.title" />
        </ScreenTitle>
        <div className={styles.widget}>
          <p>dummy screen</p>
        </div>
      </ScreenContent>
    );
  };

const BotProtectionConfigurationScreen: React.VFC =
  function BotProtectionConfigurationScreen() {
    return <BotProtectionConfigurationContent />;
  };

export default BotProtectionConfigurationScreen;
