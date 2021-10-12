import React from "react";
import cn from "classnames";
import { Persona, PersonaSize, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailSummary.module.scss";

interface UserDetailSummaryProps {
  className?: string;
  endUserAccountIdentifier: string | undefined;
  profileImageURL: string | undefined;
  createdAtISO: string | null;
  lastLoginAtISO: string | null;
}

const UserDetailSummary: React.FC<UserDetailSummaryProps> =
  function UserDetailSummary(props: UserDetailSummaryProps) {
    const {
      endUserAccountIdentifier,
      profileImageURL,
      createdAtISO,
      lastLoginAtISO,
      className,
    } = props;
    const { locale } = React.useContext(Context);

    const formatedSignedUp = React.useMemo(() => {
      return formatDatetime(locale, createdAtISO);
    }, [locale, createdAtISO]);
    const formatedLastLogin = React.useMemo(() => {
      return formatDatetime(locale, lastLoginAtISO);
    }, [locale, lastLoginAtISO]);

    return (
      <section className={cn(styles.root, className)}>
        <Persona
          className={styles.profilePic}
          imageUrl={profileImageURL}
          size={PersonaSize.size72}
          hidePersonaDetails={true}
        />

        <Text className={styles.accountID}>
          {endUserAccountIdentifier ?? ""}
        </Text>
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
