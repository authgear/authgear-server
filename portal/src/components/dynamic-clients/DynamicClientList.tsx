import React, { useCallback, useContext } from "react";
import cn from "classnames";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { DotsVerticalIcon, TrashIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import {
  OAuthClientKind,
  OAuthClientSource,
} from "../../graphql/adminapi/globalTypes.generated";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import { CardTable } from "../v2/CardTable/CardTable";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { formatDatetime } from "../../util/formatDatetime";
import styles from "./DynamicClientList.module.css";

export interface DynamicClientListItem {
  id: string;
  clientID: string;
  clientName: string | null;
  name: string;
  kind: OAuthClientKind;
  source: OAuthClientSource;
  registeredAt: string | null;
  applicationType: string | null;
  redirectURIs: string[];
  grantTypes: string[];
  responseTypes: string[];
  logoURI: string | null;
  clientURI: string | null;
  tosURI: string | null;
  policyURI: string | null;
}

interface DynamicClientListProps {
  className?: string;
  clients: DynamicClientListItem[];
  loading: boolean;
  pagination: PaginationProps;
  onDelete: (client: DynamicClientListItem) => void;
  onItemClicked: (item: DynamicClientListItem) => void;
}

function stopPropagation(e: React.SyntheticEvent) {
  e.stopPropagation();
}

interface ClientRowProps {
  client: DynamicClientListItem;
  onDelete: (client: DynamicClientListItem) => void;
  onItemClicked: (item: DynamicClientListItem) => void;
}

const ClientRow: React.VFC<ClientRowProps> = function ClientRow({
  client,
  onDelete,
  onItemClicked,
}) {
  const { renderToString, locale } = useContext(Context);

  const onRowClick = useCallback(() => {
    onItemClicked(client);
  }, [onItemClicked, client]);

  const onRowKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      // Only open when the row itself is focused; Enter/Space on the copy
      // button or the actions menu must not open the details dialog.
      if (e.target !== e.currentTarget) {
        return;
      }
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onItemClicked(client);
      }
    },
    [onItemClicked, client]
  );

  const onDeleteSelect = useCallback(() => {
    onDelete(client);
  }, [onDelete, client]);

  return (
    <CardTable.Row
      className={styles.clientRow}
      role="button"
      tabIndex={0}
      onClick={onRowClick}
      onKeyDown={onRowKeyDown}
    >
      <CardTable.Cell className={styles.colName}>
        <Text size="2" className={styles.clientName}>
          {client.name}
        </Text>
      </CardTable.Cell>
      <CardTable.Cell className={styles.colClientId} onClick={stopPropagation}>
        <Text size="2" className={styles.clientIdText}>
          {client.clientID}
        </Text>
        <CopyIconButton textToCopy={client.clientID} />
      </CardTable.Cell>
      <CardTable.Cell className={styles.colKind}>
        <Text size="2">
          {client.kind === OAuthClientKind.FirstParty ? (
            <FormattedMessage id="DynamicClientList.kind.first-party" />
          ) : (
            <FormattedMessage id="DynamicClientList.kind.third-party" />
          )}
        </Text>
      </CardTable.Cell>
      <CardTable.Cell className={styles.colRegisteredAt}>
        <Text size="2">{formatDatetime(locale, client.registeredAt)}</Text>
      </CardTable.Cell>
      <CardTable.Cell className={styles.colActions} onClick={stopPropagation}>
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            <RadixIconButton
              className={styles.rowActionsButton}
              variant="soft"
              color="gray"
              size="2"
              aria-label={renderToString("DynamicClientList.row-actions")}
            >
              <DotsVerticalIcon width="1rem" height="1rem" />
            </RadixIconButton>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item color="red" onSelect={onDeleteSelect}>
              <TrashIcon />
              <FormattedMessage id="delete" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </CardTable.Cell>
    </CardTable.Row>
  );
};

export const DynamicClientList: React.VFC<DynamicClientListProps> =
  function DynamicClientList(props) {
    const { className, clients, pagination, onDelete, onItemClicked } = props;

    return (
      <div className={cn(className, styles.listRoot)}>
        <CardTable>
          <CardTable.Header>
            <CardTable.HeaderCell className={styles.colName}>
              <FormattedMessage id="DynamicClientList.columns.name" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colClientId}>
              <FormattedMessage id="DynamicClientList.columns.client-id" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colKind}>
              <FormattedMessage id="DynamicClientList.columns.kind" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colRegisteredAt}>
              <FormattedMessage id="DynamicClientList.columns.registered-at" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colActions} />
          </CardTable.Header>
          {clients.map((client) => (
            <ClientRow
              key={client.id}
              client={client}
              onDelete={onDelete}
              onItemClicked={onItemClicked}
            />
          ))}
        </CardTable>
        <PaginationWidget className={styles.paginator} {...pagination} />
      </div>
    );
  };
