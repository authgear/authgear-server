import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { GroupsListFragment } from "./query/groupsListQuery.generated";
import useDelayedValue from "../../hook/useDelayedValue";
import {
  ColumnActionsMode,
  DetailsListLayoutMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
  SelectionMode,
  ShimmeredDetailsList,
  Text,
} from "@fluentui/react";
import styles from "./GroupsList.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import Link from "../../Link";
import ActionButton from "../../ActionButton";
import PaginationWidget from "../../PaginationWidget";

interface GroupsListProps {
  className?: string;
  isSearch: boolean;
  loading: boolean;
  groups: GroupsListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface GroupListItem {
  id: string;
  key: string;
  name: string | null;
  description: string | null;
}

const isGroupListItem = (value: unknown): value is GroupListItem => {
  if (!(value instanceof Object)) {
    return false;
  }
  return "key" in value && "id" in value;
};

const GroupsList: React.VFC<GroupsListProps> = function GroupsList(props) {
  const {
    className,
    loading: rawLoading,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  } = props;
  const edges = props.groups?.edges;
  const loading = useDelayedValue(rawLoading, 500);
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const columns: IColumn[] = [
    {
      key: "name",
      fieldName: "name",
      name: renderToString("GroupsList.column.name"),
      minWidth: 150,
      maxWidth: 260,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "key",
      fieldName: "key",
      name: renderToString("GroupsList.column.key"),
      minWidth: 150,
      maxWidth: 260,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "description",
      fieldName: "description",
      name: renderToString("GroupsList.column.description"),
      minWidth: 300,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "action",
      fieldName: "action",
      name: renderToString("GroupsList.column.action"),
      minWidth: 77,
      maxWidth: 77,
      columnActionsMode: ColumnActionsMode.disabled,
    },
  ];
  const items: GroupListItem[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          items.push({
            id: node.id,
            name: node.name ?? null,
            key: node.key,
            description: node.description ?? null,
          });
        }
      }
    }
    return items;
  }, [edges]);

  const onRenderGroupRow = React.useCallback(
    (props?: IDetailsRowProps) => {
      if (props == null) {
        return null;
      }
      const targetPath = isGroupListItem(props.item)
        ? `/project/${appID}/user-management/groups/${props.item.id}/details`
        : ".";
      return (
        <Link to={targetPath}>
          <DetailsRow {...props} />
        </Link>
      );
    },
    [appID]
  );

  const onRenderGroupItemColumn = useCallback(
    (item: GroupListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "description":
          return (
            <div className={styles.cell}>
              <div className={styles.description}>
                {item[column.key as keyof GroupListItem] ?? ""}
              </div>
            </div>
          );
        case "action": {
          return (
            <div className={styles.cell}>
              <ActionButton
                text={
                  <Text
                    className={styles.actionButtonText}
                    theme={themes.destructive}
                  >
                    <FormattedMessage id="GroupsList.delete-group" />
                  </Text>
                }
                className={styles.actionButton}
                theme={themes.destructive}
              />
            </div>
          );
        }
        default:
          return (
            <div className={styles.cell}>
              <div className={styles.cellText}>
                {item[column?.key as keyof GroupListItem] ?? ""}
              </div>
            </div>
          );
      }
    },
    [themes.destructive]
  );
  return (
    <>
      <div className={cn(styles.root, className)}>
        <div className={styles.listWrapper}>
          <ShimmeredDetailsList
            enableShimmer={loading}
            enableUpdateAnimations={false}
            onRenderRow={onRenderGroupRow}
            onRenderItemColumn={onRenderGroupItemColumn}
            selectionMode={SelectionMode.none}
            layoutMode={DetailsListLayoutMode.justified}
            items={items}
            columns={columns}
          />
        </div>
        {!isSearch ? (
          <PaginationWidget
            className={cn(styles.pagination)}
            offset={offset}
            pageSize={pageSize}
            totalCount={totalCount}
            onChangeOffset={onChangeOffset}
          />
        ) : null}
      </div>
    </>
  );
};

export default GroupsList;
