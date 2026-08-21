import React, { useContext } from "react";
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
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import { CardTable } from "../v2/CardTable/CardTable";
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
      <CardTable>
        <CardTable.Header>
          <CardTable.HeaderCell className={styles.colScope}>
            <FormattedMessage id="ScopeList.columns.scope" />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colDescription}>
            <FormattedMessage id="ScopeList.columns.description" />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colActions} />
        </CardTable.Header>
        {scopes.map((scope) => (
          <CardTable.Row key={scope.id}>
            <CardTable.Cell className={styles.colScope}>
              <span className={styles.scopeChip}>{scope.scope}</span>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colDescription}>
              <Text size="2" className={styles.description}>
                {scope.description}
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colActions}>
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
            </CardTable.Cell>
          </CardTable.Row>
        ))}
      </CardTable>
      <PaginationWidget className={styles.paginator} {...pagination} />
    </div>
  );
};
