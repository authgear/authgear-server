import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import DefaultButton from "../../DefaultButton";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  OAuthClientKind,
  OAuthClientSource,
} from "../../graphql/adminapi/globalTypes.generated";
import { TextWithCopyButton } from "../common/TextWithCopyButton";
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
      <Text
        variant="small"
        block={true}
        styles={{ root: { color: "var(--gray-11)" } }}
      >
        <FormattedMessage id={labelId} />
      </Text>
      <Text variant="medium" block={true}>
        {children}
      </Text>
    </div>
  );
}

export const DynamicClientDetailsDialog: React.VFC<DynamicClientDetailsDialogProps> =
  function DynamicClientDetailsDialog({ client, onDelete, onDismiss }) {
    const { renderToString, locale } = useContext(Context);
    const { themes } = useSystemConfig();

    const dialogContentProps = useMemo(
      () => ({
        title: client?.name ?? "",
      }),
      [client?.name]
    );

    const onDeleteClicked = useCallback(() => {
      if (client != null) {
        onDelete(client);
      }
    }, [client, onDelete]);

    return (
      <Dialog
        hidden={client == null}
        dialogContentProps={dialogContentProps}
        minWidth={480}
        onDismiss={onDismiss}
      >
        {client != null ? (
          <div className={styles.fields}>
            <Field labelId="DynamicClientDetailsDialog.client-id">
              <TextWithCopyButton text={client.clientID} />
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
        <DialogFooter>
          <ActionButton
            text={renderToString("DynamicClientDetailsDialog.delete")}
            styles={{
              label: { fontWeight: 600 },
            }}
            theme={themes.destructive}
            onClick={onDeleteClicked}
          />
          <DefaultButton
            onClick={onDismiss}
            text={<FormattedMessage id="DynamicClientDetailsDialog.close" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };
