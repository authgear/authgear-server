import React, { useMemo, useContext, useState, useCallback } from "react";
import cn from "classnames";
import {
  Avatar,
  Badge,
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  CaretDownIcon,
  CaretSortIcon,
  CaretUpIcon,
  DotsVerticalIcon,
  EyeNoneIcon,
  LockClosedIcon,
  Pencil1Icon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import { useNavigate, useParams } from "react-router-dom";
import { UsersListFragment } from "./query/usersListQuery.generated";
import {
  UserSortBy,
  SortDirection,
  Role,
  Group,
} from "./globalTypes.generated";

import PaginationWidget from "../../PaginationWidget";
import {
  AccountStatusDialog,
  AccountStatusDialogProps,
} from "./UserDetailsAccountStatus";

import { extractRawID } from "../../util/graphql";
import { formatDatetime } from "../../util/formatDatetime";
import { formatDateOnly } from "../../util/formatDateOnly";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";

import styles from "./UsersList.module.css";
import { useDebounced } from "../../hook/useDebounced";

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
  accountValidFrom: string | null;
  accountValidUntil: string | null;
  temporarilyDisabledFrom: string | null;
  temporarilyDisabledUntil: string | null;
  createdAt: string | null;
  createdAtDateOnly: string | null;
  lastLoginAt: string | null;
  lastLoginAtDateOnly: string | null;
  profilePictureURL: string | null;
  formattedName: string | null;
  endUserAccountID: string | null;
  username: string | null;
  phone: string | null;
  email: string | null;
  roles: UserListRoles;
  groups: UserListGroups;
}

interface DisableUserDialogData {
  accountStatus: UserListItem;
  mode: AccountStatusDialogProps["mode"];
}

const USER_LIST_PLACEHOLDER = "-";

interface UserInfoProps {
  item: UserListItem;
}

function UserInfo(props: UserInfoProps) {
  const {
    item: {
      profilePictureURL,
      formattedName,
      endUserAccountID,
      rawID,
      isAnonymous,
      isAnonymized,
      isDisabled,
    },
  } = props;
  const displayName = isAnonymous ? (
    <Text className={styles.anonymousUserLabel}>
      <FormattedMessage id="UsersList.anonymous-user" />
    </Text>
  ) : isAnonymized ? (
    <Text className={styles.anonymizedUserLabel}>
      <FormattedMessage id="UsersList.anonymized-user" />
    </Text>
  ) : (
    formattedName ?? endUserAccountID ?? rawID
  );
  const fallback =
    (formattedName ?? endUserAccountID ?? rawID)
      .trim()
      .charAt(0)
      .toUpperCase() || "?";

  return (
    <div className={styles.userInfo}>
      <Avatar
        className={styles.userInfoPicture}
        size="3"
        radius="full"
        src={profilePictureURL ?? undefined}
        fallback={fallback}
      />
      <div className={styles.userInfoDisplayName}>
        <Text
          size="2"
          weight="medium"
          className={styles.userInfoDisplayNameText}
        >
          {displayName}
        </Text>
        {isDisabled && !isAnonymized ? (
          <Badge
            color="red"
            size="1"
            radius="small"
            className={styles.disabledBadge}
          >
            <FormattedMessage id="AccountStatusBadge.disabled" />
          </Badge>
        ) : null}
      </div>
      <Text size="1" color="gray" className={styles.userInfoRawID}>
        {rawID}
      </Text>
    </div>
  );
}

function CellText({ value }: { value: string | null }) {
  return (
    <Text size="2" className={styles.cellText}>
      {value ?? USER_LIST_PLACEHOLDER}
    </Text>
  );
}

function UserTextCell({ value }: { value: string | null }) {
  return (
    <div className={styles.tableCellStandard}>
      <CellText value={value} />
    </div>
  );
}

// Shows the date only, with the full datetime (incl. timezone) on hover.
function DateCell({
  dateOnly,
  datetime,
}: {
  dateOnly: string | null;
  datetime: string | null;
}) {
  if (dateOnly == null || datetime == null) {
    return (
      <div className={styles.tableCellDate}>
        <CellText value={dateOnly} />
      </div>
    );
  }
  return (
    <div className={styles.tableCellDate}>
      <Tooltip content={datetime}>
        <Text size="2" className={styles.cellText}>
          {dateOnly}
        </Text>
      </Tooltip>
    </div>
  );
}

