import React from "react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import proPicPlaceholder from "../../image/profile-pic-placeholder.svg";

import { formatDatetime } from "../../util/formatDatetime";
import { UserInfo } from "../../util/user";

import styles from "./UserDetailSummary.module.scss";

interface UserDetailSummaryProps {
  className?: string;
  userInfo: UserInfo;
  createdAtISO: string | null;
  lastLoginAtISO: string | null;
}

const UserDetailSummary: React.FC<UserDetailSummaryProps> = function UserDetailSummary(
  props: UserDetailSummaryProps
) {
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
    <div className={cn(styles.root, className)}>
      <img src={proPicPlaceholder} className={styles.profilePic} />
      <div className={styles.userData}>
        <div className={styles.userInfo}>
          {email && <div className={styles.userInfoRow}>{email}</div>}
          {username && <div className={styles.userInfoRow}>{username}</div>}
        </div>
        <div className={styles.userUpdates}>
          <div className={styles.userInfoRow}>
            <FormattedMessage
              id="UserDetails.signed-up"
              values={{ datetime: formatedSignedUp ?? "" }}
            />
          </div>
          <div className={styles.userInfoRow}>
            <FormattedMessage
              id="UserDetails.last-login-at"
              values={{ datetime: formatedLastLogin ?? "" }}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default UserDetailSummary;
