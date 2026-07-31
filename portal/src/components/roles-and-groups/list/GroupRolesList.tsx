import React, { useCallback, useContext, useState } from "react";
import cn from "classnames";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { DotsVerticalIcon, TrashIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../../intl";
import { useNavigate, useParams } from "react-router-dom";

import styles from "./GroupRolesList.module.css";
import { Group, Role } from "../../../graphql/adminapi/globalTypes.generated";
import DeleteGroupRoleDialog, {
  DeleteGroupRoleDialogData,
} from "../dialog/DeleteGroupRoleDialog";
import { CopyIconButton } from "../../v2/CopyIconButton/CopyIconButton";

export interface GroupRolesListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface GroupRolesListGroup
  extends Pick<Role, "id" | "name" | "key"> {}

interface GroupRolesListProps {
  group: GroupRolesListGroup;
  className?: string;
  roles: GroupRolesListItem[];
}

export const GroupRolesList: React.VFC<GroupRolesListProps> =
  function GroupRolesList({ group, roles, className }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();

    const [deleteDialogData, setDeleteDialogData] =
      useState<DeleteGroupRoleDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteRole = useCallback(
      (item: GroupRolesListItem) => {
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

    const onItemClicked = useCallback(
      (item: GroupRolesListItem) => {
        navigate(
          `/project/${appID}/user-management/roles/${item.id}/details`
        );
      },
      [appID, navigate]
    );

    const rowActionsLabel = renderToString("GroupsList.row-actions");

    if (roles.length === 0) {
      return (
        <>
          <div className={cn(className, styles.listRoot)}>
            <Text as="p" size="2" color="gray" className={styles.empty}>
              <FormattedMessage id="GroupRolesList.empty" />
            </Text>
          </div>
          <DeleteGroupRoleDialog
            data={deleteDialogData}
            onDismiss={onDismissDeleteDialog}
          />
        </>
      );
    }

    return (
      <>
        <div className={cn(className, styles.listRoot)}>
          <div className={styles.tableWrapper}>
            <div className={styles.table}>
              <div className={styles.tableHeader}>
                <div className={styles.tableHeaderCellName}>
                  <FormattedMessage id="GroupRolesList.column.name" />
                </div>
                <div className={styles.tableHeaderCellKey}>
                  <FormattedMessage id="GroupRolesList.column.key" />
                </div>
                <div className={styles.tableHeaderCellActions} />
              </div>
              {roles.map((item) => (
                <GroupRoleRow
                  key={item.id}
                  item={item}
                  onDelete={onClickDeleteRole}
                  onItemClicked={onItemClicked}
                  rowActionsLabel={rowActionsLabel}
                />
              ))}
            </div>
          </div>
        </div>
        <DeleteGroupRoleDialog
          data={deleteDialogData}
          onDismiss={onDismissDeleteDialog}
        />
      </>
    );
  };

interface GroupRoleRowProps {
  item: GroupRolesListItem;
  onDelete: (item: GroupRolesListItem) => void;
  onItemClicked: (item: GroupRolesListItem) => void;
  rowActionsLabel: string;
}

function GroupRoleRow({
  item,
  onDelete,
  onItemClicked,
  rowActionsLabel,
}: GroupRoleRowProps) {
  const onRowClick = useCallback(() => {
    onItemClicked(item);
  }, [onItemClicked, item]);

  const onRowKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onItemClicked(item);
      }
    },
    [onItemClicked, item]
  );

  return (
    <div
      className={styles.tableRow}
      role="button"
      tabIndex={0}
      onClick={onRowClick}
      onKeyDown={onRowKeyDown}
    >
      <div className={styles.tableCellName}>
        <Text size="2" className={styles.nameText}>
          {item.name}
        </Text>
      </div>
      <div className={styles.tableCellKey}>
        <div
          className={styles.keyCell}
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <Text size="2" className={styles.keyText}>
            {item.key}
          </Text>
          <CopyIconButton textToCopy={item.key} />
        </div>
      </div>
      <div
        className={styles.tableCellActions}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
      >
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            <RadixIconButton
              className={styles.rowActionsButton}
              variant="soft"
              color="gray"
              size="2"
              aria-label={rowActionsLabel}
            >
              <DotsVerticalIcon width="1rem" height="1rem" />
            </RadixIconButton>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                onDelete(item);
              }}
            >
              <TrashIcon />
              <FormattedMessage id="GroupRolesList.actions.remove" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  );
}
