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
  MessageBar,
  IListProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Link, useParams } from "react-router-dom";
import { UsersListFragment } from "./query/usersListQuery.generated";
import {
  UserSortBy,
  SortDirection,
  Role,
  Group,
} from "./globalTypes.generated";

import PaginationWidget from "../../PaginationWidget";
import SetUserDisabledDialog from "./SetUserDisabledDialog";

import { extractRawID } from "../../util/graphql";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UsersList.module.css";
import useDelayedValue from "../../hook/useDelayedValue";
import TextCell from "../../components/roles-and-groups/list/common/TextCell";
import ActionButtonCell from "../../components/roles-and-groups/list/common/ActionButtonCell";
import BaseCell from "../../components/roles-and-groups/list/common/BaseCell";

function onShouldVirtualize(_: IListProps): boolean {
  return false;
}

interface UsersListProps {
  className?: string;
  isSearch: boolean;
  loading: boolean;
  users: UsersListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
  onColumnClick?: (columnKey: UserSortBy) => void;
  sortBy?: UserSortBy;
  sortDirection?: SortDirection;
  showRolesAndGroups: boolean;
}

interface UserListRoleItem extends Pick<Role, "id" | "name" | "key"> {}
interface UserListGroupItem extends Pick<Group, "id" | "name" | "key"> {}

interface UserListRoles {
  totalCount: number;
  items: UserListRoleItem[];
}

interface UserListGroups {
  totalCount: number;
  items: UserListGroupItem[];
}

interface UserListItem {
  id: string;
  rawID: string;
  isAnonymous: boolean;
  isAnonymized: boolean;
  isDisabled: boolean;
  isDeactivated: boolean;
  deleteAt: string | null;
  anonymizeAt: string | null;
  createdAt: string | null;
  lastLoginAt: string | null;
  profilePictureURL: string | null;
  formattedName: string | null;
  endUserAccountIdentitifer: string | null;
  username: string | null;
  phone: string | null;
  email: string | null;
  roles: UserListRoles;
  groups: UserListGroups;
}

