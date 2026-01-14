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

import styles from "./GroupRolesList.module.css";
import { Group, Role } from "../../../graphql/adminapi/globalTypes.generated";
import Link from "../../../Link";
import DeleteGroupRoleDialog, {
  DeleteGroupRoleDialogData,
} from "../dialog/DeleteGroupRoleDialog";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList from "./common/RolesAndGroupsBaseList";

export interface GroupRolesListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface GroupRolesListGroup
  extends Pick<Role, "id" | "name" | "key"> {}

export enum GroupRolesListColumnKey {
  Name = "Name",
  Key = "Key",
  Action = "Action",
}

interface GroupRolesListProps {
  group: GroupRolesListGroup;
  className?: string;
  roles: GroupRolesListItem[];
}

export const GroupRolesList: React.VFC<GroupRolesListProps> =
  function GroupRolesList({ group, roles, className }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const [deleteDialogData, setDeleteDialogData] =
      useState<DeleteGroupRoleDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteRole = useCallback(
      (e: React.MouseEvent<unknown>, item: GroupRolesListItem) => {
        e.preventDefault();
        e.stopPropagation();
        setDeleteDialogData({
          roleID: item.id,
          roleKey: item.key,
          roleName: item.name ?? null,
          groupID: group.id,
          groupKey: group.key,
          groupName: group.name ?? null,
        });
      },
      [group]
    );

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: GroupRolesListColumnKey.Name,
          fieldName: "name",
          name: renderToString("GroupRolesList.column.name"),
          minWidth: 100,
          maxWidth: 300,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: GroupRolesListColumnKey.Key,
          fieldName: "key",
          name: renderToString("GroupRolesList.column.key"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: GroupRolesListColumnKey.Action,
          fieldName: "action",
          name: renderToString("GroupRolesList.column.action"),
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
            to={`/project/${appID}/user-management/roles/${
              (props.item as GroupRolesListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: GroupRolesListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case GroupRolesListColumnKey.Action: {
            return (
              <ActionButtonCell
                variant="destructive"
                text={renderToString("GroupRolesList.actions.remove")}
                onClick={(e) => {
                  onClickDeleteRole(e, item);
                }}
              />
            );
          }
          default:
            return (
              <TextCell>
                {item[column?.fieldName as keyof GroupRolesListItem] ?? ""}
              </TextCell>
            );
        }
      },
      [onClickDeleteRole, renderToString]
    );

    const listEmptyText = renderToString("GroupRolesList.empty");

    return (
      <>
        <div className={cn(styles.root, className)}>
          <RolesAndGroupsBaseList
            emptyText={listEmptyText}
            onRenderRow={onRenderRow}
            onRenderItemColumn={onRenderItemColumn}
            items={roles}
            columns={columns}
          />
        </div>
        <DeleteGroupRoleDialog
          data={deleteDialogData}
          onDismiss={onDismissDeleteDialog}
        />
      </>
    );
  };
