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

import styles from "./UserRolesList.module.css";
import {
  Group,
  Role,
  User,
} from "../../../graphql/adminapi/globalTypes.generated";
import Link from "../../../Link";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList from "./common/RolesAndGroupsBaseList";
import DeleteUserRoleDialog, {
  DeleteUserRoleDialogData,
} from "../dialog/DeleteUserRoleDialog";

export interface UserRolesListItem extends Pick<Role, "id" | "name" | "key"> {
  groups: Pick<Group, "id" | "name" | "key">[];
}

export interface UserRolesListUser
  extends Pick<User, "id" | "formattedName" | "endUserAccountID"> {}

export enum UserRolesListColumnKey {
  Name = "Name",
  Key = "Key",
  Group = "Group",
  Action = "Action",
}

interface UserRolesListProps {
  user: UserRolesListUser;
  className?: string;
  roles: UserRolesListItem[];
  isSearch: boolean;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

export const UserRolesList: React.VFC<UserRolesListProps> =
  function UserRolesList({
    user,
    roles,
    className,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(MessageContext);

    const [deleteDialogData, setDeleteDialogData] =
      useState<DeleteUserRoleDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteRole = useCallback(
      (e: React.MouseEvent<unknown>, item: UserRolesListItem) => {
        e.preventDefault();
        e.stopPropagation();
        setDeleteDialogData({
          userID: user.id,
          userFormattedName: user.formattedName ?? null,
          userEndUserAccountID: user.endUserAccountID ?? null,
          roleID: item.id,
          roleKey: item.key,
          roleName: item.name ?? null,
        });
      },
      [user]
    );

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: UserRolesListColumnKey.Name,
          fieldName: "name",
          name: renderToString("UserRolesList.column.name"),
          minWidth: 100,
          maxWidth: 200,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserRolesListColumnKey.Key,
          fieldName: "key",
          name: renderToString("UserRolesList.column.key"),
          minWidth: 100,
          maxWidth: 200,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserRolesListColumnKey.Group,
          fieldName: "group",
          name: renderToString("UserRolesList.column.group"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserRolesListColumnKey.Action,
          fieldName: "action",
          name: renderToString("UserRolesList.column.action"),
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
              (props.item as UserRolesListItem).id
            }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: UserRolesListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case UserRolesListColumnKey.Action: {
            return (
              <ActionButtonCell
                variant="destructive"
                text={renderToString("UserRolesList.actions.remove")}
                disabled={item.groups.length !== 0}
                onClick={(e) => {
                  onClickDeleteRole(e, item);
                }}
              />
            );
          }
          case UserRolesListColumnKey.Group:
            return (
              <TextCell>
                {item.groups.length === 0
                  ? "-"
                  : item.groups.map((group) => group.name).join(", ")}
              </TextCell>
            );
          default:
            return (
              <TextCell>
                {(item[
                  column?.fieldName as keyof UserRolesListItem
                ] as React.ReactNode) ?? ""}
              </TextCell>
            );
        }
      },
      [onClickDeleteRole, renderToString]
    );

    const paginationProps = useMemo(
      () => ({
        isSearch,
        offset,
        pageSize,
        totalCount,
        onChangeOffset,
      }),
      [isSearch, offset, pageSize, totalCount, onChangeOffset]
    );

    const listEmptyText = renderToString("UserRolesList.empty");

    return (
      <>
        <div className={cn(styles.root, className)}>
          <RolesAndGroupsBaseList
            emptyText={listEmptyText}
            onRenderRow={onRenderRow}
            onRenderItemColumn={onRenderItemColumn}
            items={roles}
            columns={columns}
            pagination={paginationProps}
          />
          <DeleteUserRoleDialog
            data={deleteDialogData}
            onDismiss={onDismissDeleteDialog}
          />
        </div>
      </>
    );
  };
