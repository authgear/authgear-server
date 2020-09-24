import React, { useMemo, useCallback, useContext } from "react";
import { useNavigate } from "react-router-dom";
import cn from "classnames";
import {
  DefaultButton,
  Icon,
  List,
  PrimaryButton,
  Text,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ListCellLayout from "../../ListCellLayout";
import { destructiveTheme } from "../../theme";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAccountSecurity.module.scss";

// authenticator type recognized by portal
type PrimaryAuthenticatorType = "PASSWORD" | "OOB_OTP";
type SecondaryAuthenticatorType = "PASSWORD" | "TOTP" | "OOB_OTP";
type AuthenticatorType = PrimaryAuthenticatorType | SecondaryAuthenticatorType;

type AuthenticatorKind = "PRIMARY" | "SECONDARY";

type OOBOTPVerificationMethod = "email" | "phone" | "unknown";

interface AuthenticatorClaims extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
}

interface Authenticator {
  id: string;
  type: AuthenticatorType;
  kind: AuthenticatorKind;
  isDefault: boolean;
  claims: AuthenticatorClaims;
  createdAt: string;
  updatedAt: string;
}

interface UserDetailsAccountSecurityProps {
  authenticators: Authenticator[];
}

interface PasswordAuthenticatorData {
  id: string;
  kind: AuthenticatorKind;
  lastUpdated: string;
}

interface TOTPAuthenticatorData {
  id: string;
  kind: AuthenticatorKind;
  label: string;
  addedOn: string;
}

interface OOBOTPAuthenticatorData {
  id: string;
  iconName?: string;
  kind: AuthenticatorKind;
  label: string;
  addedOn: string;
  isDefault: boolean;
}

const LABEL_PLACEHOLDER = "---";

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

function constructPasswordAuthenticatorData(
  authenticator: Authenticator,
  locale: string
): PasswordAuthenticatorData {
  const lastUpdated = formatDatetime(locale, authenticator.updatedAt) ?? "";

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    lastUpdated,
  };
}

function getTotpDisplayName(
  totpAuthenticatorClaims: Record<string, unknown>
): string {
  for (const [key, claim] of Object.entries(totpAuthenticatorClaims)) {
    if (
      key === "https://authgear.com/claims/totp/display_name" &&
      typeof claim === "string"
    ) {
      return claim;
    }
  }
  return LABEL_PLACEHOLDER;
}

function constructTotpAuthenticatorData(
  authenticator: Authenticator,
  locale: string
): TOTPAuthenticatorData {
  const addedOn = formatDatetime(locale, authenticator.createdAt) ?? "";
  const label = getTotpDisplayName(authenticator.claims);

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    addedOn,
    label,
  };
}

function getOobOtpVerificationMethod(
  authenticator: Authenticator
): OOBOTPVerificationMethod {
  if (authenticator.claims.email != null) {
    return "email";
  }
  if (authenticator.claims.phone_number != null) {
    return "phone";
  }
  return "unknown";
}

const oobOtpVerificationMethodIconName: Partial<Record<
  OOBOTPVerificationMethod,
  string
>> = {
  email: "Mail",
  phone: "CellPhone",
};

function getOobOtpAuthenticatorLabel(
  authenticator: Authenticator,
  verificationMethod: OOBOTPVerificationMethod
) {
  switch (verificationMethod) {
    case "email":
      return authenticator.claims.email ?? "";
    case "phone":
      return authenticator.claims.phone_number ?? "";
    default:
      return "";
  }
}

function constructOobOtpAuthenticatorData(
  authenticator: Authenticator,
  locale: string
): OOBOTPAuthenticatorData {
  const addedOn = formatDatetime(locale, authenticator.createdAt) ?? "";
  const verificationMethod = getOobOtpVerificationMethod(authenticator);
  const iconName = oobOtpVerificationMethodIconName[verificationMethod];
  const label = getOobOtpAuthenticatorLabel(authenticator, verificationMethod);

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    isDefault: authenticator.isDefault,
    iconName,
    label,
    addedOn,
  };
}