interface DisableUserDialogData {
  userID: string;
  userDeleteAt: string | null;
  userIsDisabled: boolean;
  userAnonymizeAt: string | null;
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
      isAnonymized,
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
        ) : isAnonymized ? (
          <Text className={styles.anonymizedUserLabel}>
            <FormattedMessage id="UsersList.anonymized-user" />
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
    isSearch,
    loading: rawLoading,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
    onColumnClick,
    sortBy,
    sortDirection,
    showRolesAndGroups,
  } = props;
  const edges = props.users?.edges;

  const loading = useDelayedValue(rawLoading, 500);

  const { renderToString, locale } = useContext(Context);
  const { appID } = useParams() as { appID: string };

  const columns: IColumn[] = useMemo(() => {
    const rolesAndGroupsColumns = showRolesAndGroups
      ? [
          {
            key: "roles",
            name: renderToString("UsersList.column.roles"),
            minWidth: 150,
            columnActionsMode: ColumnActionsMode.disabled,
          },
          {
            key: "groups",
            name: renderToString("UsersList.column.groups"),
            minWidth: 150,
            columnActionsMode: ColumnActionsMode.disabled,
          },
        ]
      : [];

    return [
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
      ...rolesAndGroupsColumns,
      {
        key: "createdAt",
        fieldName: "createdAt",
        name: renderToString("UsersList.column.signed-up"),
        minWidth: 240,
        isSorted: sortBy === "CREATED_AT",
        isSortedDescending: sortDirection === SortDirection.Desc,
        iconName: "SortLines",
        iconClassName: styles.sortIcon,
      },
      {
        key: "lastLoginAt",
        fieldName: "lastLoginAt",
        name: renderToString("UsersList.column.last-login-at"),
        minWidth: 240,
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
  }, [renderToString, showRolesAndGroups, sortBy, sortDirection]);

  const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] =
    useState(true);
  const [disableUserDialogData, setDisableUserDialogData] =
    useState<DisableUserDialogData | null>(null);

  // eslint-disable-next-line complexity
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
            isAnonymized: node.isAnonymized,
            isDisabled: node.isDisabled,
            isDeactivated: node.isDeactivated,
            deleteAt: formatDatetime(locale, node.deleteAt),
            anonymizeAt: formatDatetime(locale, node.anonymizeAt),
            createdAt: formatDatetime(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            profilePictureURL: node.standardAttributes.picture ?? null,
            formattedName: node.formattedName ?? null,
            endUserAccountIdentitifer: node.endUserAccountID ?? null,
            username: node.standardAttributes.preferred_username ?? null,
            phone: node.standardAttributes.phone_number ?? null,
            email: node.standardAttributes.email ?? null,
            roles: {
              totalCount: node.effectiveRoles?.totalCount ?? 0,
              items: (node.effectiveRoles?.edges ?? []).flatMap(
                (edge) => edge?.node ?? []
              ),
            },
            groups: {
              totalCount: node.groups?.totalCount ?? 0,
              items: (node.groups?.edges ?? []).flatMap(
                (edge) => edge?.node ?? []
              ),
            },
          });
        }
      }
    }
    return items;
  }, [edges, locale]);

  const onRenderUserRow = React.useCallback(
    (props?: IDetailsRowProps) => {
      if (props == null) {
        return null;
      }
      const targetPath = isUserListItem(props.item)
        ? `/project/${appID}/users/${props.item.id}/details`
        : ".";
      return (
        <Link to={targetPath}>
          <DetailsRow {...props} />
        </Link>
      );
    },
    [appID]
  );

  const onUserActionClick = useCallback(
    (e: React.MouseEvent<unknown>, item: UserListItem) => {
      e.preventDefault();
      e.stopPropagation();
      setDisableUserDialogData({
        userID: item.id,
        userDeleteAt: item.deleteAt,
        userIsDisabled: item.isDisabled,
        userAnonymizeAt: item.anonymizeAt,
        endUserAccountIdentitifer: item.endUserAccountIdentitifer,
      });
      setIsDisableUserDialogHidden(false);
    },
    []
  );

  const renderUserInfoCell = useCallback((item: UserListItem) => {
    return <UserInfo item={item} />;
  }, []);
  const renderActionCell = useCallback(
    (item: UserListItem) => {
      let variant: "destructive" | "default";
      let text = "";
      if (item.deleteAt != null) {
        variant = "default";
        text = renderToString("UsersList.cancel-removal");
      } else if (item.isDisabled) {
        variant = "default";
        text = renderToString("UsersList.reenable-user");
      } else if (item.isAnonymized) {
        variant = "destructive";
        text = "";
      } else if (item.anonymizeAt != null) {
        variant = "destructive";
        text = renderToString("UsersList.cancel-anonymization");
      } else {
        variant = "destructive";
        text = renderToString("UsersList.disable-user");
      }

      return (
        <ActionButtonCell
          variant={variant}
          onClick={(e) => onUserActionClick(e, item)}
          text={text}
        />
      );
    },
    [onUserActionClick, renderToString]
  );
  const renderRoleCell = useCallback((item: UserListItem) => {
    let text = "-";
    if (item.roles.totalCount !== 0) {
      const addtionalInfo =
        item.roles.totalCount === 1 ? "" : ` +${item.roles.totalCount - 1}`;
      text = `${item.roles.items[0].name}${addtionalInfo}`;
    }
    return (
      <BaseCell>
        <Text className={"whitespace-normal"}>{text}</Text>
      </BaseCell>
    );
  }, []);
  const renderGroupCell = useCallback((item: UserListItem) => {
    let text = "-";
    if (item.groups.totalCount !== 0) {
      const addtionalInfo =
        item.groups.totalCount === 1 ? "" : ` +${item.groups.totalCount - 1}`;
      text = `${item.groups.items[0].name}${addtionalInfo}`;
    }
    return (
      <BaseCell>
        <Text className={"whitespace-normal"}>{text}</Text>
      </BaseCell>
    );
  }, []);
  const renderDefaultCell = useCallback(
    (item: UserListItem, column: IColumn | undefined) => {
      return (
        <TextCell>
          {item[column?.key as keyof UserListItem] ?? USER_LIST_PLACEHOLDER}
        </TextCell>
      );
    },
    []
  );

  const onRenderUserItemColumn = useCallback(
    (item: UserListItem, _index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "info": {
          return renderUserInfoCell(item);
        }
        case "action": {
          return renderActionCell(item);
        }
        case "groups": {
          return renderGroupCell(item);
        }
        case "roles": {
          return renderRoleCell(item);
        }
        default: {
          return renderDefaultCell(item, column);
        }
      }
    },
    [
      renderActionCell,
      renderDefaultCell,
      renderGroupCell,
      renderRoleCell,
      renderUserInfoCell,
    ]
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

  const isEmpty = !loading && items.length === 0;

  return (
    <>
      <div className={cn(styles.root, className)}>
        <div className={cn(styles.listWrapper, isEmpty && styles.empty)}>
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
            // UserList always render fixed number of items, which is not infinite scroll, so no need virtualization
            onShouldVirtualize={onShouldVirtualize}
          />
        </div>
        {!isSearch ? (
          <PaginationWidget
            className={cn(styles.pagination, isEmpty && styles.empty)}
            offset={offset}
            pageSize={pageSize}
            totalCount={totalCount}
            onChangeOffset={onChangeOffset}
          />
        ) : null}
        {isEmpty ? (
          <MessageBar>
            {isSearch ? (
              <FormattedMessage id="UsersList.empty.search" />
            ) : (
              <FormattedMessage id="UsersList.empty.normal" />
            )}
          </MessageBar>
        ) : null}
      </div>
      {disableUserDialogData != null ? (
        <SetUserDisabledDialog
          isHidden={isDisableUserDialogHidden}
          onDismiss={dismissDisableUserDialog}
          userID={disableUserDialogData.userID}
          userDeleteAt={disableUserDialogData.userDeleteAt}
          userAnonymizeAt={disableUserDialogData.userAnonymizeAt}
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
