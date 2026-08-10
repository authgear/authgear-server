import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import styles from "./ApplicationList.module.css";
import { Toggle } from "../v2/Toggle/Toggle";

export interface ApplicationListItem {
  clientID: string;
  name: string;
  authorized: boolean;
}

interface ApplicationListProps {
  className?: string;
  applications: ApplicationListItem[];
  loading: boolean;
  disabledToggleClientIDs: string[];
  onToggleAuthorized: (item: ApplicationListItem, checked: boolean) => void;
  onManageScopes: (item: ApplicationListItem) => void;
}

export const ApplicationList: React.VFC<ApplicationListProps> =
  function ApplicationList(props) {
    const {
      className,
      applications,
      onToggleAuthorized,
      onManageScopes,
      disabledToggleClientIDs,
    } = props;
    const { renderToString } = useContext(MessageContext);

    const disabledToggleClientIDsSet = useMemo(() => {
      return new Set(disabledToggleClientIDs);
    }, [disabledToggleClientIDs]);

    return (
      <div className={cn(className, styles.listRoot)}>
        <div className={styles.tableWrapper}>
          <div className={styles.table}>
            <div className={styles.tableHeader}>
              <div className={styles.tableHeaderCellApplication}>
                <FormattedMessage id="ApplicationList.columns.application" />
              </div>
              <div className={styles.tableHeaderCellAuthorized}>
                <FormattedMessage id="ApplicationList.columns.authorized" />
              </div>
              <div className={styles.tableHeaderCellActions} />
            </div>
            {applications.map((item) => {
              const toggleDisabled = disabledToggleClientIDsSet.has(
                item.clientID
              );
              const showManageScopes = item.authorized && !toggleDisabled;
              return (
                <div key={item.clientID} className={styles.tableRow}>
                  <div className={styles.tableCellApplication}>
                    <Text size="2" className={styles.applicationName}>
                      {item.name}
                    </Text>
                  </div>
                  <div className={styles.tableCellAuthorized}>
                    <Toggle
                      checked={item.authorized}
                      disabled={toggleDisabled}
                      onCheckedChange={(checked) => {
                        onToggleAuthorized(item, checked);
                      }}
                    />
                  </div>
                  <div className={styles.tableCellActions}>
                    {showManageScopes ? (
                      <button
                        type="button"
                        className={styles.manageScopesButton}
                        onClick={() => onManageScopes(item)}
                      >
                        {renderToString("ApplicationList.columns.manageScopes")}
                      </button>
                    ) : null}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    );
  };
