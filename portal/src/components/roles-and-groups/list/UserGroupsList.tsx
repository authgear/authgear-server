import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { Context as MessageContext } from "../../../intl";
import { useNavigate, useParams } from "react-router-dom";

import styles from "./UserGroupsList.module.css";
import {
  Group,
  Role,
  User,
} from "../../../graphql/adminapi/globalTypes.generated";
import ActionButtonCell from "./common/ActionButtonCell";
import TextCell from "./common/TextCell";
import RolesAndGroupsBaseList, {
  RolesAndGroupsListColumn,
} from "./common/RolesAndGroupsBaseList";
import DeleteUserGroupDialog, {
  DeleteUserGroupDialogData,
} from "../dialog/DeleteUserGroupDialog";
import BaseCell from "./common/BaseCell";
import { TrashIcon } from "@radix-ui/react-icons";

export interface UserGroupsListItem extends Pick<Group, "id" | "name" | "key"> {
  roles: {
    totalCount: number;
    items: Pick<Role, "id" | "name" | "key">[] | null;
  };
}

export interface UserGroupsListUser
  extends Pick<User, "id" | "formattedName" | "endUserAccountID"> {}

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
    const navigate = useNavigate();
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

    const columns: RolesAndGroupsListColumn[] =
      useMemo((): RolesAndGroupsListColumn[] => {
        return [
          {
            key: UserGroupsListColumnKey.Name,
            fieldName: "name",
            name: renderToString("UserGroupsList.column.name"),
            minWidth: 100,
            maxWidth: 200,
            isResizable: true,
          },
          {
            key: UserGroupsListColumnKey.Key,
            fieldName: "key",
            name: renderToString("UserGroupsList.column.key"),
            minWidth: 100,
            maxWidth: 200,
            isResizable: true,
          },
          {
            key: UserGroupsListColumnKey.Role,
            fieldName: "role",
            name: renderToString("UserGroupsList.column.role"),
            minWidth: 100,
            maxWidth: 9999,
            isResizable: true,
          },
          {
            key: UserGroupsListColumnKey.Action,
            fieldName: "action",
            name: "",
            minWidth: 56,
            maxWidth: 56,
          },
        ];
      }, [renderToString]);

    const onItemClick = useCallback(
      (item: UserGroupsListItem) => {
        navigate(`/project/${appID}/user-management/groups/${item.id}/details`);
      },
      [appID, navigate]
    );

    const onRenderItemColumn = useCallback(
      (
        item: UserGroupsListItem,
        _index?: number,
        column?: RolesAndGroupsListColumn
      ) => {
        switch (column?.key) {
          case UserGroupsListColumnKey.Action: {
            return (
              <ActionButtonCell
                icon={<TrashIcon width="1rem" height="1rem" />}
                ariaLabel={renderToString("UserGroupsList.actions.remove")}
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
                {(item[
                  column?.fieldName as keyof UserGroupsListItem
                ] as React.ReactNode) ?? ""}
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
            onItemClick={onItemClick}
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
