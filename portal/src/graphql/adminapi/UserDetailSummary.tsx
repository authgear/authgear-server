import React, { useMemo } from "react";
import cn from "classnames";
import { Avatar, Badge, Text } from "@radix-ui/themes";
import { CameraIcon, ChevronLeftIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import Link from "../../Link";
import { FormattedMessage, Context } from "../../intl";
import { AccountStatus } from "./UserDetailsAccountStatus";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";

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

function AccountStatusRadixBadge({
  accountStatus,
}: {
  accountStatus: AccountStatus;
}): React.ReactElement {
  const { isDisabled, isAnonymized, deleteAt, anonymizeAt } = accountStatus;

  if (isAnonymized) {
    return (
      <Badge color="gray" radius="small">
        <FormattedMessage id="AccountStatusBadge.anonymized" />
      </Badge>
    );
  }
  if (deleteAt != null) {
    return (
      <Badge color="red" radius="small">
        <FormattedMessage id="AccountStatusBadge.scheduled-deletion" />
      </Badge>
    );
  }
  if (anonymizeAt != null) {
    return (
      <Badge color="orange" radius="small">
        <FormattedMessage id="AccountStatusBadge.scheduled-anonymization" />
      </Badge>
    );
  }
  if (isDisabled) {
    return (
      <Badge color="red" radius="small">
        <FormattedMessage id="AccountStatusBadge.disabled" />
      </Badge>
    );
  }
  return (
    <Badge color="green" radius="small">
      <FormattedMessage id="UserDetails.account-status.active" />
    </Badge>
  );
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
      className,
      accountStatus,
    } = props;
    const { renderToString } = React.useContext(Context);
    const { appID } = useParams() as { appID: string };

    const usersListPath = useMemo(
      () => `/project/${appID}/user-management/users`,
      [appID]
    );

    const displayName = React.useMemo(() => {
      if (isAnonymized) {
        return renderToString("UsersList.anonymized-user");
      }
      if (isAnonymous) {
        return renderToString("UsersList.anonymous-user");
      }
      return (
        formattedName || endUserAccountIdentifier || rawUserID
      );
    }, [
      isAnonymized,
      isAnonymous,
      formattedName,
      endUserAccountIdentifier,
      rawUserID,
      renderToString,
    ]);

    const initials = React.useMemo(() => {
      const name = endUserAccountIdentifier || rawUserID || "U";
      return name[0].toUpperCase();
    }, [endUserAccountIdentifier, rawUserID]);

    return (
      <div className={cn(styles.root, className)}>
        <Link to={usersListPath} className={styles.backLink}>
          <ChevronLeftIcon className={styles.backLinkIcon} />
          <FormattedMessage id="UsersScreen.title" />
        </Link>
        <div className={styles.headerRow}>
          <div className={styles.profilePic}>
            <Avatar
              className={styles.avatar}
              src={profileImageURL}
              fallback={initials}
              size="5"
              radius="full"
            />
            {profileImageEditable ? (
              <Link
                to="./edit-picture"
                className={styles.cameraLink}
                aria-label={renderToString("EditPictureScreen.title")}
              >
                <CameraIcon className={styles.cameraIcon} />
              </Link>
            ) : null}
          </div>
          <div className={styles.headerInfo}>
            <div className={styles.nameRow}>
              <Text as="p" size="6" weight="bold" className={styles.displayName}>
                {displayName}
              </Text>
              <AccountStatusRadixBadge accountStatus={accountStatus} />
            </div>
            <div className={styles.userIdPill}>
              <Text size="1" color="gray">
                <FormattedMessage id="UserDetailSummary.user-id-label" />
              </Text>
              <div className={styles.userIdContent}>
                <Text size="1" color="gray">
                  {rawUserID}
                </Text>
                <CopyIconButton textToCopy={rawUserID} />
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

export default UserDetailSummary;
