import React, { useCallback, useContext } from "react";
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
import { Context, FormattedMessage } from "../../intl";
import { OAuthClientConfig } from "../../types";
import { getApplicationTypeMessageID } from "../../graphql/portal/EditOAuthClientForm";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import styles from "./OAuthClientList.module.css";

interface OAuthClientListProps {
  className?: string;
  clients: OAuthClientConfig[];
  onEdit: (client: OAuthClientConfig) => void;
  onDelete: (client: OAuthClientConfig) => void;
}

export const OAuthClientList: React.VFC<OAuthClientListProps> =
  function OAuthClientList(props) {
    const { className, clients, onEdit, onDelete } = props;
    const { renderToString } = useContext(Context);

    const rowActionsLabel = renderToString("OAuthClientList.row-actions");

    return (
      <div className={cn(className, styles.listRoot)}>
        <div className={styles.tableWrapper}>
          <div className={styles.table}>
            <div className={styles.tableHeader}>
              <div className={styles.tableHeaderCellName}>
                <FormattedMessage id="ApplicationsConfigurationScreen.client-list.name" />
              </div>
              <div className={styles.tableHeaderCellClientId}>
                <FormattedMessage id="ApplicationsConfigurationScreen.client-list.client-id" />
              </div>
              <div className={styles.tableHeaderCellApplicationType}>
                <FormattedMessage id="ApplicationsConfigurationScreen.client-list.application-type" />
              </div>
              <div className={styles.tableHeaderCellActions} />
            </div>
            {clients.map((client) => (
              <OAuthClientRow
                key={client.client_id}
                client={client}
                onEdit={onEdit}
                onDelete={onDelete}
                rowActionsLabel={rowActionsLabel}
              />
            ))}
          </div>
        </div>
      </div>
    );
  };

interface OAuthClientRowProps {
  client: OAuthClientConfig;
  onEdit: (client: OAuthClientConfig) => void;
  onDelete: (client: OAuthClientConfig) => void;
  rowActionsLabel: string;
}

function OAuthClientRow({
  client,
  onEdit,
  onDelete,
  rowActionsLabel,
}: OAuthClientRowProps) {
  const onRowClick = useCallback(() => {
    onEdit(client);
  }, [onEdit, client]);

  const onRowKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        onEdit(client);
      }
    },
    [onEdit, client]
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
          {client.name ?? ""}
        </Text>
      </div>
      <div className={styles.tableCellClientId}>
        <div
          className={styles.clientIdCell}
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <Text size="2" className={styles.clientIdText}>
            {client.client_id}
          </Text>
          <CopyIconButton textToCopy={client.client_id} />
        </div>
      </div>
      <div className={styles.tableCellApplicationType}>
        <Text size="2" className={styles.applicationTypeText}>
          <FormattedMessage
            id={getApplicationTypeMessageID(client.x_application_type)}
          />
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
                onEdit(client);
              }}
            >
              <Pencil1Icon />
              <FormattedMessage id="edit" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                onDelete(client);
              }}
            >
              <TrashIcon />
              <FormattedMessage id="ApplicationsConfigurationScreen.delete-client.label" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </div>
    </div>
  );
}
