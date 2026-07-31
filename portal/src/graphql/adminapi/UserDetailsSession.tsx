import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  Button,
  IconButton,
  Text,
  Tooltip,
} from "@radix-ui/themes";
import { CrossCircledIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsSession.module.css";
import { useRevokeSessionMutation } from "./mutations/revokeSessionMutation";
import { useRevokeAllSessionsMutation } from "./mutations/revokeAllSessionsMutation";
import ErrorDialog from "../../error/ErrorDialog";
import { OAuthClientConfig, Session } from "../../types";
import Link from "../../Link";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Callout } from "../../components/v2/Callout/Callout";

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
        title={<FormattedMessage id={titleKey} />}
        description={<FormattedMessage id={messageKey} />}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onDialogConfirm}
        onCancel={onDialogDismiss}
        loading={isLoading}
        confirmColor="red"
      />
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
      <Text as="p" size="3" weight="medium" className={styles.header}>
        <FormattedMessage id="UserDetails.session.section-header" />
      </Text>
      <div className={styles.content}>
        {sessionListItems.length === 0 ? (
          <Callout
            className={styles.emptyMessageBar}
            type="info"
            showCloseButton={false}
            text={<FormattedMessage id="UserDetails.session.empty" />}
          />
        ) : (
          <>
          <div className={styles.tableContainer}>
            <div className={styles.table}>
              <div className={styles.tableHeader}>
                  <div className={styles.deviceColumn}>
                    <FormattedMessage id="UserDetails.session.devices" />
                  </div>
                  <div className={styles.clientColumn}>
                    <FormattedMessage id="UserDetails.session.clientID" />
                  </div>
                  <div className={styles.ipColumn}>
                    <FormattedMessage id="UserDetails.session.ip-address" />
                  </div>
                  <div className={styles.activityColumn}>
                    <FormattedMessage id="UserDetails.session.last-activity" />
                  </div>
                  <div className={styles.actionColumn} aria-hidden={true} />
              </div>
                {sessionListItems.map((item, index) => {
                  const client = oauthClients.find(
                    (candidate) => candidate.client_id === item.clientID
                  );
                  const deviceName =
                    item.displayName || item.userAgent || null;
                  return (
                    <div
                      className={styles.tableRow}
                      key={`${item.clientID}-${index}`}
                    >
                      <div className={styles.deviceColumn}>
                        {deviceName != null ? (
                          deviceName
                        ) : (
                          <Text size="2" color="gray" className={styles.unknown}>
                            <FormattedMessage id="UserDetails.session.devices.unknown" />
                          </Text>
                        )}
                      </div>
                      <div className={styles.clientColumn}>
                        {client != null ? (
                          <Tooltip
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
                          </Tooltip>
                        ) : (
                          item.clientID
                        )}
                      </div>
                      <div className={styles.ipColumn}>{item.ipAddress}</div>
                      <div className={styles.activityColumn}>
                        {item.lastActivity}
                      </div>
                      <div className={styles.actionColumn}>
                        <Tooltip
                          content={renderToString(
                            "UserDetails.session.action.terminate-session"
                          )}
                        >
                          <IconButton
                            variant="ghost"
                            color="gray"
                            size="2"
                            aria-label={renderToString(
                              "UserDetails.session.action.terminate-session"
                            )}
                            onClick={item.revoke}
                          >
                            <CrossCircledIcon width="1rem" height="1rem" />
                          </IconButton>
                        </Tooltip>
                      </div>
                    </div>
                  );
                })}
            </div>
          </div>
          <Button
            className={styles.revokeAllButton}
            size="2"
            variant="outline"
            color="red"
            disabled={sessions.length === 0}
            onClick={onRevokeAllClick}
          >
            <FormattedMessage id="UserDetails.session.revoke-all" />
          </Button>
          </>
        )}
      </div>
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
