import React from "react";
import cn from "classnames";
import { Link } from "react-router-dom";
import { Persona, PersonaSize, Text, FontIcon } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailSummary.module.css";
import { explorerAddress, parseEIP681 } from "../../util/eip681";
import ExternalLink from "../../ExternalLink";

function shouldRenderExplorerURL(addressURL: string): boolean {
  try {
    parseEIP681(addressURL);
  } catch {
    return false;
  }

  return true;
}

interface UserDetailSummaryProps {
  className?: string;
  isAnonymous: boolean;
  isAnonymized: boolean;
  formattedName?: string;
  endUserAccountIdentifier: string | undefined;
  profileImageURL: string | undefined;
  profileImageEditable: boolean;
  createdAtISO: string | null;
  lastLoginAtISO: string | null;
}

const UserDetailSummary: React.VFC<UserDetailSummaryProps> =
  function UserDetailSummary(props: UserDetailSummaryProps) {
    const {
      isAnonymous,
      isAnonymized,
      formattedName,
      endUserAccountIdentifier,
      profileImageURL,
      profileImageEditable,
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
          {endUserAccountIdentifier &&
          shouldRenderExplorerURL(endUserAccountIdentifier) ? (
            <ExternalLink href={explorerAddress(endUserAccountIdentifier)}>
              <Text className={styles.explorerURL} variant="medium">
                {endUserAccountIdentifier}
              </Text>
            </ExternalLink>
          ) : (
            <Text variant="medium">{endUserAccountIdentifier ?? ""}</Text>
          )}
          <Text className={styles.formattedName} variant="medium">
            {formattedName ? formattedName : ""}
          </Text>
          <Text variant="small">
            <FormattedMessage
              id="UserDetails.signed-up"
              values={{ datetime: formatedSignedUp ?? "" }}
            />
          </Text>
          <Text variant="small">
            <FormattedMessage
              id="UserDetails.last-login-at"
              values={{ datetime: formatedLastLogin ?? "" }}
            />
          </Text>
        </div>
      </section>
    );
  };

export default UserDetailSummary;
