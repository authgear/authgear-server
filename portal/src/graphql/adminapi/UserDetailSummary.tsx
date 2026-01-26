import React from "react";
import cn from "classnames";
import { Link } from "react-router-dom";
import { Persona, PersonaSize, Text, FontIcon } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import { formatDatetime } from "../../util/formatDatetime";
import { AccountStatus, AccountStatusBadge } from "./UserDetailsAccountStatus";

import styles from "./UserDetailSummary.module.css";

interface UserDetailSummaryProps {
  className?: string;
  isAnonymous: boolean;
  isAnonymized: boolean;
  rawUserID: string;
  formattedName?: string;
  endUserAccountIdentifier: string | undefined;
  profileImageURL: string | undefined;
  profileImageEditable: boolean;
  createdAtISO: string | null;
  lastLoginAtISO: string | null;
  accountStatus: AccountStatus;
}

const UserDetailSummary: React.VFC<UserDetailSummaryProps> =
  function UserDetailSummary(props: UserDetailSummaryProps) {
    const {
      isAnonymous,
      isAnonymized,
      rawUserID,
      formattedName,
      endUserAccountIdentifier,
      profileImageURL,
      profileImageEditable,
      createdAtISO,
      lastLoginAtISO,
      className,
      accountStatus,
    } = props;
    const { locale } = React.useContext(Context);
    const formatedSignedUp = React.useMemo(() => {
      return formatDatetime(locale, createdAtISO) ?? "";
    }, [locale, createdAtISO]);
    const formatedLastLogin = React.useMemo(() => {
      return formatDatetime(locale, lastLoginAtISO) ?? "";
    }, [locale, lastLoginAtISO]);

    return (
      <section className={cn(styles.root, className)}>
        <div className={styles.profilePic}>
          <Persona
            imageUrl={profileImageURL}
            size={PersonaSize.size72}
            hidePersonaDetails={true}
          />
          {profileImageEditable ? (
            <Link className={styles.cameraLink} to="./edit-picture">
              <FontIcon className={styles.cameraIcon} iconName="Camera" />
            </Link>
          ) : null}
        </div>
        <div className={styles.lines}>
          {isAnonymous ? (
            <Text className={styles.anonymousUserLabel} variant="medium">
              <FormattedMessage id="UsersList.anonymous-user" />
            </Text>
          ) : null}
          {isAnonymized ? (
            <Text className={styles.anonymizedUserLabel} variant="medium">
              <FormattedMessage id="UsersList.anonymized-user" />
            </Text>
          ) : null}
          <Text variant="medium">{rawUserID}</Text>
          <Text variant="medium">{endUserAccountIdentifier ?? ""}</Text>
          <Text className={styles.formattedName} variant="medium">
            {formattedName ? formattedName : ""}
          </Text>
          <AccountStatusBadge
            className={styles.inlineGridItem}
            accountStatus={accountStatus}
          />
          <Text variant="small">
            <FormattedMessage
              id="UserDetails.signed-up"
              values={{ datetime: formatedSignedUp }}
            />
          </Text>
          <Text variant="small">
            <FormattedMessage
              id="UserDetails.last-login-at"
              values={{ datetime: formatedLastLogin }}
            />
          </Text>
        </div>
      </section>
    );
  };

export default UserDetailSummary;
