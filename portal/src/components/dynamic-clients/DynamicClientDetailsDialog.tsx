import React, { useCallback, useContext } from "react";
import { Button, Dialog, Flex, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import {
  OAuthClientKind,
  OAuthClientSource,
} from "../../graphql/adminapi/globalTypes.generated";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { formatDatetime } from "../../util/formatDatetime";
import { DynamicClientListItem } from "./DynamicClientList";
import styles from "./DynamicClientDetailsDialog.module.css";

export interface DynamicClientDetailsDialogProps {
  client: DynamicClientListItem | null;
  onDelete: (client: DynamicClientListItem) => void;
  onDismiss: () => void;
}

function Field({
  labelId,
  children,
}: {
  labelId: string;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <div className={styles.field}>
      <Text size="1" color="gray">
        <FormattedMessage id={labelId} />
      </Text>
      <Text size="2">{children}</Text>
    </div>
  );
}

export const DynamicClientDetailsDialog: React.VFC<DynamicClientDetailsDialogProps> =
  function DynamicClientDetailsDialog({ client, onDelete, onDismiss }) {
    const { locale } = useContext(Context);

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss();
        }
      },
      [onDismiss]
    );

    const onDeleteClicked = useCallback(() => {
      if (client != null) {
        onDelete(client);
      }
    }, [client, onDelete]);

    return (
      <Dialog.Root open={client != null} onOpenChange={onOpenChange}>
        <Dialog.Content maxWidth="480px" size="3">
          <Dialog.Title>{client?.name ?? ""}</Dialog.Title>
          {client != null ? (
            <div className={styles.fields}>
              <Field labelId="DynamicClientDetailsDialog.client-id">
                <span className={styles.copyRow}>
                  <span className={styles.copyRowText}>{client.clientID}</span>
                  <CopyIconButton textToCopy={client.clientID} />
                </span>
              </Field>
              <Field labelId="DynamicClientDetailsDialog.kind">
                {client.kind === OAuthClientKind.FirstParty ? (
                  <FormattedMessage id="DynamicClientDetailsDialog.kind.first-party" />
                ) : (
                  <FormattedMessage id="DynamicClientDetailsDialog.kind.third-party" />
                )}
              </Field>
              <Field labelId="DynamicClientDetailsDialog.source">
                {client.source === OAuthClientSource.Cimd ? (
                  <FormattedMessage id="DynamicClientDetailsDialog.source.cimd" />
                ) : (
                  <FormattedMessage id="DynamicClientDetailsDialog.source.dcr" />
                )}
              </Field>
              <Field labelId="DynamicClientDetailsDialog.registered-at">
                {formatDatetime(locale, client.registeredAt) ?? ""}
              </Field>
              {client.applicationType != null ? (
                <Field labelId="DynamicClientDetailsDialog.application-type">
                  {client.applicationType}
                </Field>
              ) : null}
              <Field labelId="DynamicClientDetailsDialog.redirect-uris">
                <div className={styles.uriList}>
                  {client.redirectURIs.map((uri) => (
                    <span key={uri} className={styles.uri}>
                      {uri}
                    </span>
                  ))}
                </div>
              </Field>
              <Field labelId="DynamicClientDetailsDialog.grant-types">
                {client.grantTypes.join(", ")}
              </Field>
              <Field labelId="DynamicClientDetailsDialog.response-types">
                {client.responseTypes.join(", ")}
              </Field>
              {client.clientURI != null ? (
                <Field labelId="DynamicClientDetailsDialog.client-uri">
                  {client.clientURI}
                </Field>
              ) : null}
              {client.logoURI != null ? (
                <Field labelId="DynamicClientDetailsDialog.logo-uri">
                  {client.logoURI}
                </Field>
              ) : null}
              {client.tosURI != null ? (
                <Field labelId="DynamicClientDetailsDialog.tos-uri">
                  {client.tosURI}
                </Field>
              ) : null}
              {client.policyURI != null ? (
                <Field labelId="DynamicClientDetailsDialog.policy-uri">
                  {client.policyURI}
                </Field>
              ) : null}
            </div>
          ) : null}
          <Flex gap="3" mt="4" justify="end">
            <Button
              size="2"
              variant="soft"
              color="red"
              onClick={onDeleteClicked}
            >
              <FormattedMessage id="DynamicClientDetailsDialog.delete" />
            </Button>
            <SecondaryButton
              size="2"
              text={<FormattedMessage id="DynamicClientDetailsDialog.close" />}
              onClick={onDismiss}
            />
          </Flex>
        </Dialog.Content>
      </Dialog.Root>
    );
  };