function constructAuthenticatorLists(
  authenticators: Authenticator[],
  kind: AuthenticatorKind,
  locale: string
) {
  const passwordAuthenticatorList: PasswordAuthenticatorData[] = [];
  const oobOtpAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const totpAuthenticatorList: TOTPAuthenticatorData[] = [];

  const filteredAuthenticators = authenticators.filter((a) => a.kind === kind);

  for (const authenticator of filteredAuthenticators) {
    switch (authenticator.type) {
      case "PASSWORD":
        passwordAuthenticatorList.push(
          constructPasswordAuthenticatorData(authenticator, locale)
        );
        break;
      case "OOB_OTP":
        oobOtpAuthenticatorList.push(
          constructOobOtpAuthenticatorData(authenticator, locale)
        );
        break;
      case "TOTP":
        if (kind === "PRIMARY") {
          break;
        }
        totpAuthenticatorList.push(
          constructTotpAuthenticatorData(authenticator, locale)
        );
        break;
      default:
        break;
    }
  }

  return kind === "PRIMARY"
    ? {
        password: passwordAuthenticatorList,
        oobOtp: oobOtpAuthenticatorList,
        hasVisibleList: [
          passwordAuthenticatorList,
          oobOtpAuthenticatorList,
        ].some((list) => list.length > 0),
      }
    : {
        password: passwordAuthenticatorList,
        oobOtp: oobOtpAuthenticatorList,
        totp: totpAuthenticatorList,
        hasVisibleList: [
          passwordAuthenticatorList,
          oobOtpAuthenticatorList,
          totpAuthenticatorList,
        ].some((list) => list.length > 0),
      };
}

const PasswordAuthenticatorCell: React.FC<PasswordAuthenticatorData> = function PasswordAuthenticatorCell(
  props: PasswordAuthenticatorData
) {
  const { kind, lastUpdated } = props;
  const navigate = useNavigate();

  const labelId = getLocaleKeyWithAuthenticatorType("PASSWORD", kind);

  const onResetPasswordClicked = useCallback(() => {
    navigate("./reset-password");
  }, [navigate]);

  const onRemoveClicked = useCallback(() => {
    // TODO: implement mutation
  }, []);

  return (
    <ListCellLayout className={cn(styles.cell, styles.passwordCell)}>
      <Text className={cn(styles.cellLabel, styles.passwordCellLabel)}>
        <FormattedMessage id={labelId!} />
      </Text>
      <Text className={cn(styles.cellDesc, styles.passwordCellDesc)}>
        <FormattedMessage
          id="UserDetails.account-security.last-updated"
          values={{ datetime: lastUpdated }}
        />
      </Text>
      {kind === "PRIMARY" && (
        <PrimaryButton
          className={cn(styles.button, styles.resetPasswordButton)}
          onClick={onResetPasswordClicked}
        >
          <FormattedMessage id="UserDetails.account-security.reset-password" />
        </PrimaryButton>
      )}
      {kind === "SECONDARY" && (
        <DefaultButton
          className={cn(
            styles.button,
            styles.removeButton,
            styles.removePasswordButton
          )}
          onClick={onRemoveClicked}
          theme={destructiveTheme}
        >
          <FormattedMessage id="remove" />
        </DefaultButton>
      )}
    </ListCellLayout>
  );
};

const TOTPAuthenticatorCell: React.FC<TOTPAuthenticatorData> = function TOTPAuthenticatorCell(
  props: TOTPAuthenticatorData
) {
  const { kind, label, addedOn } = props;
  const onRemoveClicked = useCallback(() => {
    // TODO: implement mutation
  }, []);
  return (
    <ListCellLayout className={cn(styles.cell, styles.totpCell)}>
      <Text className={cn(styles.cellLabel, styles.totpCellLabel)}>
        {label}
      </Text>
      <Text className={cn(styles.cellDesc, styles.totpCellDesc)}>
        <FormattedMessage
          id="UserDetails.account-security.added-on"
          values={{ datetime: addedOn }}
        />
      </Text>
      {kind === "SECONDARY" && (
        <DefaultButton
          className={cn(
            styles.button,
            styles.removeButton,
            styles.totpRemoveButton
          )}
          onClick={onRemoveClicked}
          theme={destructiveTheme}
        >
          <FormattedMessage id="remove" />
        </DefaultButton>
      )}
    </ListCellLayout>
  );
};

