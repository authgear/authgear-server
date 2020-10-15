import React, { useContext } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";

import { usePivotNavigation } from "../../hook/usePivot";
import PortalAdminsSettings from "./PortalAdminsSettings";

import styles from "./SettingsScreen.module.scss";

const PORTAL_ADMINS_PIVOT_KEY = "portal_admins";

const SettingsScreen: React.FC = function SettingsScreen() {
  const { renderToString } = useContext(Context);
  const { selectedKey, onLinkClick } = usePivotNavigation([
    PORTAL_ADMINS_PIVOT_KEY,
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
              headerText={renderToString("SettingsScreen.portal_admins.title")}
              itemKey={PORTAL_ADMINS_PIVOT_KEY}
            >
              <PortalAdminsSettings />
            </PivotItem>
          </Pivot>
        </div>
      </div>
    </main>
  );
};

export default SettingsScreen;
