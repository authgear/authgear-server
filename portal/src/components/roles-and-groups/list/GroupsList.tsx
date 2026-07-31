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
import styles from "./GroupsList.module.css";
import { Context, FormattedMessage } from "../../../intl";
import PaginationWidget from "../../../PaginationWidget";
import { CopyIconButton } from "../../v2/CopyIconButton/CopyIconButton";
import DeleteGroupDialog, {
  DeleteGroupDialogData,
} from "../dialog/DeleteGroupDialog";
import { GroupsListFragment } from "../../../graphql/adminapi/query/groupsListQuery.generated";

interface GroupsListProps {
  className?: string;
  isSearch: boolean;
  loading: boolean;
  groups: GroupsListFragment | null;
  offset: number;
  pageSize: number;
  totalCount?: number;
  onChangeOffset?: (offset: number) => void;
}

interface GroupListItem {
  id: string;
  key: string;
  name: string | null;
  description: string | null;
}

const GroupsList: React.VFC<GroupsListProps> = function GroupsList(props) {
  const {
    className,
    loading,
    isSearch,
    offset,
    pageSize,
    totalCount,
    onChangeOffset,
  } = props;
  const edges = props.groups?.edges;
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const items: GroupListItem[] = useMemo(() => {
    const items: GroupListItem[] = [];
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
    (item: GroupListItem) => {
      navigate(
        `/project/${appID}/user-management/groups/${item.id}/details`
      );
    },
    [appID, navigate]
  );

  const [deleteGroupDialogData, setDeleteGroupDialogData] =
    useState<DeleteGroupDialogData | null>(null);
  const onClickDeleteGroup = useCallback((item: GroupListItem) => {
    setDeleteGroupDialogData({
      groupID: item.id,
      groupName: item.name,
      groupKey: item.key,
    });
  }, []);
  const dismissDeleteGroupDialog = useCallback(() => {
    setDeleteGroupDialogData(null);
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

  const rowActionsLabel = renderToString("GroupsList.row-actions");

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
          <FormattedMessage id="GroupsList.empty.search" />
        </Text>
        <DeleteGroupDialog
          onDismiss={dismissDeleteGroupDialog}
          data={deleteGroupDialogData}
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
              <FormattedMessage id="GroupsList.column.name" />
            </div>
            <div className={styles.tableHeaderCellKey}>
              <FormattedMessage id="GroupsList.column.key" />
            </div>
            <div className={styles.tableHeaderCellDescription}>
              <FormattedMessage id="GroupsList.column.description" />
            </div>
            <div className={styles.tableHeaderCellActions} />
          </div>
          {items.map((item) => (
            <GroupRow
              key={item.id}
              item={item}
              onDelete={onClickDeleteGroup}
              onItemClicked={onItemClicked}
              rowActionsLabel={rowActionsLabel}
            />
          ))}
        </div>
      </div>
      {!isSearch ? (
        <PaginationWidget className={styles.paginator} {...paginationProps} />
      ) : null}
      <DeleteGroupDialog
        onDismiss={dismissDeleteGroupDialog}
        data={deleteGroupDialogData}
      />
    </div>
  );
};

interface GroupRowProps {
  item: GroupListItem;
  onDelete: (item: GroupListItem) => void;
  onItemClicked: (item: GroupListItem) => void;
  rowActionsLabel: string;
}

function GroupRow({
  item,
  onDelete,
  onItemClicked,
  rowActionsLabel,
}: GroupRowProps) {
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
              <FormattedMessage id="GroupsList.delete-group" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  );
}

export default GroupsList;
