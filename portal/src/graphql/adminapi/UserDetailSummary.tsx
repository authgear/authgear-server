import React from "react";
import cn from "classnames";
import { Persona, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";
import { UserInfo } from "../../util/user";

import styles from "./UserDetailSummary.module.scss";

interface UserDetailSummaryProps {
  className?: string;
  userInfo: UserInfo;
  createdAtISO: string | null;
  lastLoginAtISO: string | null;
}

const UserDetailSummary: React.FC<UserDetailSummaryProps> =
  function UserDetailSummary(props: UserDetailSummaryProps) {
    const { userInfo, createdAtISO, lastLoginAtISO, className } = props;
    const { username, email } = userInfo;
    const { locale } = React.useContext(Context);

    const formatedSignedUp = React.useMemo(() => {
      return formatDatetime(locale, createdAtISO);
    }, [locale, createdAtISO]);
    const formatedLastLogin = React.useMemo(() => {
      return formatDatetime(locale, lastLoginAtISO);
    }, [locale, lastLoginAtISO]);

    return (
      <section className={cn(styles.root, className)}>
        <Persona className={styles.profilePic} />
        <Text className={styles.email}>{email ?? ""}</Text>
        <Text className={styles.username}>{username ?? ""}</Text>
        <Text className={styles.createdAt}>
          <FormattedMessage
            id="UserDetails.signed-up"
            values={{ datetime: formatedSignedUp ?? "" }}
          />
        </Text>
        <Text className={styles.lastLoginAt}>
          <FormattedMessage
            id="UserDetails.last-login-at"
            values={{ datetime: formatedLastLogin ?? "" }}
          />
        </Text>
      </section>
    );
  };

export default UserDetailSummary;
