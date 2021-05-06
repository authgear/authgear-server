import React, { useMemo, useContext, useState, useCallback } from "react";
import cn from "classnames";
import {
  ShimmeredDetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
  IDetailsRowProps,
  DetailsRow,
  ActionButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Link } from "react-router-dom";
import { UsersListQuery_users } from "./__generated__/UsersListQuery";
import { SearchUsersSortBy, SortDirection } from "./__generated__/globalTypes";

import PaginationWidget from "../../PaginationWidget";
import SetUserDisabledDialog from "./SetUserDisabledDialog";

import { formatDatetime } from "../../util/formatDatetime";
import { extractUserInfoFromIdentities } from "../../util/user";
import { nonNullable } from "../../util/types";

import styles from "./UsersList.module.scss";
import { useSystemConfig } from "../../context/SystemConfigContext";

interface UsersListProps {
  className?: string;
  loading: boolean;
  users: UsersListQuery_users | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
  onColumnClick?: (columnKey: SearchUsersSortBy) => void;
  sortBy?: SearchUsersSortBy;
  sortDirection?: SortDirection;
}

interface UserListItem {
  id: string;
  isDisabled: boolean;
  createdAt: string | null;
  username: string | null;
  phone: string | null;
  email: string | null;
  lastLoginAt: string | null;
}

interface DisableUserDialogData {
  isDisablingUser: boolean;
  userID: string;
  username: string | null;
}

const USER_LIST_PLACEHOLDER = "-";

const isUserListItem = (value: unknown): value is UserListItem => {
  if (!(value instanceof Object)) {
    return false;
  }
  return (
    "id" in value && "username" in value && "phone" in value && "email" in value
  );
};

const UsersList: React.FC<UsersListProps> = function UsersList(props) {
  const {
    className,
    loading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
    onColumnClick,
    sortBy,
    sortDirection,
  } = props;
  const edges = props.users?.edges;

  const { renderToString, locale } = useContext(Context);
  const { themes } = useSystemConfig();

  const columns: IColumn[] = [
    {
      key: "username",
      fieldName: "username",
      name: renderToString("UsersList.column.username"),
      minWidth: 150,
    },
    {
      key: "email",
      fieldName: "email",
      name: renderToString("UsersList.column.email"),
      minWidth: 150,
    },
    {
      key: "phone",
      fieldName: "phone",
      name: renderToString("UsersList.column.phone"),
      minWidth: 150,
    },
    {
      key: "createdAt",
      fieldName: "createdAt",
      name: renderToString("UsersList.column.signed-up"),
      minWidth: 200,
      isSorted: sortBy === "CREATED_AT",
      isSortedDescending: sortDirection === SortDirection.DESC,
    },
    {
      key: "lastLoginAt",
      fieldName: "lastLoginAt",
      name: renderToString("UsersList.column.last-login-at"),
      minWidth: 200,
      isSorted: sortBy === "LAST_LOGIN_AT",
      isSortedDescending: sortDirection === SortDirection.DESC,
    },
    {
      key: "action",
      fieldName: "action",
      name: renderToString("action"),
      minWidth: 150,
    },
  ];

  const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] = useState(
    true
  );
  const [
    disableUserDialogData,
    setDisableUserDialogData,
  ] = useState<DisableUserDialogData | null>(null);

  // TODO: consider update UI design to allow multiple email, username and phone number
  const items: UserListItem[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          const identities =
            node.identities?.edges
              ?.map((edge) => edge?.node)
              ?.filter(nonNullable) ?? [];
          const userInfo = extractUserInfoFromIdentities(identities);
          items.push({
            id: node.id,
            isDisabled: node.isDisabled,
            createdAt: formatDatetime(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            username: userInfo.username,
            phone: userInfo.phone,
            email: userInfo.email,
          });
        }
      }
    }
    return items;
  }, [edges, locale]);

  const onRenderUserRow = React.useCallback((props?: IDetailsRowProps) => {
    if (props == null) {
      return null;
    }
    const targetPath = isUserListItem(props.item)
      ? `./${props.item.id}/details`
      : ".";
    return (
      <Link to={targetPath}>
        <DetailsRow {...props} />
      </Link>
    );
  }, []);

  const onDisableUserClicked = useCallback(
    (
      event: React.MouseEvent<unknown>,
      isDisablingUser: boolean,
      userID: string,
      username: string | null
    ) => {
      // prevent triggering the link to user detail page
      event.preventDefault();
      event.stopPropagation();
      setDisableUserDialogData({ isDisablingUser, userID, username });
      setIsDisableUserDialogHidden(false);
    },
    []
  );

  const onRenderUserItemColumn = useCallback(
    (item: UserListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "action":
          return (
            <ActionButton
              className={styles.actionButton}
              styles={{ flexContainer: { alignItems: "normal" } }}
              theme={item.isDisabled ? themes.actionButton : themes.destructive}
              onClick={(event) =>
                onDisableUserClicked(
                  event,
                  !item.isDisabled,
                  item.id,
                  item.username ?? item.email ?? item.phone
                )
              }
            >
              {item.isDisabled ? (
                <FormattedMessage id="UsersList.enable-user" />
              ) : (
                <FormattedMessage id="UsersList.disable-user" />
              )}
            </ActionButton>
          );
        default:
          return (
            <span>
              {item[column?.key as keyof UserListItem] ?? USER_LIST_PLACEHOLDER}
            </span>
          );
      }
    },
    [onDisableUserClicked, themes.actionButton, themes.destructive]
  );

  const dismissDisableUserDialog = useCallback(() => {
    setIsDisableUserDialogHidden(true);
  }, []);

  const onColumnHeaderClick = useCallback(
    (_e, column) => {
      if (column != null) {
        if (column.key === "createdAt") {
          onColumnClick?.(SearchUsersSortBy.CREATED_AT);
        }
        if (column.key === "lastLoginAt") {
          onColumnClick?.(SearchUsersSortBy.LAST_LOGIN_AT);
        }
      }
    },
    [onColumnClick]
  );

  return (
    <>
      <div className={cn(styles.root, className)}>
        <ShimmeredDetailsList
          className={styles.list}
          enableShimmer={loading}
          enableUpdateAnimations={false}
          onRenderRow={onRenderUserRow}
          onRenderItemColumn={onRenderUserItemColumn}
          onColumnHeaderClick={onColumnHeaderClick}
          selectionMode={SelectionMode.none}
          layoutMode={DetailsListLayoutMode.justified}
          columns={columns}
          items={items}
        />
        <PaginationWidget
          className={styles.pagination}
          offset={offset}
          pageSize={pageSize}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
      </div>
      {disableUserDialogData != null && (
        <SetUserDisabledDialog
          isHidden={isDisableUserDialogHidden}
          onDismiss={dismissDisableUserDialog}
          isDisablingUser={disableUserDialogData.isDisablingUser}
          userID={disableUserDialogData.userID}
          username={disableUserDialogData.username}
        />
      )}
    </>
  );
};

export default UsersList;
