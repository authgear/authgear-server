import React, { useCallback, useContext, useMemo } from "react";
import { Badge, Button, Dialog, Flex, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import {
  AuthorizationScopeKind,
  OAuthClientKind,
  OAuthClientSource,
} from "./globalTypes.generated";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import ExternalLink from "../../ExternalLink";
import { formatDatetime } from "../../util/formatDatetime";
import {
  Authorization,
  AuthorizationScope,
  OAuthClientConfig,
} from "../../types";
import { DynamicClientListItem } from "../../components/dynamic-clients/DynamicClientList";
import styles from "./AuthorizationDetailsDialog.module.css";

// The client an authorization was granted to. A static client comes from
// authgear.yaml; a dynamic one from the Admin API's dynamicClients query.
// "unknown" when neither lookup has it, e.g. while the dynamic clients are
// still loading or failed to load.
export type AuthorizedClient =
  | { kind: "static"; config: OAuthClientConfig }
  | { kind: "dynamic"; client: DynamicClientListItem }
  | { kind: "unknown" };

export interface AuthorizationDetails {
  authorization: Authorization;
  clientName: string;
  client: AuthorizedClient;
  hasFullUserInfo: boolean;
  resolvedPermissionScopes: AuthorizationScope[];
}

export interface AuthorizationDetailsDialogProps {
  details: AuthorizationDetails | null;
  onRevoke: (details: AuthorizationDetails) => void;
  onDismiss: () => void;
}

// A titled block of the dialog. The section title is the strongest small text
// in the dialog; field labels below it are gray, and the resource names inside
// the Permissions cards are weaker still, so the nesting reads top-down.
function Section({
  titleId,
  children,
}: {
  titleId: string;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <section className={styles.section}>
      <Text as="p" size="1" weight="bold" className={styles.sectionTitle}>
        <FormattedMessage id={titleId} />
      </Text>
      <div className={styles.sectionBody}>{children}</div>
    </section>
  );
}

// One label/value row. The label sits in its own column so a run of rows reads
// as a table rather than as alternating lines of text.
function Row({
  labelId,
  children,
}: {
  labelId: string;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <div className={styles.row}>
      <Text size="1" color="gray" className={styles.rowLabel}>
        <FormattedMessage id={labelId} />
      </Text>
      <Text size="2" className={styles.rowValue}>
        {children}
      </Text>
    </div>
  );
}

function URIRow({
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
    <Row labelId={labelId}>
      <ExternalLink href={uri} className={styles.uri}>
        {uri}
      </ExternalLink>
    </Row>
  );
}

// One heading of the Permissions field: a Resource, or the project itself.
// Scopes the project no longer defines get their own group -- filing them under
// the project would claim they are project-level scopes, which they are not.
interface ScopeGroup {
  key: string;
  // null for the project group and for the unknown-scope group, where
  // labelId names the heading instead.
  resource: AuthorizationScope["resource"];
  labelId: string | null;
  scopes: AuthorizationScope[];
}

const PROJECT_GROUP_KEY = "__project__";
const UNKNOWN_GROUP_KEY = "__unknown__";

// Resource groups first, in the order the server resolved them, then the
// project group, then scopes no resource defines.
function groupScopes(resolvedScopes: AuthorizationScope[]): ScopeGroup[] {
  const resourceGroups = new Map<string, ScopeGroup>();
  const projectScopes: AuthorizationScope[] = [];
  const unknownScopes: AuthorizationScope[] = [];

  for (const resolved of resolvedScopes) {
    if (resolved.kind === AuthorizationScopeKind.Project) {
      projectScopes.push(resolved);
      continue;
    }
    const resource = resolved.resource;
    if (resource == null) {
      unknownScopes.push(resolved);
      continue;
    }
    const existing = resourceGroups.get(resource.id);
    if (existing != null) {
      existing.scopes.push(resolved);
    } else {
      resourceGroups.set(resource.id, {
        key: resource.id,
        resource,
        labelId: null,
        scopes: [resolved],
      });
    }
  }

  const groups = [...resourceGroups.values()];
  if (projectScopes.length > 0) {
    groups.push({
      key: PROJECT_GROUP_KEY,
      resource: null,
      labelId: "AuthorizationDetailsDialog.permissions.project",
      scopes: projectScopes,
    });
  }
  if (unknownScopes.length > 0) {
    groups.push({
      key: UNKNOWN_GROUP_KEY,
      resource: null,
      labelId: "AuthorizationDetailsDialog.permissions.unknown-resource",
      scopes: unknownScopes,
    });
  }
  return groups;
}

// Each resource is a bordered card so its name cannot be mistaken for a field
// label of the dialog itself -- the containment carries the nesting, not weight.
function ScopeGroupView({ group }: { group: ScopeGroup }): React.ReactElement {
  return (
    <div className={styles.scopeGroup}>
      <Text as="p" size="2" weight="medium">
        {group.resource != null ? (
          group.resource.name ?? group.resource.resourceURI
        ) : group.labelId != null ? (
          <FormattedMessage id={group.labelId} />
        ) : null}
      </Text>
      {group.resource?.name != null ? (
        <Text as="p" size="1" color="gray" className={styles.uri}>
          {group.resource.resourceURI}
        </Text>
      ) : null}
      <div className={styles.scopeList}>
        {group.scopes.map((resolved) => (
          <Badge
            key={resolved.scope}
            color="gray"
            radius="small"
            className={styles.scopeBadge}
            title={resolved.description ?? undefined}
          >
            {resolved.scope}
          </Badge>
        ))}
      </div>
    </div>
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

    const scopeGroups = useMemo(
      () =>
        details != null ? groupScopes(details.resolvedPermissionScopes) : [],
      [details]
    );

    const dynamicClient =
      details?.client.kind === "dynamic" ? details.client.client : null;

    const hasLinks =
      metadata != null &&
      [
        metadata.clientURI,
        metadata.logoURI,
        metadata.tosURI,
        metadata.policyURI,
      ].some((uri) => uri != null && uri !== "");

    return (
      <Dialog.Root open={details != null} onOpenChange={onOpenChange}>
        <Dialog.Content
          maxWidth="600px"
          size="3"
          onOpenAutoFocus={onOpenAutoFocus}
        >
          <Dialog.Title>{details?.clientName ?? ""}</Dialog.Title>
          {details != null ? (
            <div className={styles.sections}>
              <Section titleId="AuthorizationDetailsDialog.permissions">
                {details.hasFullUserInfo || scopeGroups.length > 0 ? (
                  <div className={styles.scopeGroups}>
                    {details.hasFullUserInfo ? (
                      <Text as="p" size="2">
                        <FormattedMessage id="UserDetails.authorization.scopes.full-userinfo" />
                      </Text>
                    ) : null}
                    {scopeGroups.map((group) => (
                      <ScopeGroupView key={group.key} group={group} />
                    ))}
                  </div>
                ) : (
                  <Text size="2" color="gray">
                    <FormattedMessage id="UserDetails.authorization.scopes.none" />
                  </Text>
                )}
              </Section>

              <Section titleId="AuthorizationDetailsDialog.section.client">
                <Row labelId="AuthorizationDetailsDialog.client-id">
                  <span className={styles.copyRow}>
                    <span className={styles.copyRowText}>
                      {details.authorization.clientID}
                    </span>
                    <CopyIconButton
                      textToCopy={details.authorization.clientID}
                    />
                  </span>
                </Row>

                <Row labelId="AuthorizationDetailsDialog.client-type">
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
                </Row>

                {dynamicClient != null ? (
                  <Row labelId="AuthorizationDetailsDialog.kind">
                    {dynamicClient.kind === OAuthClientKind.FirstParty ? (
                      <FormattedMessage id="DynamicClientDetailsDialog.kind.first-party" />
                    ) : (
                      <FormattedMessage id="DynamicClientDetailsDialog.kind.third-party" />
                    )}
                  </Row>
                ) : null}

                {metadata?.applicationType != null ? (
                  <Row labelId="DynamicClientDetailsDialog.application-type">
                    {metadata.applicationType}
                  </Row>
                ) : null}

                {metadata != null && metadata.redirectURIs.length > 0 ? (
                  <Row labelId="DynamicClientDetailsDialog.redirect-uris">
                    <div className={styles.uriList}>
                      {metadata.redirectURIs.map((uri) => (
                        <span key={uri} className={styles.uri}>
                          {uri}
                        </span>
                      ))}
                    </div>
                  </Row>
                ) : null}
              </Section>

              <Section titleId="AuthorizationDetailsDialog.section.activity">
                <Row labelId="AuthorizationDetailsDialog.authorized-at">
                  {formatDatetime(locale, details.authorization.createdAt) ??
                    ""}
                </Row>
                {/* updatedAt moves each time the client is granted further
                    scopes, so it only carries information when it differs from
                    createdAt. */}
                {details.authorization.updatedAt !==
                details.authorization.createdAt ? (
                  <Row labelId="AuthorizationDetailsDialog.last-extended-at">
                    {formatDatetime(locale, details.authorization.updatedAt) ??
                      ""}
                  </Row>
                ) : null}
                {dynamicClient?.registeredAt != null ? (
                  <Row labelId="DynamicClientDetailsDialog.registered-at">
                    {formatDatetime(locale, dynamicClient.registeredAt) ?? ""}
                  </Row>
                ) : null}
              </Section>

              {hasLinks ? (
                <Section titleId="AuthorizationDetailsDialog.section.links">
                  <URIRow
                    labelId="DynamicClientDetailsDialog.client-uri"
                    uri={metadata.clientURI}
                  />
                  <URIRow
                    labelId="DynamicClientDetailsDialog.logo-uri"
                    uri={metadata.logoURI}
                  />
                  <URIRow
                    labelId="DynamicClientDetailsDialog.tos-uri"
                    uri={metadata.tosURI}
                  />
                  <URIRow
                    labelId="DynamicClientDetailsDialog.policy-uri"
                    uri={metadata.policyURI}
                  />
                </Section>
              ) : null}
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
