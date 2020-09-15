import React, { useMemo, useCallback, useContext } from "react";
import cn from "classnames";
import { List, PrimaryButton, Text } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ListCellLayout from "../../ListCellLayout";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAccountSecurity.module.scss";

// authenticator type recognized by portal
type AuthenticatorType = "PASSWORD" | "TOTP" | "OOB_OTP";

type AuthenticatorKind = "PRIMARY" | "SECONDARY";

interface Authenticator {
  id: string;
  type: AuthenticatorType;
  kind: AuthenticatorKind;
  isDefault: boolean;
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

const primaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "UserDetails.account-security.primary.password",
  OOB_OTP: "UserDetails.account-security.primary.oob-otp",
};

const secondaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "UserDetails.account-security.secondary.password",
  TOTP: "UserDetails.account-security.secondary.totp",
  OOB_OTP: "UserDetails.account-security.secondary.oob-otp",
};

function getLocaleKeyWithAuthenticatorType(
  type: AuthenticatorType,
  kind: AuthenticatorKind
): string | undefined {
  switch (kind) {
    case "PRIMARY":
      return primaryAuthenticatorTypeLocaleKeyMap[type];
    case "SECONDARY":
      return secondaryAuthenticatorTypeLocaleKeyMap[type];
    default:
      return undefined;
  }
}

function getDescriptionFromAuthenticator(
  authenticator: Authenticator,
  locale: string
) {
  const showLastUpdated = authenticator.type === "PASSWORD";
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
    authenticator.kind
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

  const primaryAuthenticators = authenticators.filter(
    (a) => a.kind === "PRIMARY"
  );
  const secondaryAuthenticators = authenticators.filter(
    (a) => a.kind === "SECONDARY"
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
          <Text as="h2" className={styles.authenticatorHeader}>
            <FormattedMessage id="UserDetails.account-security.primary" />
          </Text>
          <List
            items={primaryAuthenticatorListItems}
            onRenderCell={onRenderAuthenticatorDetailCell}
          />
        </div>
      )}
      {secondaryAuthenticatorListItems.length > 0 && (
        <div className={styles.authenticatorContainer}>
          <Text as="h2" className={styles.authenticatorHeader}>
            <FormattedMessage id="UserDetails.account-security.secondary" />
          </Text>
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
