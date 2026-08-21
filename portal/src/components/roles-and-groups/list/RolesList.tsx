import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  DotsVerticalIcon,
  Pencil1Icon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import styles from "./RolesList.module.css";
import { Context, FormattedMessage } from "../../../intl";
import PaginationWidget from "../../../PaginationWidget";
import { CopyIconButton } from "../../v2/CopyIconButton/CopyIconButton";
import DeleteRoleDialog, {
  DeleteRoleDialogData,
} from "../dialog/DeleteRoleDialog";
import { RolesListFragment } from "../../../graphql/adminapi/query/rolesListQuery.generated";

interface RolesListProps {
  className?: string;
  isSearch: boolean;
  loading: boolean;
  roles: RolesListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface RoleListItem {
  id: string;
  key: string;
  name: string | null;
  description: string | null;
}

const RolesList: React.VFC<RolesListProps> = function RolesList(props) {
  const {
    className,
    loading,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  } = props;
  const edges = props.roles?.edges;
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const items: RoleListItem[] = useMemo(() => {
    const items: RoleListItem[] = [];
    if (edges != null) {
      for (const edge of edges) {
        const node = edge?.node;
        if (node != null) {
          items.push({
            id: node.id,
            name: node.name ?? null,
            key: node.key,
            description: node.description ?? null,
          });
        }
      }
    }
    return items;
  }, [edges]);

  const onItemClicked = useCallback(
    (item: RoleListItem) => {
      navigate(`/project/${appID}/user-management/roles/${item.id}/details`);
    },
    [appID, navigate]
  );

  const [deleteRoleDialogData, setDeleteRoleDialogData] =
    useState<DeleteRoleDialogData | null>(null);
  const onClickDeleteRole = useCallback((item: RoleListItem) => {
    setDeleteRoleDialogData({
      roleID: item.id,
      roleName: item.name,
      roleKey: item.key,
    });
  }, []);
  const dismissDeleteRoleDialog = useCallback(() => {
    setDeleteRoleDialogData(null);
  }, []);

  const paginationProps = useMemo(
    () => ({
      offset,
      pageSize,
      totalCount,
      onChangeOffset,
    }),
    [offset, pageSize, totalCount, onChangeOffset]
  );

  const rowActionsLabel = renderToString("RolesList.row-actions");

  if (items.length === 0 && loading) {
    return (
      <Text as="p" size="2" color="gray" className={styles.loading}>
        <FormattedMessage id="loading" />
      </Text>
    );
  }

  if (items.length === 0 && !loading) {
    return (
      <>
        <Text as="p" size="2" color="gray" className={styles.empty}>
          <FormattedMessage id="RolesList.empty.search" />
        </Text>
        <DeleteRoleDialog
          onDismiss={dismissDeleteRoleDialog}
          data={deleteRoleDialogData}
        />
      </>
    );
  }

  return (
    <div className={cn(className, styles.listRoot)}>
      <div className={styles.tableWrapper}>
        <div className={styles.table}>
          <div className={styles.tableHeader}>
            <div className={styles.tableHeaderCellName}>
              <FormattedMessage id="RolesList.column.name" />
            </div>
            <div className={styles.tableHeaderCellKey}>
              <FormattedMessage id="RolesList.column.key" />
            </div>
            <div className={styles.tableHeaderCellDescription}>
              <FormattedMessage id="RolesList.column.description" />
            </div>
            <div className={styles.tableHeaderCellActions} />
          </div>
          {items.map((item) => (
            <RoleRow
              key={item.id}
              item={item}
              onDelete={onClickDeleteRole}
              onItemClicked={onItemClicked}
              rowActionsLabel={rowActionsLabel}
            />
          ))}
        </div>
      </div>
      {!isSearch ? (
        <PaginationWidget className={styles.paginator} {...paginationProps} />
      ) : null}
      <DeleteRoleDialog
        onDismiss={dismissDeleteRoleDialog}
        data={deleteRoleDialogData}
      />
    </div>
  );
};

interface RoleRowProps {
  item: RoleListItem;
  onDelete: (item: RoleListItem) => void;
  onItemClicked: (item: RoleListItem) => void;
  rowActionsLabel: string;
}

function RoleRow({
  item,
  onDelete,
  onItemClicked,
  rowActionsLabel,
}: RoleRowProps) {
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
      <div className={styles.tableCellDescription}>
        <Text size="2" color="gray" className={styles.descriptionText}>
          {item.description}
        </Text>
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
              onSelect={() => {
                onItemClicked(item);
              }}
            >
              <Pencil1Icon />
              <FormattedMessage id="edit" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                onDelete(item);
              }}
            >
              <TrashIcon />
              <FormattedMessage id="RolesList.delete-role" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  );
}

export default RolesList;
