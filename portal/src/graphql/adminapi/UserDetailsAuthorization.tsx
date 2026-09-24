import React, { useCallback, useContext, useMemo, useState } from "react";
import { Badge, Heading, IconButton, Text, Tooltip } from "@radix-ui/themes";
import { CrossCircledIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAuthorization.module.css";
import ErrorDialog from "../../error/ErrorDialog";
import {
  Authorization,
  AuthorizationScope,
  OAuthClientConfig,
} from "../../types";
import { useDeleteAuthorizationMutation } from "./mutations/deleteAuthorizationMutation";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Callout } from "../../components/v2/Callout/Callout";
import { DynamicClientListItem } from "../../components/dynamic-clients/DynamicClientList";
import {
  AuthorizationDetails,
  AuthorizationDetailsDialog,
  AuthorizedClient,
} from "./AuthorizationDetailsDialog";

// A dynamic client (DCR-registered or CIMD-resolved) is not in
// authgear.yaml, so the static client list cannot describe it. The caller
// passes the project's dynamic clients as the Admin API returned them, keyed
// by client ID.
export type DynamicClients = ReadonlyMap<string, DynamicClientListItem>;

function resolveClient(
  oauthConfig: OAuthClientConfig[],
  dynamicClients: DynamicClients,
  clientID: string
): AuthorizedClient {
  for (const config of oauthConfig) {
    if (config.client_id === clientID) {
      return { kind: "static", config };
    }
  }
  const client = dynamicClients.get(clientID);
  if (client != null) {
    return { kind: "dynamic", client };
  }
  return { kind: "unknown" };
}

function displayNameForClient(
  client: AuthorizedClient,
  clientID: string
): string {
  switch (client.kind) {
    case "static":
      return client.config.name ?? client.config.client_id;
    case "dynamic":
      return client.client.name;
    case "unknown":
      // The raw client ID rather than a dash: it identifies the grant even
      // when neither list knows the client.
      return clientID;
  }
}

const FULL_USERINFO_SCOPE = "https://authgear.com/scopes/full-userinfo";

// How many scope badges a row shows before collapsing the rest into a
// "+N more" badge; the details dialog lists them all.
const VISIBLE_SCOPE_BADGES = 2;

// Scopes that every grant carries and that grant no permission of their own,
// so listing them would only add noise to the Permission column.
const PROTOCOL_SCOPES: ReadonlySet<string> = new Set([
  "openid",
  "offline_access",
]);

function hasFullUserInfoAccess(scopes: string[]): boolean {
  return scopes.includes(FULL_USERINFO_SCOPE);
}

// Whether a granted scope is worth showing as a permission: everything except
// the protocol scopes and the full-userinfo scope, which gets its own label.
function isPermissionScope(scope: string): boolean {
  return !PROTOCOL_SCOPES.has(scope) && scope !== FULL_USERINFO_SCOPE;
}

function permissionScopes(scopes: string[]): string[] {
  return scopes.filter(isPermissionScope);
}

// The same scopes as permissionScopes, resolved to the resource that defines
// each one. A name two resources define appears once per resource, so this can
// be longer than permissionScopes -- the row counts names, the dialog groups
// by resource.
function resolvedPermissionScopes(
  resolved: AuthorizationScope[]
): AuthorizationScope[] {
  return resolved.filter((r) => isPermissionScope(r.scope));
}

interface RemoveConfirmationDialogProps {
  isHidden: boolean;
  isLoading: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onDismiss: () => void;
}

const RemoveConfirmationDialog: React.VFC<RemoveConfirmationDialogProps> =
  function RemoveConfirmationDialog(props) {
    const { isHidden, isLoading, title, message, onConfirm, onDismiss } = props;

    const onDialogConfirm = useCallback(() => {
      if (!isHidden && !isLoading) {
        onConfirm();
      }
    }, [isHidden, isLoading, onConfirm]);

    const onDialogDismiss = useCallback(() => {
      if (!isHidden && !isLoading) {
        onDismiss();
      }
    }, [isHidden, isLoading, onDismiss]);

    return (
      <ConfirmationDialog
        open={!isHidden}
        onOpenChange={(open) => {
          if (!open) {
            onDialogDismiss();
          }
        }}
        title={title}
        description={message}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onDialogConfirm}
        onCancel={onDialogDismiss}
        loading={isLoading}
        confirmColor="red"
      />
    );
  };

