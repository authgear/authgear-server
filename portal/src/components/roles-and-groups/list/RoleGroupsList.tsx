import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  ColumnActionsMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
} from "@fluentui/react";
import { Context as MessageContext } from "../../../intl";
import { useParams } from "react-router-dom";

import styles from "./RoleGroupsList.module.css";
import { Group, Role } from "../../../graphql/adminapi/globalTypes.generated";
import Link from "../../../Link";
import DeleteRoleGroupDialog, {
  DeleteRoleGroupDialogData,
} from "../dialog/DeleteRoleGroupDialog";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList from "./common/RolesAndGroupsBaseList";

export interface RoleGroupsListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface RoleGroupsListRole extends Pick<Role, "id" | "name" | "key"> {}

export enum RoleGroupsListColumnKey {
  Name = "Name",
  Key = "Key",
  Action = "Action",
}

interface RoleGroupsListProps {
  role: RoleGroupsListRole;
  className?: string;
  groups: RoleGroupsListItem[];
}

export const RoleGroupsList: React.VFC<RoleGroupsListProps> =
  function RoleGroupsList({ role, groups, className }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const [deleteDialogData, setDeleteDialogData] =
      useState<DeleteRoleGroupDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteGroup = useCallback(
      (e: React.MouseEvent<unknown>, item: RoleGroupsListItem) => {
        e.preventDefault();
        e.stopPropagation();
        setDeleteDialogData({
          roleID: role.id,
          roleKey: role.key,
          roleName: role.name ?? null,
          groupID: item.id,
          groupKey: item.key,
          groupName: item.name ?? null,
        });
      },
      [role]
    );

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: RoleGroupsListColumnKey.Name,
          fieldName: "name",
          name: renderToString("RoleGroupsList.column.name"),
          minWidth: 100,
          maxWidth: 300,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: RoleGroupsListColumnKey.Key,
          fieldName: "key",
          name: renderToString("RoleGroupsList.column.key"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: RoleGroupsListColumnKey.Action,
          fieldName: "action",
          name: renderToString("RoleGroupsList.column.action"),
          minWidth: 67,
          maxWidth: 67,
          columnActionsMode: ColumnActionsMode.disabled,
        },
      ];
    }, [renderToString]);

    const onRenderRow = React.useCallback(
      (props?: IDetailsRowProps) => {
        if (props == null) {
          return null;
        }
        return (
          <Link
            className="contents"
            to={`/project/${appID}/user-management/groups/${
              (props.item as RoleGroupsListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: RoleGroupsListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case RoleGroupsListColumnKey.Action: {
            return (
              <ActionButtonCell
                variant="destructive"
                text={renderToString("RoleGroupsList.actions.remove")}
                onClick={(e) => {
                  onClickDeleteGroup(e, item);
                }}
              />
            );
          }
          default:
            return (
              <TextCell>
                {item[column?.fieldName as keyof RoleGroupsListItem] ?? ""}
              </TextCell>
            );
        }
      },
      [onClickDeleteGroup, renderToString]
    );

    const listEmptyText = renderToString("RoleGroupsList.empty");

    return (
      <>
        <div className={cn(styles.root, className)}>
          <RolesAndGroupsBaseList
            emptyText={listEmptyText}
            onRenderRow={onRenderRow}
            onRenderItemColumn={onRenderItemColumn}
            items={groups}
            columns={columns}
          />
        </div>
        <DeleteRoleGroupDialog
          data={deleteDialogData}
          onDismiss={onDismissDeleteDialog}
        />
      </>
    );
  };
