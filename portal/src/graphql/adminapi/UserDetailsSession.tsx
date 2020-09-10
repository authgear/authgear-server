import React, { useContext, useMemo } from "react";
import { DetailsList, IColumn, SelectionMode } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsSession.module.scss";

interface UserDetailsSession {
  deviceName: string | null;
  lastActivity: string | null;
  ipAddress: string | null;
}

// TODO: replace with actual data
const mockSessionData: UserDetailsSession[] = [
  {
    deviceName: "iPhone 11 Pro",
    lastActivity: "2021-12-15T15:55:20.309775Z",
    ipAddress: "10.2.4.2",
  },
  {
    deviceName: "Chrome 83.0.4.000.84 (DESKTOP-L)",
    lastActivity: "2021-12-15T15:55:20.309775Z",
    ipAddress: "10.2.4.2",
  },
  {
    deviceName: "Pixel XL",
    lastActivity: "2021-12-15T15:55:20.309775Z",
    ipAddress: "10.2.4.2",
  },
];

const UserDetailsSession: React.FC = function UserDetailsSession() {
  const { locale, renderToString } = useContext(Context);
  const sessionColumns: IColumn[] = [
    {
      key: "deviceName",
      fieldName: "deviceName",
      name: renderToString("UserDetails.session.devices"),
      minWidth: 300,
      maxWidth: 300,
    },
    {
      key: "lastActivity",
      fieldName: "lastActivity",
      name: renderToString("UserDetails.session.last-activity"),
      minWidth: 150,
      maxWidth: 150,
    },
    {
      key: "ipAddress",
      fieldName: "ipAddress",
      name: renderToString("UserDetails.session.ip-address"),
      minWidth: 75,
      maxWidth: 75,
    },
  ];

  const sessionListItems = useMemo(
    () =>
      mockSessionData.map((session) => {
        const lastActivityText = formatDatetime(locale, session.lastActivity);
        return { ...session, lastActivity: lastActivityText };
      }),
    [locale]
  );

  return (
    <div className={styles.root}>
      <h1 className={styles.header}>
        <FormattedMessage id="UserDetails.session.header" />
      </h1>
      <DetailsList
        items={sessionListItems}
        columns={sessionColumns}
        selectionMode={SelectionMode.none}
      />
    </div>
  );
};

export default UserDetailsSession;
