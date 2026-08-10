import React, { useContext, useCallback } from "react";
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
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { ResourceListEmptyView } from "./ResourceListEmptyView";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ResourceList.module.css";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";

interface ResourceListItem
  extends Pick<Resource, "id" | "name" | "resourceURI"> {}

interface ResourceListProps {
  className?: string;
  resources: ResourceListItem[];
  loading: boolean;
  pagination: PaginationProps;
  onDelete: (resource: Resource) => void;
  onItemClicked: (item: ResourceListItem) => void;
  onCreateClick: () => void;
}

export const ResourceList: React.VFC<ResourceListProps> = function ResourceList(
  props
) {
  const {
    className,
    resources,
    loading,
    pagination,
    onDelete,
    onItemClicked,
    onCreateClick,
  } = props;
  const { renderToString } = useContext(Context);

  if (resources.length === 0 && !loading) {
    return (
      <ResourceListEmptyView
        className={className}
        onCreateClick={onCreateClick}
      />
    );
  }

  if (resources.length === 0 && loading) {
    return (
      <Text as="p" size="2" color="gray" className={styles.loading}>
        <FormattedMessage id="loading" />
      </Text>
    );
  }

  return (
    <div className={cn(className, styles.listRoot)}>
      <div className={styles.tableWrapper}>
        <div className={styles.table}>
          <div className={styles.tableHeader}>
            <div className={styles.tableHeaderCellName}>
              <FormattedMessage id="ResourceList.columns.name" />
            </div>
            <div className={styles.tableHeaderCellIdentifier}>
              <FormattedMessage id="ResourceList.columns.identifier" />
            </div>
            <div className={styles.tableHeaderCellActions} />
          </div>
          {resources.map((resource) => (
            <ResourceRow
              key={resource.id}
              resource={resource}
              onDelete={onDelete}
              onItemClicked={onItemClicked}
              rowActionsLabel={renderToString("ResourceList.row-actions")}
            />
          ))}
        </div>
      </div>
      <PaginationWidget className={styles.paginator} {...pagination} />
    </div>
  );
};

interface ResourceRowProps {
  resource: ResourceListItem;
  onDelete: (resource: Resource) => void;
  onItemClicked: (item: ResourceListItem) => void;
  rowActionsLabel: string;
}

function ResourceRow({
  resource,
  onDelete,
  onItemClicked,
  rowActionsLabel,
}: ResourceRowProps) {
  const onRowClick = useCallback(() => {
    onItemClicked(resource);
  }, [onItemClicked, resource]);

  const onRowKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onItemClicked(resource);
      }
    },
    [onItemClicked, resource]
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
        <Text size="2" className={styles.resourceName}>
          {resource.name}
        </Text>
      </div>
      <div className={styles.tableCellIdentifier}>
        <div
          className={styles.identifier}
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <Text size="2" className={styles.identifierText}>
            {resource.resourceURI}
          </Text>
          <CopyIconButton textToCopy={resource.resourceURI} />
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
              onSelect={() => {
                onItemClicked(resource);
              }}
            >
              <Pencil1Icon />
              <FormattedMessage id="edit" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                onDelete(resource as Resource);
              }}
            >
              <TrashIcon />
              <FormattedMessage id="delete" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  );
}
