import React, { useContext } from "react";
import cn from "classnames";
import { DropdownMenu, IconButton as RadixIconButton, Text } from "@radix-ui/themes";
import { DotsVerticalIcon, Pencil1Icon, TrashIcon } from "@radix-ui/react-icons";
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ScopeList.module.css";

interface ScopeListProps {
  className?: string;
  scopes: Scope[];
  loading: boolean;
  pagination: PaginationProps;
  onEdit: (scope: Scope) => void;
  onDelete: (scope: Scope) => void;
}

export const ScopeList: React.VFC<ScopeListProps> = function ScopeList(props) {
  const { className, scopes, pagination, onEdit, onDelete } = props;
  const { renderToString } = useContext(Context);

  return (
    <div className={cn(className, styles.listRoot)}>
      <div className={styles.tableWrapper}>
        <div className={styles.table}>
          <div className={styles.tableHeader}>
            <div className={styles.tableHeaderCellScope}>
              <FormattedMessage id="ScopeList.columns.scope" />
            </div>
            <div className={styles.tableHeaderCellDescription}>
              <FormattedMessage id="ScopeList.columns.description" />
            </div>
            <div className={styles.tableHeaderCellActions} />
          </div>
          {scopes.map((scope) => (
            <div key={scope.id} className={styles.tableRow}>
              <div className={styles.tableCellScope}>
                <span className={styles.scopeChip}>{scope.scope}</span>
              </div>
              <div className={styles.tableCellDescription}>
                <Text size="2" className={styles.description}>
                  {scope.description}
                </Text>
              </div>
              <div className={styles.tableCellActions}>
                <DropdownMenu.Root>
                  <DropdownMenu.Trigger>
                    <RadixIconButton
                      className={styles.rowActionsButton}
                      variant="soft"
                      color="gray"
                      size="2"
                      aria-label={renderToString("ScopeList.row-actions")}
                    >
                      <DotsVerticalIcon width="1rem" height="1rem" />
                    </RadixIconButton>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Content align="end">
                    <DropdownMenu.Item onSelect={() => onEdit(scope)}>
                      <Pencil1Icon />
                      <FormattedMessage id="edit" />
                    </DropdownMenu.Item>
                    <DropdownMenu.Item
                      color="red"
                      onSelect={() => onDelete(scope)}
                    >
                      <TrashIcon />
                      <FormattedMessage id="delete" />
                    </DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu.Root>
              </div>
            </div>
          ))}
        </div>
      </div>
      <PaginationWidget className={styles.paginator} {...pagination} />
    </div>
  );
};
