import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  ColumnActionsMode,
  DetailsRow,
  IColumn,
  IDetailsRowProps,
  Text,
} from "@fluentui/react";
import { Context as MessageContext } from "../../../intl";
import { useParams } from "react-router-dom";

import styles from "./UserGroupsList.module.css";
import {
  Group,
  Role,
  User,
} from "../../../graphql/adminapi/globalTypes.generated";
import Link from "../../../Link";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList from "./common/RolesAndGroupsBaseList";
import DeleteUserGroupDialog, {
  DeleteUserGroupDialogData,
} from "../dialog/DeleteUserGroupDialog";
import BaseCell from "./common/BaseCell";

export interface UserGroupsListItem extends Pick<Group, "id" | "name" | "key"> {
  roles: {
    totalCount: number;
    items: Pick<Role, "id" | "name" | "key">[] | null;
  };
}

export interface UserGroupsListUser
  extends Pick<User, "id" | "formattedName" | "endUserAccountID"> { }

export enum UserGroupsListColumnKey {
  Name = "Name",
  Key = "Key",
  Role = "Role",
  Action = "Action",
}

interface UserGroupsListProps {
  user: UserGroupsListUser;
  className?: string;
  groups: UserGroupsListItem[];
  isSearch: boolean;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

export const UserGroupsList: React.VFC<UserGroupsListProps> =
  function UserGroupsList({
    user,
    groups,
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
      useState<DeleteUserGroupDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteGroup = useCallback(
      (e: React.MouseEvent<unknown>, item: UserGroupsListItem) => {
        e.preventDefault();
        e.stopPropagation();
        setDeleteDialogData({
          userID: user.id,
          userFormattedName: user.formattedName ?? null,
          userEndUserAccountID: user.endUserAccountID ?? null,
          groupID: item.id,
          groupKey: item.key,
          groupName: item.name ?? null,
        });
      },
      [user]
    );

    const columns: IColumn[] = useMemo((): IColumn[] => {
      return [
        {
          key: UserGroupsListColumnKey.Name,
          fieldName: "name",
          name: renderToString("UserGroupsList.column.name"),
          minWidth: 100,
          maxWidth: 200,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserGroupsListColumnKey.Key,
          fieldName: "key",
          name: renderToString("UserGroupsList.column.key"),
          minWidth: 100,
          maxWidth: 200,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserGroupsListColumnKey.Role,
          fieldName: "role",
          name: renderToString("UserGroupsList.column.role"),
          minWidth: 100,
          maxWidth: 9999,
          isResizable: true,
          columnActionsMode: ColumnActionsMode.disabled,
        },
        {
          key: UserGroupsListColumnKey.Action,
          fieldName: "action",
          name: renderToString("UserGroupsList.column.action"),
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
            to={`/project/${appID}/user-management/groups/${(props.item as UserGroupsListItem).id
              }/details`}
          >
            <DetailsRow {...props} />
          </Link>
        );
      },
      [appID]
    );

    const onRenderItemColumn = useCallback(
      (item: UserGroupsListItem, _index?: number, column?: IColumn) => {
        switch (column?.key) {
          case UserGroupsListColumnKey.Action: {
            return (
              <ActionButtonCell
                variant="destructive"
                text={renderToString("UserGroupsList.actions.remove")}
                onClick={(e) => {
                  onClickDeleteGroup(e, item);
                }}
              />
            );
          }
          case UserGroupsListColumnKey.Role: {
            const text =
              item.roles.totalCount === 0
                ? "-"
                : item.roles.items
                  ?.slice(0, 3)
                  .map((item) => item.name)
                  .join(", ");
            const addtionalInfo =
              item.roles.totalCount > 3 ? ` +${item.roles.totalCount - 3}` : "";
            return (
              <BaseCell>
                <Text className="whitespace-normal line-clamp-4">{`${text}${addtionalInfo}`}</Text>
              </BaseCell>
            );
          }
          default:
            return (
              <TextCell>
                {(item[column?.fieldName as keyof UserGroupsListItem] as React.ReactNode) ?? ""}
              </TextCell>
            );
        }
      },
      [onClickDeleteGroup, renderToString]
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

    const listEmptyText = renderToString("UserGroupsList.empty");

    return (
      <>
        <div className={cn(styles.root, className)}>
          <RolesAndGroupsBaseList
            emptyText={listEmptyText}
            onRenderRow={onRenderRow}
            onRenderItemColumn={onRenderItemColumn}
            items={groups}
            columns={columns}
            pagination={paginationProps}
          />
        </div>
        <DeleteUserGroupDialog
          data={deleteDialogData}
          onDismiss={onDismissDeleteDialog}
        />
      </>
    );
  };
