import React, { useMemo, useContext, useState, useCallback } from "react";
import cn from "classnames";
import {
  ShimmeredDetailsList,
  DetailsListLayoutMode,
  SelectionMode,
  IColumn,
  IDetailsRowProps,
  DetailsRow,
  ColumnActionsMode,
  Persona,
  PersonaSize,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Link } from "react-router-dom";
import { UsersListFragment } from "./query/usersListQuery.generated";
import { UserSortBy, SortDirection } from "./globalTypes.generated";

import PaginationWidget from "../../PaginationWidget";
import SetUserDisabledDialog from "./SetUserDisabledDialog";

import { extractRawID } from "../../util/graphql";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UsersList.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import useDelayedValue from "../../hook/useDelayedValue";
import ActionButton from "../../ActionButton";

interface UsersListProps {
  className?: string;
  loading: boolean;
  users: UsersListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
  onColumnClick?: (columnKey: UserSortBy) => void;
  sortBy?: UserSortBy;
  sortDirection?: SortDirection;
}

interface UserListItem {
  id: string;
  rawID: string;
  isAnonymous: boolean;
  isDisabled: boolean;
  isDeactivated: boolean;
  deleteAt: string | null;
  createdAt: string | null;
  lastLoginAt: string | null;
  profilePictureURL: string | null;
  formattedName: string | null;
  endUserAccountIdentitifer: string | null;
  username: string | null;
  phone: string | null;
  email: string | null;
}

interface DisableUserDialogData {
  userID: string;
  userDeleteAt: string | null;
  userIsDisabled: boolean;
  endUserAccountIdentitifer: string | null;
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

interface UserInfoProps {
  item: UserListItem;
}

function UserInfo(props: UserInfoProps) {
  const {
    item: {
      profilePictureURL,
      formattedName,
      endUserAccountIdentitifer,
      rawID,
      isAnonymous,
    },
  } = props;
  return (
    <div className={styles.userInfo}>
      <div className={styles.userInfoPicture}>
        <Persona
          imageUrl={profilePictureURL ?? undefined}
          size={PersonaSize.size40}
          hidePersonaDetails={true}
        />
      </div>

      <Text className={styles.userInfoDisplayName}>
        {isAnonymous ? (
          <Text className={styles.anonymousUserLabel}>
            <FormattedMessage id="UsersList.anonymous-user" />
          </Text>
        ) : (
          formattedName ?? endUserAccountIdentitifer
        )}
      </Text>
      <div className={styles.userInfoRawID}>{rawID}</div>
    </div>
  );
}

const UsersList: React.VFC<UsersListProps> = function UsersList(props) {
  const {
    className,
    loading: rawLoading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
    onColumnClick,
    sortBy,
    sortDirection,
  } = props;
  const edges = props.users?.edges;

  const loading = useDelayedValue(rawLoading, 500);

  const { renderToString, locale } = useContext(Context);
  const { themes } = useSystemConfig();

  const columns: IColumn[] = [
    {
      key: "info",
      name: renderToString("UsersList.column.raw-id"),
      minWidth: 300,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "username",
      fieldName: "username",
      name: renderToString("UsersList.column.username"),
      minWidth: 150,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "email",
      fieldName: "email",
      name: renderToString("UsersList.column.email"),
      minWidth: 150,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "phone",
      fieldName: "phone",
      name: renderToString("UsersList.column.phone"),
      minWidth: 120,
      columnActionsMode: ColumnActionsMode.disabled,
    },
    {
      key: "createdAt",
      fieldName: "createdAt",
      name: renderToString("UsersList.column.signed-up"),
      minWidth: 150,
      isSorted: sortBy === "CREATED_AT",
      isSortedDescending: sortDirection === SortDirection.Desc,
      iconName: "SortLines",
      iconClassName: styles.sortIcon,
    },
    {
      key: "lastLoginAt",
      fieldName: "lastLoginAt",
      name: renderToString("UsersList.column.last-login-at"),
      minWidth: 150,
      isSorted: sortBy === "LAST_LOGIN_AT",
      isSortedDescending: sortDirection === SortDirection.Desc,
      iconName: "SortLines",
      iconClassName: styles.sortIcon,
    },
    {
      key: "action",
      fieldName: "action",
      name: renderToString("action"),
      minWidth: 150,
      columnActionsMode: ColumnActionsMode.disabled,
    },
  ];

  const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] =
    useState(true);
  const [disableUserDialogData, setDisableUserDialogData] =
    useState<DisableUserDialogData | null>(null);

