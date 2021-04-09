import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ActionButton,
  DefaultButton,
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  SelectionMode,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsSession.module.scss";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useRevokeSessionMutation } from "./mutations/revokeSessionMutation";
import { useRevokeAllSessionsMutation } from "./mutations/revokeAllSessionsMutation";
import { useParams } from "react-router-dom";
import ErrorDialog from "../../error/ErrorDialog";
import { Session, SessionType } from "../../util/user";

interface RevokeConfirmationDialogProps {
  isHidden: boolean;
  isLoading: boolean;
  titleKey: string;
  messageKey: string;
  onConfirm: () => void;
  onDismiss: () => void;
}

const RevokeConfirmationDialog: React.FC<RevokeConfirmationDialogProps> = function RevokeConfirmationDialog(
  props
) {
  const {
    isHidden,
    isLoading,
    titleKey,
    messageKey,
    onConfirm,
    onDismiss,
  } = props;

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
        <DefaultButton disabled={isLoading} onClick={onDialogDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

function sessionTypeDisplayKey(type: SessionType): string {
  switch (type) {
    case "IDP":
      return "UserDetails.session.kind.idp";
    case "OFFLINE_GRANT":
      return "UserDetails.session.kind.offline-grant";
  }
  return "";
}

interface SessionItemViewModel {
  deviceName: string;
  kind: string;
  ipAddress: string;
  lastActivity: string;
  revoke: () => void;
}

interface Props {
  sessions: Session[];
}

const UserDetailsSession: React.FC<Props> = function UserDetailsSession(props) {
  const { locale, renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { userID } = useParams();
  const { sessions } = props;

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
  const [
    confirmDialogProps,
    setConfirmDialogProps,
  ] = useState<ConfirmDialogProps | null>(null);
  const [isConfirmDialogHidden, setIsConfirmDialogHidden] = useState(true);

  const onConfirmDialogDismiss = useCallback(() => {
    setIsConfirmDialogHidden(true);
  }, []);

  const sessionColumns: IColumn[] = useMemo(
    () => [
      {
        key: "deviceName",
        fieldName: "deviceName",
        name: renderToString("UserDetails.session.devices"),
        className: styles.cell,
        minWidth: 200,
        maxWidth: 200,
      },
      {
        key: "kind",
        fieldName: "kind",
        name: renderToString("UserDetails.session.kind"),
        className: styles.cell,
        minWidth: 100,
        maxWidth: 100,
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
        minWidth: 140,
        maxWidth: 140,
      },
      {
        key: "action",
        name: renderToString("UserDetails.session.action"),
        minWidth: 60,
        maxWidth: 60,
        onRender: (item: SessionItemViewModel) => (
          <ActionButton
            className={styles.actionButton}
            theme={themes.destructive}
            onClick={item.revoke}
          >
            <FormattedMessage id="UserDetails.session.action.revoke" />
          </ActionButton>
        ),
      },
    ],
    [themes.destructive, renderToString]
  );

  const sessionListItems = useMemo(
    () =>
      sessions.map(
        (session): SessionItemViewModel => ({
          deviceName: "FIXME(session-info)",
          kind: renderToString(sessionTypeDisplayKey(session.type)),
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
    [sessions, locale, renderToString, revokeSession]
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
      >
        <FormattedMessage id="UserDetails.session.revoke-all" />
      </DefaultButton>
      {confirmDialogProps && (
        <RevokeConfirmationDialog
          {...confirmDialogProps}
          isHidden={isConfirmDialogHidden}
          isLoading={isLoading}
          onDismiss={onConfirmDialogDismiss}
        />
      )}
      <ErrorDialog
        error={error}
        rules={[]}
        fallbackErrorMessageID="UserDetails.session.revoke-error.generic"
      />
    </div>
  );
};

export default UserDetailsSession;
