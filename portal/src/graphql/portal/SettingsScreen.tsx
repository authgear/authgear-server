import React, { useContext } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";

import { usePivotNavigation } from "../../hook/usePivot";

import styles from "./SettingsScreen.module.scss";

const GENERAL_PIVOT_KEY = "general";
const PORTAL_ADMINS_PIVOT_KEY = "portal_admins";
const SMTP_PIVOT_KEY = "smtp";
const SMS_PROVIDER_PIVOT_KEY = "sms_provider";
const SESSION_PIVOT_KEY = "session";

const SettingsScreen: React.FC = function SettingsScreen() {
  const { renderToString } = useContext(Context);
  const { selectedKey, onLinkClick } = usePivotNavigation([
    GENERAL_PIVOT_KEY,
    PORTAL_ADMINS_PIVOT_KEY,
    SMTP_PIVOT_KEY,
    SMS_PROVIDER_PIVOT_KEY,
    SESSION_PIVOT_KEY,
  ]);

  return (
    <main className={styles.root}>
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="SettingsScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
            <PivotItem
              headerText={renderToString("SettingsScreen.general.title")}
              itemKey={GENERAL_PIVOT_KEY}
            />
            <PivotItem
              headerText={renderToString("SettingsScreen.portal_admins.title")}
              itemKey={PORTAL_ADMINS_PIVOT_KEY}
            />
            <PivotItem
              headerText={renderToString("SettingsScreen.smtp.title")}
              itemKey={SMTP_PIVOT_KEY}
            />
            <PivotItem
              headerText={renderToString("SettingsScreen.sms_provider.title")}
              itemKey={SMS_PROVIDER_PIVOT_KEY}
            />
            <PivotItem
              headerText={renderToString("SettingsScreen.session.title")}
              itemKey={SESSION_PIVOT_KEY}
            />
          </Pivot>
        </div>
      </div>
    </main>
  );
};

export default SettingsScreen;
