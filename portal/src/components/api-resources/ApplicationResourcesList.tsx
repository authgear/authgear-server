import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ApplicationResourcesList.module.css";
import { useAppContext } from "../../context/AppContext";
import { Toggle } from "../v2/Toggle/Toggle";
import Link from "../../Link";

export interface ApplicationResourceListItem {
  id: string;
  name?: string | null;
  resourceURI: string;
  isAuthorized: boolean;
}

interface ApplicationResourcesListProps {
  className?: string;
  resources: ApplicationResourceListItem[];
  loading: boolean;
  pagination: PaginationProps;
  onToggleAuthorization: (
    item: ApplicationResourceListItem,
    isAuthorized: boolean
  ) => void;
  disabledToggleClientIDs?: string[];
  onManageScopes?: (item: ApplicationResourceListItem) => void;
}

export const ApplicationResourcesList: React.FC<ApplicationResourcesListProps> =
  function ApplicationResourcesList(props) {
    const {
      className,
      resources,
      loading,
      pagination,
      onToggleAuthorization,
      onManageScopes,
      disabledToggleClientIDs,
    } = props;
    const { appNodeID } = useAppContext();
    const { renderToString } = useContext(Context);

    const disabledSet = useMemo(() => {
      return new Set(disabledToggleClientIDs ?? []);
    }, [disabledToggleClientIDs]);

    const isEmpty = !loading && resources.length === 0;

    return (
      <div className={cn(className, styles.listRoot)}>
        {!isEmpty ? (
          <>
            <div className={styles.tableWrapper}>
              <div className={styles.table}>
                <div className={styles.tableHeader}>
                  <div className={styles.tableHeaderCellResources}>
                    <FormattedMessage id="ApplicationResourcesList.columns.resources" />
                  </div>
                  <div className={styles.tableHeaderCellAuthorized}>
                    <FormattedMessage id="ApplicationResourcesList.columns.authorized" />
                  </div>
                  {onManageScopes ? (
                    <div className={styles.tableHeaderCellActions} />
                  ) : null}
                </div>
                {resources.map((item) => {
                  const toggleDisabled = disabledSet.has(item.id);
                  return (
                    <div key={item.id} className={styles.tableRow}>
                      <div className={styles.tableCellResources}>
                        <Text size="2" className={styles.resourceName}>
                          {item.name || item.resourceURI}
                        </Text>
                      </div>
                      <div className={styles.tableCellAuthorized}>
                        <Toggle
                          checked={item.isAuthorized}
                          disabled={toggleDisabled}
                          onCheckedChange={(checked) => {
                            onToggleAuthorization(item, checked);
                          }}
                        />
                      </div>
                      {onManageScopes ? (
                        <div className={styles.tableCellActions}>
                          {item.isAuthorized ? (
                            <button
                              type="button"
                              className={styles.manageScopesButton}
                              onClick={() => onManageScopes(item)}
                            >
                              {renderToString(
                                "ApplicationResourcesList.columns.manageScopes"
                              )}
                            </button>
                          ) : null}
                        </div>
                      ) : null}
                    </div>
                  );
                })}
              </div>
            </div>
            <PaginationWidget className={styles.paginator} {...pagination} />
          </>
        ) : null}

        {isEmpty ? (
          <Text as="p" size="2" color="gray" className={styles.empty}>
            <FormattedMessage
              id="ApplicationResourcesList.empty"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                ReactRouterLink: (chunks: React.ReactNode) => (
                  <Link to={`/project/${appNodeID}/api-resources`}>
                    {chunks}
                  </Link>
                ),
              }}
            />
          </Text>
        ) : null}
      </div>
    );
  };
