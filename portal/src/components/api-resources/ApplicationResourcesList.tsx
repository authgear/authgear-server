import React, { useMemo, useContext, useCallback } from "react";
import cn from "classnames";
import {
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  DetailsListLayoutMode,
} from "@fluentui/react";
import Toggle from "../../Toggle";
import { Context } from "@oursky/react-messageformat";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ApplicationResourcesList.module.css";

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
}

export const ApplicationResourcesList: React.FC<ApplicationResourcesListProps> =
  function ApplicationResourcesList(props) {
    const { className, resources, loading, pagination, onToggleAuthorization } =
      props;
    const { renderToString } = useContext(Context);

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
      () => [
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
      ],
      [renderToString, renderAuthorizedToggle]
    );

    return (
      <div className={cn(className, styles.listRoot)}>
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
      </div>
    );
  };
