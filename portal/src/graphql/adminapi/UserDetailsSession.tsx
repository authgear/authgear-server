import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  MessageBar,
  SelectionMode,
  Text,
  TooltipHost,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsSession.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useRevokeSessionMutation } from "./mutations/revokeSessionMutation";
import { useRevokeAllSessionsMutation } from "./mutations/revokeAllSessionsMutation";
import ErrorDialog from "../../error/ErrorDialog";
import { OAuthClientConfig, Session } from "../../types";
import DefaultButton from "../../DefaultButton";
import ActionButton from "../../ActionButton";
import Link from "../../Link";

interface RevokeConfirmationDialogProps {
  isHidden: boolean;
  isLoading: boolean;
  titleKey: string;
  messageKey: string;
  onConfirm: () => void;
  onDismiss: () => void;
}

const RevokeConfirmationDialog: React.VFC<RevokeConfirmationDialogProps> =
  function RevokeConfirmationDialog(props) {
    const { isHidden, isLoading, titleKey, messageKey, onConfirm, onDismiss } =
      props;

    const { renderToString } = useContext(Context);

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
        title: <FormattedMessage id={titleKey} />,
        subText: renderToString(messageKey),
      };
    }, [titleKey, messageKey, renderToString]);

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

interface SessionItemViewModel {
  displayName: string;
  userAgent: string | null;
  clientID: string;
  ipAddress: string;
  lastActivity: string;
  revoke: () => void;
}

interface Props {
  sessions: Session[];
  oauthClients: OAuthClientConfig[];
}

const UserDetailsSession: React.VFC<Props> = function UserDetailsSession(
  props
) {
  const { locale, renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { appID, userID } = useParams() as { appID: string; userID: string };
  const { sessions, oauthClients } = props;

  const {
    revokeSession,
    error: revokeError,
    loading: isRevokeLoading,
  } = useRevokeSessionMutation();
  const {
    revokeAllSessions,
    error: revokeAllError,
    loading: isRevokeAllLoading,
  } = useRevokeAllSessionsMutation();
  const isLoading = isRevokeLoading || isRevokeAllLoading;
  const error = revokeError || revokeAllError;

  interface ConfirmDialogProps {
    titleKey: string;
    messageKey: string;
    onConfirm: () => void;
  }
  const [confirmDialogProps, setConfirmDialogProps] =
    useState<ConfirmDialogProps | null>(null);
  const [isConfirmDialogHidden, setIsConfirmDialogHidden] = useState(true);

  const onConfirmDialogDismiss = useCallback(() => {
    setIsConfirmDialogHidden(true);
  }, []);

  const sessionColumns: IColumn[] = useMemo(
    () => [
      {
        key: "displayName",
        fieldName: "displayName",
        name: renderToString("UserDetails.session.devices"),
        className: styles.cell,
        minWidth: 200,
        maxWidth: 200,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: SessionItemViewModel) => {
          if (item.displayName !== "") {
            return item.displayName;
          }

          if (item.userAgent !== null && item.userAgent !== "") {
            return item.userAgent;
          }

          return (
            <Text
              variant="small"
              styles={(_, theme) => ({
                root: {
                  color: theme.palette.neutralSecondary,
                  fontStyle: "italic",
                },
              })}
            >
              {renderToString("UserDetails.session.devices.unknown")}
            </Text>
          );
        },
      },
      {
        key: "clientID",
        fieldName: "clientID",
        name: renderToString("UserDetails.session.clientID"),
        className: styles.cell,
        minWidth: 120,
        maxWidth: 120,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: SessionItemViewModel) => {
          const client = oauthClients.find(
            (client) => client.client_id === item.clientID
          );

          if (client !== undefined) {
            return (
              <TooltipHost
                content={renderToString(
                  "UserDetails.session.clientID.tooltip.message",
                  { clientID: item.clientID }
                )}
              >
                <Link
                  to={`/project/${appID}/configuration/apps/${item.clientID}/edit`}
                  className={styles.clientID}
                >
                  {client.name}
                </Link>
              </TooltipHost>
            );
          }

          return item.clientID;
        },
      },
      {
        key: "ipAddress",
        fieldName: "ipAddress",
        name: renderToString("UserDetails.session.ip-address"),
        className: styles.cell,
        minWidth: 80,
        maxWidth: 80,
      },
      {
        key: "lastActivity",
        fieldName: "lastActivity",
        name: renderToString("UserDetails.session.last-activity"),
        className: styles.cell,
        minWidth: 220,
        maxWidth: 220,
      },
      {
        key: "action",
        name: renderToString("UserDetails.session.action"),
        minWidth: 60,
        maxWidth: 60,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: SessionItemViewModel) => (
          <ActionButton
            className={styles.actionButton}
            theme={themes.destructive}
            onClick={item.revoke}
            text={<FormattedMessage id="UserDetails.session.action.revoke" />}
          />
        ),
      },
    ],
    [renderToString, oauthClients, appID, themes.destructive]
  );

  const sessionListItems = useMemo(
    () =>
      sessions.map(
        (session): SessionItemViewModel => ({
          displayName: session.displayName,
          userAgent: session.userAgent ?? null,
          clientID: session.clientID ?? "-",
          ipAddress: session.lastAccessedByIP,
          lastActivity: formatDatetime(locale, session.lastAccessedAt) ?? "",
          revoke: () => {
            setConfirmDialogProps({
              titleKey: "UserDetails.session.confirm-dialog.revoke.title",
              messageKey: "UserDetails.session.confirm-dialog.revoke.message",
              onConfirm: () => {
                revokeSession(session.id).finally(() =>
                  setIsConfirmDialogHidden(true)
                );
              },
            });
            setIsConfirmDialogHidden(false);
          },
        })
      ),
    [sessions, locale, revokeSession]
  );

  const onRevokeAllClick = useCallback(() => {
    setConfirmDialogProps({
      titleKey: "UserDetails.session.confirm-dialog.revoke-all.title",
      messageKey: "UserDetails.session.confirm-dialog.revoke-all.message",
      onConfirm: () => {
        revokeAllSessions(userID).finally(() => setIsConfirmDialogHidden(true));
      },
    });
    setIsConfirmDialogHidden(false);
  }, [revokeAllSessions, userID]);

  return (
    <div className={styles.root}>
      <Text as="h2" className={styles.header}>
        <FormattedMessage id="UserDetails.session.header" />
      </Text>
      {sessionListItems.length === 0 ? (
        <MessageBar className={styles.emptyMessageBar}>
          <FormattedMessage id="UserDetails.session.empty" />
        </MessageBar>
      ) : (
        <>
          <DetailsList
            items={sessionListItems}
            columns={sessionColumns}
            selectionMode={SelectionMode.none}
          />
          <DefaultButton
            className={styles.revokeAllButton}
            theme={themes.destructive}
            iconProps={{ iconName: "ErrorBadge" }}
            styles={{
              menuIcon: { paddingLeft: "3px" },
              icon: { paddingRight: "3px" },
            }}
            disabled={sessions.length === 0}
            onClick={onRevokeAllClick}
            text={<FormattedMessage id="UserDetails.session.revoke-all" />}
          />
        </>
      )}
      {confirmDialogProps ? (
        <RevokeConfirmationDialog
          {...confirmDialogProps}
          isHidden={isConfirmDialogHidden}
          isLoading={isLoading}
          onDismiss={onConfirmDialogDismiss}
        />
      ) : null}
      <ErrorDialog
        error={error}
        rules={[]}
        fallbackErrorMessageID="UserDetails.session.revoke-error.generic"
      />
    </div>
  );
};

export default UserDetailsSession;