  const items: UserListItem[] = useMemo(() => {
    const items = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          items.push({
            id: node.id,
            rawID: extractRawID(node.id),
            isAnonymous: node.isAnonymous,
            isDisabled: node.isDisabled,
            isDeactivated: node.isDeactivated,
            deleteAt: formatDatetime(locale, node.deleteAt),
            createdAt: formatDatetime(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            profilePictureURL: node.standardAttributes.picture ?? null,
            formattedName: node.formattedName ?? null,
            endUserAccountIdentitifer: node.endUserAccountID ?? null,
            username: node.standardAttributes.preferred_username ?? null,
            phone: node.standardAttributes.phone_number ?? null,
            email: node.standardAttributes.email ?? null,
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

  const onUserActionClick = useCallback(
    (e: React.MouseEvent<unknown>, item: UserListItem) => {
      e.preventDefault();
      e.stopPropagation();
      setDisableUserDialogData({
        userID: item.id,
        userDeleteAt: item.deleteAt,
        userIsDisabled: item.isDisabled,
        endUserAccountIdentitifer: item.endUserAccountIdentitifer,
      });
      setIsDisableUserDialogHidden(false);
    },
    []
  );

  const onRenderUserItemColumn = useCallback(
    (item: UserListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "info": {
          return <UserInfo item={item} />;
        }
        case "action": {
          const theme =
            item.deleteAt != null
              ? themes.actionButton
              : item.isDisabled
              ? themes.actionButton
              : themes.destructive;

          const children =
            item.deleteAt != null ? (
              <FormattedMessage id="UsersList.cancel-removal" />
            ) : item.isDisabled ? (
              <FormattedMessage id="UsersList.reenable-user" />
            ) : (
              <FormattedMessage id="UsersList.disable-user" />
            );

          return (
            <div className={styles.cell}>
              <ActionButton
                className={styles.actionButton}
                theme={theme}
                onClick={(event) => onUserActionClick(event, item)}
                text={children}
              />
            </div>
          );
        }
        default:
          return (
            <div className={styles.cell}>
              {item[column?.key as keyof UserListItem] ?? USER_LIST_PLACEHOLDER}
            </div>
          );
      }
    },
    [onUserActionClick, themes.actionButton, themes.destructive]
  );

  const dismissDisableUserDialog = useCallback(() => {
    setIsDisableUserDialogHidden(true);
  }, []);

  const onColumnHeaderClick = useCallback(
    (_e, column) => {
      if (column != null) {
        if (column.key === "createdAt") {
          onColumnClick?.(UserSortBy.CreatedAt);
        }
        if (column.key === "lastLoginAt") {
          onColumnClick?.(UserSortBy.LastLoginAt);
        }
      }
    },
    [onColumnClick]
  );

  return (
    <>
      <div className={cn(styles.root, className)}>
        <div className={styles.listWrapper}>
          <ShimmeredDetailsList
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
        </div>
        <PaginationWidget
          className={styles.pagination}
          offset={offset}
          pageSize={pageSize}
          totalCount={totalCount}
          onChangeOffset={onChangeOffset}
        />
      </div>
      {disableUserDialogData != null ? (
        <SetUserDisabledDialog
          isHidden={isDisableUserDialogHidden}
          onDismiss={dismissDisableUserDialog}
          userID={disableUserDialogData.userID}
          userDeleteAt={disableUserDialogData.userDeleteAt}
          userIsDisabled={disableUserDialogData.userIsDisabled}
          endUserAccountIdentifier={
            disableUserDialogData.endUserAccountIdentitifer ?? undefined
          }
        />
      ) : null}
    </>
  );
};

export default UsersList;
