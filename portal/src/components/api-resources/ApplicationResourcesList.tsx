import React, { useMemo, useContext, useCallback } from "react";
import cn from "classnames";
import {
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  DetailsListLayoutMode,
  Text,
} from "@fluentui/react";
import Toggle from "../../Toggle";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ApplicationResourcesList.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ActionButton from "../../ActionButton";
import { useAppContext } from "../../context/AppContext";

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
    } = props;
    const { appNodeID } = useAppContext();
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const renderAuthorizedToggle = useCallback(
      (item: ApplicationResourceListItem) => {
        return (
          <Toggle
            checked={item.isAuthorized}
            onChange={(_: unknown, checked: boolean | undefined) => {
              onToggleAuthorization(item, checked ?? false);
            }}
            disabled={props.disabledToggleClientIDs?.includes(item.id)}
          />
        );
      },
      [onToggleAuthorization, props.disabledToggleClientIDs]
    );

    const columns: IColumn[] = useMemo(
      () =>
        [
          {
            key: "resources",
            name: renderToString("ApplicationResourcesList.columns.resources"),
            minWidth: 200,
            maxWidth: 400,
            isResizable: true,
            onRender: (item: ApplicationResourceListItem) => {
              return item.name || item.resourceURI;
            },
          },
          {
            key: "authorized",
            name: renderToString("ApplicationResourcesList.columns.authorized"),
            minWidth: 150,
            isResizable: true,
            onRender: renderAuthorizedToggle,
          },
          onManageScopes
            ? {
                key: "actions",
                name: "",
                minWidth: 100,
                maxWidth: 100,
                isResizable: false,
                // eslint-disable-next-line react/no-unstable-nested-components
                onRender: (item: ApplicationResourceListItem) => {
                  if (!item.isAuthorized) {
                    return null;
                  }
                  const handleClick = () => {
                    onManageScopes(item);
                  };
                  return (
                    <ActionButton
                      text={renderToString(
                        "ApplicationResourcesList.columns.manageScopes"
                      )}
                      styles={{
                        label: { fontWeight: 600 },
                        root: { height: "auto" },
                      }}
                      theme={themes.actionButton}
                      onClick={handleClick}
                    />
                  );
                },
              }
            : null,
        ].filter((it) => !!it),
      [
        renderToString,
        renderAuthorizedToggle,
        themes.actionButton,
        onManageScopes,
      ]
    );

    const isEmpty = !loading && resources.length === 0;

    return (
      <div className={cn(className, styles.listRoot)}>
        {!isEmpty ? (
          <>
            <div data-is-scrollable="true" className={styles.listWrapper}>
              <ShimmeredDetailsList
                items={resources}
                enableShimmer={loading}
                columns={columns}
                layoutMode={DetailsListLayoutMode.justified}
                selectionMode={SelectionMode.none}
              />
            </div>
            <PaginationWidget className={styles.paginator} {...pagination} />
          </>
        ) : null}

        {isEmpty ? (
          <Text
            styles={{ root: { color: themes.main.palette.neutralTertiary } }}
          >
            <FormattedMessage
              id="ApplicationResourcesList.empty"
              values={{
                to: `/project/${appNodeID}/api-resources`,
              }}
            />
          </Text>
        ) : null}
      </div>
    );
  };
