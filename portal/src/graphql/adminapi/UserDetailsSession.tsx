import React, { useContext, useMemo } from "react";
import {
  ActionButton,
  DefaultButton,
  DetailsList,
  IColumn,
  SelectionMode,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsSession.module.scss";
import { useSystemConfig } from "../../context/SystemConfigContext";

type SessionType = "IDP" | "OFFLINE_GRANT";

interface SessionUserAgent {
  name: string;
  version: string;
}

interface Session {
  id: string;
  type: SessionType;
  lastAccessedAt: string;
  lastAccessedByIP: string;
  userAgent: SessionUserAgent;
}

function userAgentDisplayName(ua: SessionUserAgent): string {
  let name = ua.name;
  if (name !== "" && ua.version !== "") {
    name += " " + ua.version;
  }
  return name;
}

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
  const { sessions } = props;

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
          deviceName: userAgentDisplayName(session.userAgent),
          kind: renderToString(sessionTypeDisplayKey(session.type)),
          ipAddress: session.lastAccessedByIP,
          lastActivity: formatDatetime(locale, session.lastAccessedAt) ?? "",
          revoke: () => {},
        })
      ),
    [sessions, locale, renderToString]
  );

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
      >
        <FormattedMessage id="UserDetails.session.revoke-all" />
      </DefaultButton>
    </div>
  );
};

export default UserDetailsSession;
