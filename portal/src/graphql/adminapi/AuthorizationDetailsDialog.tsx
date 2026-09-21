import React, { useCallback, useContext, useMemo } from "react";
import { Badge, Button, Dialog, Flex, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import { OAuthClientKind, OAuthClientSource } from "./globalTypes.generated";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import ExternalLink from "../../ExternalLink";
import { formatDatetime } from "../../util/formatDatetime";
import { Authorization, OAuthClientConfig } from "../../types";
import { DynamicClientListItem } from "../../components/dynamic-clients/DynamicClientList";
import styles from "./AuthorizationDetailsDialog.module.css";

// The client an authorization was granted to. A static client comes from
// authgear.yaml; a dynamic one from the Admin API's dynamicClients query.
// Neither is guaranteed: a client can be deleted while its grants remain.
export type AuthorizedClient =
  | { kind: "static"; config: OAuthClientConfig }
  | { kind: "dynamic"; client: DynamicClientListItem }
  | { kind: "unknown" };

export interface AuthorizationDetails {
  authorization: Authorization;
  clientName: string;
  client: AuthorizedClient;
  hasFullUserInfo: boolean;
  permissionScopes: string[];
}

export interface AuthorizationDetailsDialogProps {
  details: AuthorizationDetails | null;
  onRevoke: (details: AuthorizationDetails) => void;
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

function URIField({
  labelId,
  uri,
}: {
  labelId: string;
  uri: string | null | undefined;
}): React.ReactElement | null {
  if (uri == null || uri === "") {
    return null;
  }
  return (
    <Field labelId={labelId}>
      <ExternalLink href={uri} className={styles.uri}>
        {uri}
      </ExternalLink>
    </Field>
  );
}

interface ClientMetadata {
  applicationType: string | null;
  redirectURIs: string[];
  clientURI: string | null;
  logoURI: string | null;
  tosURI: string | null;
  policyURI: string | null;
}

function clientMetadata(client: AuthorizedClient): ClientMetadata | null {
  switch (client.kind) {
    case "static":
      return {
        applicationType: client.config.x_application_type ?? null,
        redirectURIs: client.config.redirect_uris ?? [],
        clientURI: client.config.client_uri ?? null,
        logoURI: null,
        tosURI: client.config.tos_uri ?? null,
        policyURI: client.config.policy_uri ?? null,
      };
    case "dynamic":
      return {
        applicationType: client.client.applicationType,
        redirectURIs: client.client.redirectURIs,
        clientURI: client.client.clientURI,
        logoURI: client.client.logoURI,
        tosURI: client.client.tosURI,
        policyURI: client.client.policyURI,
      };
    case "unknown":
      return null;
  }
}

export const AuthorizationDetailsDialog: React.VFC<AuthorizationDetailsDialogProps> =
  function AuthorizationDetailsDialog({ details, onRevoke, onDismiss }) {
    const { locale } = useContext(Context);

    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open) {
          onDismiss();
        }
      },
      [onDismiss]
    );

    // Same reasoning as DynamicClientDetailsDialog: Radix would focus the copy
    // button on open and its tooltip would appear unprompted, so focus the
    // dialog itself instead.
    const onOpenAutoFocus = useCallback((event: Event) => {
      event.preventDefault();
      (event.currentTarget as HTMLElement | null)?.focus();
    }, []);

    const onRevokeClicked = useCallback(() => {
      if (details != null) {
        onRevoke(details);
      }
    }, [details, onRevoke]);

    const metadata = useMemo(
      () => (details != null ? clientMetadata(details.client) : null),
      [details]
    );

    const dynamicClient =
      details?.client.kind === "dynamic" ? details.client.client : null;

    return (
      <Dialog.Root open={details != null} onOpenChange={onOpenChange}>
        <Dialog.Content
          maxWidth="480px"
          size="3"
          onOpenAutoFocus={onOpenAutoFocus}
        >
          <Dialog.Title>{details?.clientName ?? ""}</Dialog.Title>
          {details != null ? (
            <div className={styles.fields}>
              <Field labelId="AuthorizationDetailsDialog.client-id">
                <span className={styles.copyRow}>
                  <span className={styles.copyRowText}>
                    {details.authorization.clientID}
                  </span>
                  <CopyIconButton textToCopy={details.authorization.clientID} />
                </span>
              </Field>

              <Field labelId="AuthorizationDetailsDialog.client-type">
                {details.client.kind === "static" ? (
                  <FormattedMessage id="AuthorizationDetailsDialog.client-type.static" />
                ) : dynamicClient != null ? (
                  dynamicClient.source === OAuthClientSource.Cimd ? (
                    <FormattedMessage id="AuthorizationDetailsDialog.client-type.cimd" />
                  ) : (
                    <FormattedMessage id="AuthorizationDetailsDialog.client-type.dcr" />
                  )
                ) : (
                  <Text color="gray">
                    <FormattedMessage id="AuthorizationDetailsDialog.client-type.unknown" />
                  </Text>
                )}
              </Field>

              {dynamicClient != null ? (
                <Field labelId="AuthorizationDetailsDialog.kind">
                  {dynamicClient.kind === OAuthClientKind.FirstParty ? (
                    <FormattedMessage id="DynamicClientDetailsDialog.kind.first-party" />
                  ) : (
                    <FormattedMessage id="DynamicClientDetailsDialog.kind.third-party" />
                  )}
                </Field>
              ) : null}

              <Field labelId="AuthorizationDetailsDialog.permissions">
                {details.hasFullUserInfo ||
                details.permissionScopes.length > 0 ? (
                  <div className={styles.scopeList}>
                    {details.hasFullUserInfo ? (
                      <span>
                        <FormattedMessage id="UserDetails.authorization.scopes.full-userinfo" />
                      </span>
                    ) : null}
                    {details.permissionScopes.map((scope) => (
                      <Badge
                        key={scope}
                        color="gray"
                        radius="small"
                        className={styles.scopeBadge}
                      >
                        {scope}
                      </Badge>
                    ))}
                  </div>
                ) : (
                  <Text color="gray">
                    <FormattedMessage id="UserDetails.authorization.scopes.none" />
                  </Text>
                )}
              </Field>

              <Field labelId="AuthorizationDetailsDialog.authorized-at">
                {formatDatetime(locale, details.authorization.createdAt) ?? ""}
              </Field>
              {/* updatedAt moves each time the client is granted further
                  scopes, so it only carries information when it differs from
                  createdAt. */}
              {details.authorization.updatedAt !==
              details.authorization.createdAt ? (
                <Field labelId="AuthorizationDetailsDialog.last-extended-at">
                  {formatDatetime(locale, details.authorization.updatedAt) ??
                    ""}
                </Field>
              ) : null}

              {dynamicClient?.registeredAt != null ? (
                <Field labelId="DynamicClientDetailsDialog.registered-at">
                  {formatDatetime(locale, dynamicClient.registeredAt) ?? ""}
                </Field>
              ) : null}

              {metadata?.applicationType != null ? (
                <Field labelId="DynamicClientDetailsDialog.application-type">
                  {metadata.applicationType}
                </Field>
              ) : null}

              {metadata != null && metadata.redirectURIs.length > 0 ? (
                <Field labelId="DynamicClientDetailsDialog.redirect-uris">
                  <div className={styles.uriList}>
                    {metadata.redirectURIs.map((uri) => (
                      <span key={uri} className={styles.uri}>
                        {uri}
                      </span>
                    ))}
                  </div>
                </Field>
              ) : null}

              <URIField
                labelId="DynamicClientDetailsDialog.client-uri"
                uri={metadata?.clientURI}
              />
              <URIField
                labelId="DynamicClientDetailsDialog.logo-uri"
                uri={metadata?.logoURI}
              />
              <URIField
                labelId="DynamicClientDetailsDialog.tos-uri"
                uri={metadata?.tosURI}
              />
              <URIField
                labelId="DynamicClientDetailsDialog.policy-uri"
                uri={metadata?.policyURI}
              />
            </div>
          ) : null}
          <Flex gap="3" mt="4" justify="end">
            <Button
              size="2"
              variant="soft"
              color="red"
              onClick={onRevokeClicked}
            >
              <FormattedMessage id="UserDetails.authorization.action.revoke-access" />
            </Button>
            <SecondaryButton
              size="2"
              text={<FormattedMessage id="AuthorizationDetailsDialog.close" />}
              onClick={onDismiss}
            />
          </Flex>
        </Dialog.Content>
      </Dialog.Root>
    );
  };
