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

import styles from "./RoleGroupsList.module.css";
import { Group, Role } from "../../../graphql/adminapi/globalTypes.generated";
import DeleteRoleGroupDialog, {
  DeleteRoleGroupDialogData,
} from "../dialog/DeleteRoleGroupDialog";
import { CopyIconButton } from "../../v2/CopyIconButton/CopyIconButton";

export interface RoleGroupsListItem
  extends Pick<Group, "id" | "name" | "key"> {}

export interface RoleGroupsListRole extends Pick<Role, "id" | "name" | "key"> {}

interface RoleGroupsListProps {
  role: RoleGroupsListRole;
  className?: string;
  groups: RoleGroupsListItem[];
}

export const RoleGroupsList: React.VFC<RoleGroupsListProps> =
  function RoleGroupsList({ role, groups, className }) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();

    const [deleteDialogData, setDeleteDialogData] =
      useState<DeleteRoleGroupDialogData | null>(null);
    const onDismissDeleteDialog = useCallback(
      () => setDeleteDialogData(null),
      []
    );
    const onClickDeleteGroup = useCallback(
      (item: RoleGroupsListItem) => {
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

    const onItemClicked = useCallback(
      (item: RoleGroupsListItem) => {
        navigate(
          `/project/${appID}/user-management/groups/${item.id}/details`
        );
      },
      [appID, navigate]
    );

    const rowActionsLabel = renderToString("RolesList.row-actions");

    if (groups.length === 0) {
      return (
        <>
          <div className={cn(className, styles.listRoot)}>
            <Text as="p" size="2" color="gray" className={styles.empty}>
              <FormattedMessage id="RoleGroupsList.empty" />
            </Text>
          </div>
          <DeleteRoleGroupDialog
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
                  <FormattedMessage id="RoleGroupsList.column.name" />
                </div>
                <div className={styles.tableHeaderCellKey}>
                  <FormattedMessage id="RoleGroupsList.column.key" />
                </div>
                <div className={styles.tableHeaderCellActions} />
              </div>
              {groups.map((item) => (
                <RoleGroupRow
                  key={item.id}
                  item={item}
                  onDelete={onClickDeleteGroup}
                  onItemClicked={onItemClicked}
                  rowActionsLabel={rowActionsLabel}
                />
              ))}
            </div>
          </div>
        </div>
        <DeleteRoleGroupDialog
          data={deleteDialogData}
          onDismiss={onDismissDeleteDialog}
        />
      </>
    );
  };

interface RoleGroupRowProps {
  item: RoleGroupsListItem;
  onDelete: (item: RoleGroupsListItem) => void;
  onItemClicked: (item: RoleGroupsListItem) => void;
  rowActionsLabel: string;
}

function RoleGroupRow({
  item,
  onDelete,
  onItemClicked,
  rowActionsLabel,
}: RoleGroupRowProps) {
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
              <FormattedMessage id="RoleGroupsList.actions.remove" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  );
}
