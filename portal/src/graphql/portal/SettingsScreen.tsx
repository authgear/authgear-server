import React, { useContext } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";

import { usePivotNavigation } from "../../hook/usePivot";
import GeneralSettings from "./GeneralSettings";
import PortalAdminsSettings from "./PortalAdminsSettings";
import SessionSettings from "./SessionSettings";
import { ModifiedIndicatorWrapper } from "../../ModifiedIndicatorPortal";

import styles from "./SettingsScreen.module.scss";
import HooksSettings from "./HooksSettings";

const GENERAL_PIVOT_KEY = "general";
const PORTAL_ADMINS_PIVOT_KEY = "portal_admins";
const SESSION_PIVOT_KEY = "session";
const HOOK_PIVOT_KEY = "hooks";

const SettingsScreen: React.FC = function SettingsScreen() {
  const { renderToString } = useContext(Context);
  const { selectedKey, onLinkClick } = usePivotNavigation([
    GENERAL_PIVOT_KEY,
    PORTAL_ADMINS_PIVOT_KEY,
    SESSION_PIVOT_KEY,
    HOOK_PIVOT_KEY,
  ]);

  return (
    <main className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="SettingsScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
            <PivotItem
              headerText={renderToString("SettingsScreen.general.title")}
              itemKey={GENERAL_PIVOT_KEY}
            >
              <GeneralSettings />
            </PivotItem>
            <PivotItem
              headerText={renderToString("SettingsScreen.portal_admins.title")}
              itemKey={PORTAL_ADMINS_PIVOT_KEY}
            >
              <PortalAdminsSettings />
            </PivotItem>
            <PivotItem
              headerText={renderToString("SettingsScreen.session.title")}
              itemKey={SESSION_PIVOT_KEY}
            >
              <SessionSettings />
            </PivotItem>
            <PivotItem
              headerText={renderToString("SettingsScreen.hooks.title")}
              itemKey={HOOK_PIVOT_KEY}
            >
              <HooksSettings />
            </PivotItem>
          </Pivot>
        </div>
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default SettingsScreen;