const OOBOTPAuthenticatorCell: React.FC<OOBOTPAuthenticatorData> = function (
  props: OOBOTPAuthenticatorData
) {
  const { label, iconName, kind, addedOn } = props;

  const onRemoveClicked = useCallback(() => {
    // TODO: implement mutation
  }, []);

  return (
    <ListCellLayout className={cn(styles.cell, styles.oobOtpCell)}>
      <Icon className={styles.oobOtpCellIcon} iconName={iconName} />
      <Text className={cn(styles.cellLabel, styles.oobOtpCellLabel)}>
        {label}
      </Text>
      <Text className={cn(styles.cellDesc, styles.oobOtpCellAddedOn)}>
        <FormattedMessage
          id="UserDetails.account-security.added-on"
          values={{ datetime: addedOn }}
        />
      </Text>

      {kind === "SECONDARY" && (
        <DefaultButton
          className={cn(
            styles.button,
            styles.removeButton,
            styles.oobOtpRemoveButton
          )}
          onClick={onRemoveClicked}
          theme={destructiveTheme}
        >
          <FormattedMessage id="remove" />
        </DefaultButton>
      )}
    </ListCellLayout>
  );
};

const UserDetailsAccountSecurity: React.FC<UserDetailsAccountSecurityProps> = function UserDetailsAccountSecurity(
  props: UserDetailsAccountSecurityProps
) {
  const { authenticators } = props;
  const { locale } = useContext(Context);

  const primaryAuthenticatorLists = useMemo(() => {
    return constructAuthenticatorLists(authenticators, "PRIMARY", locale);
  }, [locale, authenticators]);

  const secondaryAuthenticatorLists = useMemo(() => {
    return constructAuthenticatorLists(authenticators, "SECONDARY", locale);
  }, [locale, authenticators]);

  const onRenderPasswordAuthenticatorDetailCell = useCallback(
    (item?: PasswordAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return <PasswordAuthenticatorCell {...item} />;
    },
    []
  );

  const onRenderOobOtpAuthenticatorDetailCell = useCallback(
    (item?: OOBOTPAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return <OOBOTPAuthenticatorCell {...item} />;
    },
    []
  );

  const onRenderTotpAuthenticatorDetailCell = useCallback(
    (item?: TOTPAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return <TOTPAuthenticatorCell {...item} />;
    },
    []
  );

  return (
    <div className={styles.root}>
      {primaryAuthenticatorLists.hasVisibleList && (
        <div className={styles.authenticatorContainer}>
          <Text
            as="h2"
            className={cn(styles.header, styles.authenticatorKindHeader)}
          >
            <FormattedMessage id="UserDetails.account-security.primary" />
          </Text>
          {primaryAuthenticatorLists.password.length > 0 && (
            <List
              className={styles.list}
              items={primaryAuthenticatorLists.password}
              onRenderCell={onRenderPasswordAuthenticatorDetailCell}
            />
          )}
          {primaryAuthenticatorLists.oobOtp.length > 0 && (
            <>
              <Text
                as="h3"
                className={cn(styles.header, styles.authenticatorTypeHeader)}
              >
                <FormattedMessage id="UserDetails.account-security.primary.oob-otp" />
              </Text>
              <List
                className={cn(styles.list, styles.oobOtpList)}
                items={primaryAuthenticatorLists.oobOtp}
                onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
              />
            </>
          )}
        </div>
      )}
      {secondaryAuthenticatorLists.hasVisibleList && (
        <div className={styles.authenticatorContainer}>
          <Text
            as="h2"
            className={cn(styles.header, styles.authenticatorKindHeader)}
          >
            <FormattedMessage id="UserDetails.account-security.secondary" />
          </Text>
          {secondaryAuthenticatorLists.totp != null &&
            secondaryAuthenticatorLists.totp.length > 0 && (
              <>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="UserDetails.account-security.secondary.totp" />
                </Text>
                <List
                  className={cn(styles.list, styles.totpList)}
                  items={secondaryAuthenticatorLists.totp}
                  onRenderCell={onRenderTotpAuthenticatorDetailCell}
                />
              </>
            )}
          {secondaryAuthenticatorLists.oobOtp.length > 0 && (
            <>
              <Text
                as="h3"
                className={cn(styles.header, styles.authenticatorTypeHeader)}
              >
                <FormattedMessage id="UserDetails.account-security.secondary.oob-otp" />
              </Text>
              <List
                className={cn(styles.list, styles.oobOtpList)}
                items={secondaryAuthenticatorLists.oobOtp}
                onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
              />
            </>
          )}
          {secondaryAuthenticatorLists.password.length > 0 && (
            <List
              className={cn(styles.list, styles.passwordList)}
              items={secondaryAuthenticatorLists.password}
              onRenderCell={onRenderPasswordAuthenticatorDetailCell}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default UserDetailsAccountSecurity;
