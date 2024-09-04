import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  MessageBar,
  SelectionMode,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAuthorization.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";
import { Authorization, OAuthClientConfig } from "../../types";
import DefaultButton from "../../DefaultButton";
import ActionButton from "../../ActionButton";
import { useDeleteAuthorizationMutation } from "./mutations/deleteAuthorizationMutation";

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

    const dialogContentProps = useMemo(() => {
      return {
        title: title,
        subText: message,
      };
    }, [title, message]);

    return (
      <Dialog
        hidden={isHidden}
        dialogContentProps={dialogContentProps}
        modalProps={{ isBlocking: isLoading }}
        onDismiss={onDialogDismiss}
      >
        <DialogFooter>
          <ButtonWithLoading
            onClick={onDialogConfirm}
            labelId="confirm"
            loading={isLoading}
          />
          <DefaultButton
            disabled={isLoading}
            onClick={onDialogDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
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
    const { themes } = useSystemConfig();
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

    const authzColumns: IColumn[] = useMemo(
      () => [
        {
          key: "clientName",
          fieldName: "clientName",
          name: renderToString("UserDetails.authorization.client-name"),
          className: styles.cell,
          minWidth: 140,
          maxWidth: 140,
        },
        {
          key: "scopesDesc",
          fieldName: "scopesDesc",
          name: renderToString("UserDetails.authorization.scopes"),
          className: styles.cell,
          minWidth: 200,
          maxWidth: 200,
        },
        {
          key: "createdAt",
          fieldName: "createdAt",
          name: renderToString("UserDetails.authorization.created-at"),
          className: styles.cell,
          minWidth: 220,
          maxWidth: 220,
        },
        {
          key: "action",
          name: renderToString("UserDetails.authorization.action"),
          minWidth: 60,
          maxWidth: 60,
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (item: AuthzItemViewModel) => (
            <ActionButton
              className={styles.actionButton}
              theme={themes.destructive}
              onClick={item.remove}
              text={
                <FormattedMessage id="UserDetails.authorization.action.remove" />
              }
            />
          ),
        },
      ],
      [themes.destructive, renderToString]
    );

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
        <Text as="h2" className={styles.header}>
          <FormattedMessage id="UserDetails.authorization.header" />
        </Text>
        {authzListItems.length === 0 ? (
          <MessageBar className={styles.emptyMessageBar}>
            <FormattedMessage id="UserDetails.authorization.empty" />
          </MessageBar>
        ) : (
          <>
            <DetailsList
              items={authzListItems}
              columns={authzColumns}
              selectionMode={SelectionMode.none}
            />
          </>
        )}
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
