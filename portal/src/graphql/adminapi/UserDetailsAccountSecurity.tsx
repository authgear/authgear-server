import React, { useMemo, useCallback, useContext } from "react";
import cn from "classnames";
import { List, PrimaryButton, Text } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ListCellLayout from "../../ListCellLayout";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAccountSecurity.module.scss";

// authenticator type recognized by portal
enum AuthenticatorType {
  Password = "password",
  OneTimePassword = "oob_otp",
  TimeBasedOneTimePassword = "totp",
}

type AuthenticatorKind = "primary" | "secondary";

interface Authenticator {
  id: string;
  type: string;
  is_secondary?: boolean;
  claims: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
}

interface UserDetailsAccountSecurityProps {
  authenticators: Authenticator[];
}

interface AuthenticatorListItem {
  label: React.ReactNode;
  description: React.ReactNode;
  descriptionClassName?: string;
  onDetailButtonClick?: () => void;
}

const primaryAuthenticatorTypeLocaleKeyMap: { [key: string]: string } = {
  [AuthenticatorType.Password]: "UserDetails.account-security.primary.password",
  [AuthenticatorType.OneTimePassword]:
    "UserDetails.account-security.primary.oob-otp",
};

const secondaryAuthenticatorTypeLocaleKeyMap: { [key: string]: string } = {
  [AuthenticatorType.Password]:
    "UserDetails.account-security.secondary.password",
  [AuthenticatorType.TimeBasedOneTimePassword]:
    "UserDetails.account-security.secondary.totp",
  [AuthenticatorType.OneTimePassword]:
    "UserDetails.account-security.secondary.oob-otp",
};

function getLocaleKeyWithAuthenticatorType(
  type: string,
  kind: AuthenticatorKind
) {
  switch (kind) {
    case "primary":
      return primaryAuthenticatorTypeLocaleKeyMap[type];
    case "secondary":
      return secondaryAuthenticatorTypeLocaleKeyMap[type];
  }
  return null;
}

function getAuthenticatorWithKind(
  authenticators: Authenticator[],
  kind: AuthenticatorKind
): Authenticator[] {
  switch (kind) {
    // TODO: add flag and modify this function
    case "primary":
      return authenticators.filter(
        (authenticator) => !authenticator.is_secondary
      );
    case "secondary":
      return authenticators.filter(
        (authenticator) => !!authenticator.is_secondary
      );
  }
  return [];
}

function getDescriptionFromAuthenticator(
  authenticator: Authenticator,
  locale: string
) {
  const showLastUpdated = authenticator.type === AuthenticatorType.Password;
  if (showLastUpdated) {
    const lastUpdatedText =
      formatDatetime(locale, authenticator.updatedAt) ?? "";
    return {
      description: (
        <FormattedMessage
          id="UserDetails.account-security.last-updated"
          values={{ datetime: lastUpdatedText }}
        />
      ),
    };
  }
  return {
    description: <FormattedMessage id="activated" />,
    descriptionClassName: styles.activated,
  };
}

function getLabelFromAuthenticator(authenticator: Authenticator) {
  const labelLocaleKey = getLocaleKeyWithAuthenticatorType(
    authenticator.type,
    "primary"
  );
  return labelLocaleKey ? <FormattedMessage id={labelLocaleKey} /> : null;
}

const AuthenticatorDetailListCell: React.FC<{
  item?: AuthenticatorListItem;
}> = function AuthenticatorDetailListCell(props: {
  item?: AuthenticatorListItem;
}) {
  const { item } = props;
  if (item == null) {
    return null;
  }
  return (
    <ListCellLayout className={styles.cell}>
      <Text className={styles.cellLabel}>{item.label}</Text>
      <Text className={cn(styles.cellDesc, item.descriptionClassName)}>
        {item.description}
      </Text>
      <PrimaryButton
        className={styles.detailsButton}
        onClick={item.onDetailButtonClick}
      >
        <FormattedMessage id="details" />
      </PrimaryButton>
    </ListCellLayout>
  );
};

const UserDetailsAccountSecurity: React.FC<UserDetailsAccountSecurityProps> = function UserDetailsAccountSecurity(
  props: UserDetailsAccountSecurityProps
) {
  const { authenticators } = props;
  const { locale } = useContext(Context);

  const primaryAuthenticators = getAuthenticatorWithKind(
    authenticators,
    "primary"
  );
  const secondaryAuthenticators = getAuthenticatorWithKind(
    authenticators,
    "secondary"
  );

  const primaryAuthenticatorListItems: AuthenticatorListItem[] = useMemo(() => {
    return primaryAuthenticators.map((authenticator) => {
      const label = getLabelFromAuthenticator(authenticator);
      const {
        description,
        descriptionClassName,
      } = getDescriptionFromAuthenticator(authenticator, locale);
      return {
        label,
        description,
        descriptionClassName,
      };
    });
  }, [locale, primaryAuthenticators]);

  const secondaryAuthenticatorListItems: AuthenticatorListItem[] = useMemo(() => {
    return secondaryAuthenticators.map((authenticator) => {
      const label = getLabelFromAuthenticator(authenticator);
      const {
        description,
        descriptionClassName,
      } = getDescriptionFromAuthenticator(authenticator, locale);
      return {
        label,
        description,
        descriptionClassName,
      };
    });
  }, [locale, secondaryAuthenticators]);

  const onRenderAuthenticatorDetailCell = useCallback(
    (item?: AuthenticatorListItem, _index?: number): React.ReactNode => {
      return <AuthenticatorDetailListCell item={item} />;
    },
    []
  );

  return (
    <div className={styles.root}>
      {primaryAuthenticatorListItems.length > 0 && (
        <div className={styles.authenticatorContainer}>
          <h1 className={styles.authenticatorHeader}>
            <FormattedMessage id="UserDetails.account-security.primary" />
          </h1>
          <List
            items={primaryAuthenticatorListItems}
            onRenderCell={onRenderAuthenticatorDetailCell}
          />
        </div>
      )}
      {secondaryAuthenticatorListItems.length > 0 && (
        <div className={styles.authenticatorContainer}>
          <h3 className={styles.authenticatorHeader}>
            <FormattedMessage id="UserDetails.account-security.secondary" />
          </h3>
          <List
            items={secondaryAuthenticatorListItems}
            onRenderCell={onRenderAuthenticatorDetailCell}
          />
        </div>
      )}
    </div>
  );
};

export default UserDetailsAccountSecurity;