function getRelatedItemsText(
  relatedItems: UserListRoles | UserListGroups
): string {
  if (relatedItems.totalCount === 0) {
    return USER_LIST_PLACEHOLDER;
  }
  const firstItem = relatedItems.items[0];
  const additionalCount = relatedItems.totalCount - 1;
  return `${firstItem.name ?? firstItem.key}${
    additionalCount > 0 ? ` +${additionalCount}` : ""
  }`;
}

function SortHeader({
  active,
  sortDirection,
  onClick,
  children,
}: {
  active: boolean;
  sortDirection?: SortDirection;
  onClick: () => void;
  children: React.ReactNode;
}) {
  const SortIcon = !active
    ? CaretSortIcon
    : sortDirection === SortDirection.Asc
    ? CaretUpIcon
    : CaretDownIcon;
  return (
    <button type="button" className={styles.sortButton} onClick={onClick}>
      {children}
      <SortIcon className={styles.sortIcon} />
    </button>
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

  const [loading] = useDebounced(rawLoading, 500);

  const { renderToString, locale } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const [isDisableUserDialogHidden, setIsDisableUserDialogHidden] =
    useState(true);
  const [disableUserDialogData, setDisableUserDialogData] =
    useState<DisableUserDialogData | null>(null);

  const items: UserListItem[] = useMemo(() => {
    const items: UserListItem[] = [];
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
            accountValidFrom: node.accountValidFrom,
            accountValidUntil: node.accountValidUntil,
            temporarilyDisabledFrom: node.temporarilyDisabledFrom,
            temporarilyDisabledUntil: node.temporarilyDisabledUntil,
            createdAt: formatDatetime(locale, node.createdAt),
            createdAtDateOnly: formatDateOnly(locale, node.createdAt),
            lastLoginAt: formatDatetime(locale, node.lastLoginAt),
            lastLoginAtDateOnly: formatDateOnly(locale, node.lastLoginAt),
            profilePictureURL: node.standardAttributes.picture ?? null,
            formattedName: node.formattedName ?? null,
            endUserAccountID: node.endUserAccountID ?? null,
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

  const openAccountStatusDialog = useCallback(
    (item: UserListItem, mode: AccountStatusDialogProps["mode"]) => {
      setDisableUserDialogData({
        accountStatus: item,
        mode,
      });
      setIsDisableUserDialogHidden(false);
    },
    []
  );

  const renderActionCell = useCallback(
    (item: UserListItem) => {
      return (
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            <RadixIconButton
              className={styles.rowActionsButton}
              variant="soft"
              color="gray"
              size="2"
              aria-label={renderToString("action")}
            >
              <DotsVerticalIcon width="1rem" height="1rem" />
            </RadixIconButton>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item
              onSelect={() => {
                navigate(`/project/${appID}/users/${item.id}/details`);
              }}
            >
              <Pencil1Icon />
              <FormattedMessage id="edit" />
            </DropdownMenu.Item>
            <DropdownMenu.Separator />
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                openAccountStatusDialog(item, "disable");
              }}
            >
              <LockClosedIcon />
              <FormattedMessage id="disable" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                openAccountStatusDialog(item, "anonymize-or-schedule");
              }}
            >
              <EyeNoneIcon />
              <FormattedMessage id="UserDetailsAccountStatus.anonymize-user.action.anonymize" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                openAccountStatusDialog(item, "delete-or-schedule");
              }}
            >
              <TrashIcon />
              <FormattedMessage id="delete" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      );
    },
    [appID, navigate, openAccountStatusDialog, renderToString]
  );

  const dismissDisableUserDialog = useCallback(() => {
    setIsDisableUserDialogHidden(true);
  }, []);

  const isEmpty = !loading && items.length === 0;

  return (
    <>
      <div className={cn(styles.root, className)}>
        <div className={styles.tableWrapper}>
          <div className={styles.table}>
            <div className={styles.tableHeader}>
              <div className={styles.tableHeaderCellInfo}>
                <FormattedMessage id="UsersList.column.raw-id" />
              </div>
              <div className={styles.tableHeaderCellStandard}>
                <FormattedMessage id="UsersList.column.username" />
              </div>
              <div className={styles.tableHeaderCellStandard}>
                <FormattedMessage id="UsersList.column.email" />
              </div>
              <div className={styles.tableHeaderCellPhone}>
                <FormattedMessage id="UsersList.column.phone" />
              </div>
              {showRolesAndGroups ? (
                <>
                  <div className={styles.tableHeaderCellStandard}>
                    <FormattedMessage id="UsersList.column.roles" />
                  </div>
                  <div className={styles.tableHeaderCellStandard}>
                    <FormattedMessage id="UsersList.column.groups" />
                  </div>
                </>
              ) : null}
              <div className={styles.tableHeaderCellDate}>
                <SortHeader
                  active={sortBy === UserSortBy.CreatedAt}
                  sortDirection={sortDirection}
                  onClick={() => onColumnClick?.(UserSortBy.CreatedAt)}
                >
                  <FormattedMessage id="UsersList.column.signed-up" />
                </SortHeader>
              </div>
              <div className={styles.tableHeaderCellDate}>
                <SortHeader
                  active={sortBy === UserSortBy.LastLoginAt}
                  sortDirection={sortDirection}
                  onClick={() => onColumnClick?.(UserSortBy.LastLoginAt)}
                >
                  <FormattedMessage id="UsersList.column.last-login-at" />
                </SortHeader>
              </div>
              <div className={styles.tableHeaderCellAction} />
            </div>
            {loading
              ? Array.from({ length: 5 }).map((_, index) => (
                  <div key={index} className={styles.shimmerRow}>
                    <div className={styles.shimmerBlock} />
                  </div>
                ))
              : items.map((item) => (
                  <div
                    key={item.id}
                    className={styles.tableRow}
                    role="button"
                    tabIndex={0}
                    onClick={() => {
                      navigate(`/project/${appID}/users/${item.id}/details`);
                    }}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        navigate(`/project/${appID}/users/${item.id}/details`);
                      }
                    }}
                  >
                    <div className={styles.tableCellInfo}>
                      <UserInfo item={item} />
                    </div>
                    <UserTextCell value={item.username} />
                    <UserTextCell value={item.email} />
                    <div className={styles.tableCellPhone}>
                      <CellText value={item.phone} />
                    </div>
                    {showRolesAndGroups ? (
                      <>
                        <UserTextCell value={getRelatedItemsText(item.roles)} />
                        <UserTextCell
                          value={getRelatedItemsText(item.groups)}
                        />
                      </>
                    ) : null}
                    <DateCell
                      dateOnly={item.createdAtDateOnly}
                      datetime={item.createdAt}
                    />
                    <DateCell
                      dateOnly={item.lastLoginAtDateOnly}
                      datetime={item.lastLoginAt}
                    />
                    <div
                      className={styles.tableCellAction}
                      onClick={(e) => e.stopPropagation()}
                      onKeyDown={(e) => e.stopPropagation()}
                    >
                      {renderActionCell(item)}
                    </div>
                  </div>
                ))}
            {isEmpty ? (
              <div className={styles.emptyRow}>
                <Text size="2" color="gray">
                  <FormattedMessage
                    id={
                      isSearch
                        ? "UsersList.empty.search"
                        : "UsersList.empty.normal"
                    }
                  />
                </Text>
              </div>
            ) : null}
          </div>
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
      </div>
      {disableUserDialogData != null ? (
        <AccountStatusDialog
          key={`${disableUserDialogData.accountStatus.id}:${disableUserDialogData.mode}`}
          isHidden={isDisableUserDialogHidden}
          onDismiss={dismissDisableUserDialog}
          accountStatus={disableUserDialogData.accountStatus}
          mode={disableUserDialogData.mode}
        />
      ) : null}
    </>
  );
};

export default UsersList;
