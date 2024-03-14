import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { RolesListFragment } from "./query/rolesListQuery.generated";
import useDelayedValue from "../../hook/useDelayedValue";
import {
  ColumnActionsMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
} from "@fluentui/react";
import styles from "./RolesList.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useParams } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import Link from "../../Link";
import DeleteRoleDialog, { DeleteRoleDialogData } from "./DeleteRoleDialog";
import RolesAndGroupsBaseList from "../../components/roles-and-groups/list/RolesAndGroupsBaseList";
import ActionButtonCell from "../../components/roles-and-groups/list/ActionButtonCell";
import TextCell from "../../components/roles-and-groups/list/TextCell";
import DescriptionCell from "../../components/roles-and-groups/list/DescriptionCell";

interface RolesListProps {
  className?: string;
  isSearch: boolean;
  loading: boolean;
  roles: RolesListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface RoleListItem {
  id: string;
  key: string;
  name: string | null;
  description: string | null;
}

const isRoleListItem = (value: unknown): value is RoleListItem => {
  if (!(value instanceof Object)) {
    return false;
  }
  return "key" in value && "id" in value;
};

const RolesList: React.VFC<RolesListProps> = function RolesList(props) {
  const {
    className,
    loading: rawLoading,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  } = props;
  const edges = props.roles?.edges;
  const loading = useDelayedValue(rawLoading, 500);
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const columns: IColumn[] = [
    {
      key: "name",
      fieldName: "name",
      name: renderToString("RolesList.column.name"),
      minWidth: 150,
      maxWidth: 260,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "key",
      fieldName: "key",
      name: renderToString("RolesList.column.key"),
      minWidth: 150,
      maxWidth: 260,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "description",
      fieldName: "description",
      name: renderToString("RolesList.column.description"),
      minWidth: 300,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "action",
      fieldName: "action",
      name: renderToString("RolesList.column.action"),
      minWidth: 67,
      maxWidth: 67,
      columnActionsMode: ColumnActionsMode.disabled,
    },
  ];
  const items: RoleListItem[] = useMemo(() => {
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

  const onRenderRoleRow = React.useCallback(
    (props?: IDetailsRowProps) => {
      if (props == null) {
        return null;
      }
      const targetPath = isRoleListItem(props.item)
        ? `/project/${appID}/user-management/roles/${props.item.id}/details`
        : ".";
      return (
        <Link to={targetPath} className="contents">
          <DetailsRow {...props} />
        </Link>
      );
    },
    [appID]
  );
  const [deleteRoleDialogData, setDeleteRoleDialogData] =
    useState<DeleteRoleDialogData | null>(null);
  const onClickDeleteRole = useCallback(
    (e: React.MouseEvent<unknown>, item: RoleListItem) => {
      e.preventDefault();
      e.stopPropagation();
      setDeleteRoleDialogData({
        roleID: item.id,
        roleName: item.name,
        roleKey: item.key,
      });
    },
    []
  );
  const dismissDeleteRoleDialog = useCallback(() => {
    setDeleteRoleDialogData(null);
  }, []);

  const onRenderRoleItemColumn = useCallback(
    (item: RoleListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "description":
          return (
            <DescriptionCell>
              {item[column.key as keyof RoleListItem] ?? ""}
            </DescriptionCell>
          );
        case "action": {
          return (
            <ActionButtonCell
              text={renderToString("RolesList.delete-role")}
              onClick={(e) => {
                onClickDeleteRole(e, item);
              }}
            />
          );
        }
        default:
          return (
            <TextCell>{item[column?.key as keyof RoleListItem] ?? ""}</TextCell>
          );
      }
    },
    [themes.destructive, onClickDeleteRole]
  );
  return (
    <>
      <div className={cn(styles.root, className)}>
        <RolesAndGroupsBaseList
          loading={loading}
          onRenderRow={onRenderRoleRow}
          onRenderItemColumn={onRenderRoleItemColumn}
          items={items}
          columns={columns}
          isSearch={isSearch}
          offset={offset}
          pageSize={pageSize}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
        <DeleteRoleDialog
          onDismiss={dismissDeleteRoleDialog}
          data={deleteRoleDialogData}
        />
      </div>
    </>
  );
};

export default RolesList;
