import React, { useCallback, useContext, useMemo, useState } from "react";
import { Heading, IconButton, Tooltip } from "@radix-ui/themes";
import { CrossCircledIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAuthorization.module.css";
import ErrorDialog from "../../error/ErrorDialog";
import { Authorization, OAuthClientConfig } from "../../types";
import { useDeleteAuthorizationMutation } from "./mutations/deleteAuthorizationMutation";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Callout } from "../../components/v2/Callout/Callout";

function getDisplayNameForClient(
  oauthConfig: OAuthClientConfig[],
  clientID: string
): string {
  for (const config of oauthConfig) {
    if (config.client_id === clientID) {
      return config.name ?? config.client_id;
    }
  }
  return "-";
}

function hasFullUserInfoAccess(scopes: string[]): boolean {
  if (scopes.indexOf("https://authgear.com/scopes/full-userinfo") !== -1) {
    return true;
  }
  return false;
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
  clientName: string;
  remove: () => void;
  createdAt: string;
  scopesDesc: string;
}

interface Props {
  authorizations: Authorization[];
  oauthClientConfig: OAuthClientConfig[];
}

const UserDetailsAuthorization: React.VFC<Props> =
  function UserDetailsAuthorization(props) {
    const { locale, renderToString } = useContext(Context);
    const { authorizations, oauthClientConfig } = props;

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

    const authzListItems = useMemo(() => {
      return authorizations.map(
        (authz): AuthzItemViewModel => ({
          clientName: getDisplayNameForClient(
            oauthClientConfig,
            authz.clientID
          ),
          createdAt: formatDatetime(locale, authz.createdAt) ?? "",
          scopesDesc: hasFullUserInfoAccess(authz.scopes)
            ? renderToString("UserDetails.authorization.scopes.full-userinfo")
            : "-",
          remove: () => {
            setConfirmDialogProps({
              title: renderToString(
                "UserDetails.authorization.confirm-dialog.remove.title"
              ),
              message: renderToString(
                "UserDetails.authorization.confirm-dialog.remove.message"
              ),
              onConfirm: () => {
                deleteAuthorization(authz.id).finally(() =>
                  setIsConfirmDialogHidden(true)
                );
              },
            });
            setIsConfirmDialogHidden(false);
          },
        })
      );
    }, [
      authorizations,
      locale,
      renderToString,
      oauthClientConfig,
      deleteAuthorization,
    ]);

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
                {authzListItems.map((item, index) => (
                  <div
                    className={styles.tableRow}
                    key={`${item.clientName}-${index}`}
                  >
                    <div className={styles.clientColumn}>{item.clientName}</div>
                    <div className={styles.scopeColumn}>{item.scopesDesc}</div>
                    <div className={styles.createdAtColumn}>
                      {item.createdAt}
                    </div>
                    <div className={styles.actionColumn}>
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
        <ErrorDialog
          error={error}
          rules={[]}
          fallbackErrorMessageID="UserDetails.authorization.remove-error.generic"
        />
      </div>
    );
  };

export default UserDetailsAuthorization;
