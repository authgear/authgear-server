import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import styles from "./ApplicationList.module.css";
import { Toggle } from "../v2/Toggle/Toggle";
import { CardTable } from "../v2/CardTable/CardTable";

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
        <CardTable>
          <CardTable.Header>
            <CardTable.HeaderCell className={styles.colApplication}>
              <FormattedMessage id="ApplicationList.columns.application" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colAuthorized}>
              <FormattedMessage id="ApplicationList.columns.authorized" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colActions} />
          </CardTable.Header>
          {applications.map((item) => {
            const toggleDisabled = disabledToggleClientIDsSet.has(
              item.clientID
            );
            const showManageScopes = item.authorized && !toggleDisabled;
            return (
              <CardTable.Row key={item.clientID}>
                <CardTable.Cell className={styles.colApplication}>
                  <Text size="2" className={styles.applicationName}>
                    {item.name}
                  </Text>
                </CardTable.Cell>
                <CardTable.Cell className={styles.colAuthorized}>
                  <Toggle
                    checked={item.authorized}
                    disabled={toggleDisabled}
                    onCheckedChange={(checked) => {
                      onToggleAuthorized(item, checked);
                    }}
                  />
                </CardTable.Cell>
                <CardTable.Cell className={styles.colActions}>
                  {showManageScopes ? (
                    <button
                      type="button"
                      className={styles.manageScopesButton}
                      onClick={() => onManageScopes(item)}
                    >
                      {renderToString("ApplicationList.columns.manageScopes")}
                    </button>
                  ) : null}
                </CardTable.Cell>
              </CardTable.Row>
            );
          })}
        </CardTable>
      </div>
    );
  };