interface AuthzItemViewModel {
  details: AuthorizationDetails;
  remove: () => void;
  createdAt: string;
}

interface Props {
  authorizations: Authorization[];
  oauthClientConfig: OAuthClientConfig[];
  dynamicClients: DynamicClients;
}

const UserDetailsAuthorization: React.VFC<Props> =
  function UserDetailsAuthorization(props) {
    const { locale, renderToString } = useContext(Context);
    const { authorizations, oauthClientConfig, dynamicClients } = props;

    const {
      deleteAuthorization,
      error: deleteAuthorizationError,
      loading: isDeletingAuthorization,
    } = useDeleteAuthorizationMutation();

    const isLoading = isDeletingAuthorization;
    const error = deleteAuthorizationError;

    interface ConfirmDialogProps {
      title: string;
      message: string;
      onConfirm: () => void;
    }
    const [confirmDialogProps, setConfirmDialogProps] =
      useState<ConfirmDialogProps | null>(null);
    const [isConfirmDialogHidden, setIsConfirmDialogHidden] = useState(true);

    const onConfirmDialogDismiss = useCallback(() => {
      setIsConfirmDialogHidden(true);
    }, []);

    const [detailsItem, setDetailsItem] = useState<AuthorizationDetails | null>(
      null
    );
    const onDismissDetails = useCallback(() => {
      setDetailsItem(null);
    }, []);

    const authzListItems = useMemo(() => {
      return authorizations.map((authz): AuthzItemViewModel => {
        const client = resolveClient(
          oauthClientConfig,
          dynamicClients,
          authz.clientID
        );
        return {
          details: {
            authorization: authz,
            clientName: displayNameForClient(client, authz.clientID),
            client,
            hasFullUserInfo: hasFullUserInfoAccess(authz.scopes),
            permissionScopes: permissionScopes(authz.scopes),
            resolvedPermissionScopes: resolvedPermissionScopes(
              authz.resolvedScopes
            ),
          },
          createdAt: formatDatetime(locale, authz.createdAt) ?? "",
          remove: () => {
            setConfirmDialogProps({
              title: renderToString(
                "UserDetails.authorization.confirm-dialog.remove.title"
              ),
              message: renderToString(
                "UserDetails.authorization.confirm-dialog.remove.message"
              ),
              onConfirm: () => {
                deleteAuthorization(authz.id).finally(() => {
                  setIsConfirmDialogHidden(true);
                  setDetailsItem(null);
                });
              },
            });
            setIsConfirmDialogHidden(false);
          },
        };
      });
    }, [
      authorizations,
      locale,
      renderToString,
      oauthClientConfig,
      dynamicClients,
      deleteAuthorization,
    ]);

    // Revoke from the details dialog: reuse the row's confirmation so both
    // paths ask the same question and run the same mutation.
    const onRevokeFromDetails = useCallback(
      (details: AuthorizationDetails) => {
        const item = authzListItems.find(
          (candidate) =>
            candidate.details.authorization.id === details.authorization.id
        );
        item?.remove();
      },
      [authzListItems]
    );

    return (
      <div className={styles.root}>
        <Heading as="h2" size="3" weight="medium" className={styles.header}>
          <FormattedMessage id="UserDetails.authorization.header" />
        </Heading>
        <div className={styles.content}>
          {authzListItems.length === 0 ? (
            <Callout
              className={styles.emptyMessageBar}
              type="info"
              showCloseButton={false}
              text={<FormattedMessage id="UserDetails.authorization.empty" />}
            />
          ) : (
            <div className={styles.tableContainer}>
              <div className={styles.table}>
                <div className={styles.tableHeader}>
                  <div className={styles.clientColumn}>
                    <FormattedMessage id="UserDetails.authorization.client-name" />
                  </div>
                  <div className={styles.scopeColumn}>
                    <FormattedMessage id="UserDetails.authorization.scopes" />
                  </div>
                  <div className={styles.createdAtColumn}>
                    <FormattedMessage id="UserDetails.authorization.created-at" />
                  </div>
                  <div className={styles.actionColumn} aria-hidden={true} />
                </div>
                {authzListItems.map((item) => (
                  <div
                    className={styles.tableRow}
                    key={item.details.authorization.id}
                    role="button"
                    tabIndex={0}
                    onClick={() => setDetailsItem(item.details)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" || e.key === " ") {
                        e.preventDefault();
                        setDetailsItem(item.details);
                      }
                    }}
                  >
                    <div className={styles.clientColumn}>
                      <Tooltip
                        content={renderToString(
                          "UserDetails.session.clientID.tooltip.message",
                          { clientID: item.details.authorization.clientID }
                        )}
                      >
                        <span className={styles.clientName}>
                          {item.details.clientName}
                        </span>
                      </Tooltip>
                    </div>
                    <div className={styles.scopeColumn}>
                      {item.details.hasFullUserInfo ||
                      item.details.permissionScopes.length > 0 ? (
                        <div className={styles.scopeList}>
                          {item.details.hasFullUserInfo ? (
                            <Text size="2">
                              <FormattedMessage id="UserDetails.authorization.scopes.full-userinfo" />
                            </Text>
                          ) : null}
                          {item.details.permissionScopes
                            .slice(0, VISIBLE_SCOPE_BADGES)
                            .map((scope) => (
                              <Badge
                                key={scope}
                                color="gray"
                                radius="small"
                                className={styles.scopeBadge}
                              >
                                {scope}
                              </Badge>
                            ))}
                          {item.details.permissionScopes.length >
                          VISIBLE_SCOPE_BADGES ? (
                            <Badge
                              color="gray"
                              radius="small"
                              variant="outline"
                            >
                              <FormattedMessage
                                id="UserDetails.authorization.scopes.more"
                                values={{
                                  count:
                                    item.details.permissionScopes.length -
                                    VISIBLE_SCOPE_BADGES,
                                }}
                              />
                            </Badge>
                          ) : null}
                        </div>
                      ) : (
                        <Text size="2" color="gray">
                          <FormattedMessage id="UserDetails.authorization.scopes.none" />
                        </Text>
                      )}
                    </div>
                    <div className={styles.createdAtColumn}>
                      {item.createdAt}
                    </div>
                    <div
                      className={styles.actionColumn}
                      // Stop the row's click-to-open so the action buttons
                      // do only what they say.
                      onClick={(e) => e.stopPropagation()}
                      onKeyDown={(e) => e.stopPropagation()}
                    >
                      <Tooltip
                        content={renderToString(
                          "UserDetails.authorization.action.revoke-access"
                        )}
                      >
                        <IconButton
                          variant="ghost"
                          color="gray"
                          size="2"
                          aria-label={renderToString(
                            "UserDetails.authorization.action.revoke-access"
                          )}
                          onClick={item.remove}
                        >
                          <CrossCircledIcon width="1rem" height="1rem" />
                        </IconButton>
                      </Tooltip>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
        {confirmDialogProps ? (
          <RemoveConfirmationDialog
            {...confirmDialogProps}
            isHidden={isConfirmDialogHidden}
            isLoading={isLoading}
            onDismiss={onConfirmDialogDismiss}
          />
        ) : null}
        <AuthorizationDetailsDialog
          details={detailsItem}
          onRevoke={onRevokeFromDetails}
          onDismiss={onDismissDetails}
        />
        <ErrorDialog
          error={error}
          rules={[]}
          fallbackErrorMessageID="UserDetails.authorization.remove-error.generic"
        />
      </div>
    );
  };

export default UserDetailsAuthorization;
