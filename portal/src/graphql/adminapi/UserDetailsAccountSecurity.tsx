import React, { useMemo, useCallback, useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import cn from "classnames";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  Icon,
  List,
  PrimaryButton,
  Text,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import { useDeleteAuthenticatorMutation } from "./mutations/deleteAuthenticatorMutation";
import ListCellLayout from "../../ListCellLayout";
import ButtonWithLoading from "../../ButtonWithLoading";
import ErrorDialog from "../../error/ErrorDialog";
import { destructiveTheme } from "../../theme";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsAccountSecurity.module.scss";

// authenticator type recognized by portal
type PrimaryAuthenticatorType = "PASSWORD" | "OOB_OTP";
type SecondaryAuthenticatorType = "PASSWORD" | "TOTP" | "OOB_OTP";
type AuthenticatorType = PrimaryAuthenticatorType | SecondaryAuthenticatorType;

type AuthenticatorKind = "PRIMARY" | "SECONDARY";

type OOBOTPVerificationMethod = "email" | "phone" | "unknown";

interface AuthenticatorClaims {
  "https://authgear.com/claims/totp/display_name"?: string;
  "https://authgear.com/claims/oob_otp/channel_type"?: string;
  "https://authgear.com/claims/oob_otp/email"?: string;
  "https://authgear.com/claims/oob_otp/phone"?: string;
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

interface PasswordAuthenticatorCellProps extends PasswordAuthenticatorData {
  showConfirmationDialog: (
    authenticatorID: string,
    authenticatorName: string
  ) => void;
}

interface TOTPAuthenticatorCellProps extends TOTPAuthenticatorData {
  showConfirmationDialog: (
    authenticatorID: string,
    authenticatorName: string
  ) => void;
}

interface OOBOTPAuthenticatorCellProps extends OOBOTPAuthenticatorData {
  showConfirmationDialog: (
    authenticatorID: string,
    authenticatorName: string
  ) => void;
}

interface RemoveConfirmationDialogData {
  authenticatorID: string;
  authenticatorName: string;
}

interface RemoveConfirmationDialogProps
  extends Partial<RemoveConfirmationDialogData> {
  visible: boolean;
  deleteAuthenticator: (authenticatorID: string) => void;
  deletingAuthenticator: boolean;
  onDismiss: () => void;
}

const LABEL_PLACEHOLDER = "---";

const primaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "AuthenticatorType.primary.password",
  OOB_OTP: "AuthenticatorType.primary.oob-otp",
};

const secondaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "AuthenticatorType.secondary.password",
  TOTP: "AuthenticatorType.secondary.totp",
  OOB_OTP: "AuthenticatorType.secondary.oob-otp",
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
  totpAuthenticatorClaims: AuthenticatorClaims
): string {
  return (
    totpAuthenticatorClaims["https://authgear.com/claims/totp/display_name"] ??
    LABEL_PLACEHOLDER
  );
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
  switch (
    authenticator.claims["https://authgear.com/claims/oob_otp/channel_type"]
  ) {
    case "email":
      return "email";
    case "phone":
      return "phone";
    default:
      return "unknown";
  }
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
      return (
        authenticator.claims["https://authgear.com/claims/oob_otp/email"] ?? ""
      );
    case "phone":
      return (
        authenticator.claims["https://authgear.com/claims/oob_otp/phone"] ?? ""
      );
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

const RemoveConfirmationDialog: React.FC<RemoveConfirmationDialogProps> = function RemoveConfirmationDialog(
  props: RemoveConfirmationDialogProps
) {
  const {
    visible,
    deleteAuthenticator,
    deletingAuthenticator,
    authenticatorID,
    authenticatorName,
    onDismiss: onDismissProps,
  } = props;

  const { renderToString } = useContext(Context);

  const onConfirmClicked = useCallback(() => {
    deleteAuthenticator(authenticatorID!);
  }, [deleteAuthenticator, authenticatorID]);

  const onDismiss = useCallback(() => {
    if (!deletingAuthenticator) {
      onDismissProps();
    }
  }, [onDismissProps, deletingAuthenticator]);

  const dialogMessage = useMemo(() => {
    return renderToString(
      "UserDetails.account-security.remove-confirm-dialog.message",
      { authenticatorName: authenticatorName ?? "" }
    );
  }, [renderToString, authenticatorName]);

  const removeConfirmDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="UserDetails.account-security.remove-confirm-dialog.title" />
      ),
      subText: dialogMessage,
    };
  }, [dialogMessage]);

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={removeConfirmDialogContentProps}
      modalProps={{ isBlocking: deletingAuthenticator }}
      onDismiss={onDismiss}
    >
      <DialogFooter>
        <ButtonWithLoading
          onClick={onConfirmClicked}
          labelId="confirm"
          loading={deletingAuthenticator}
          disabled={!visible}
        />
        <DefaultButton
          disabled={deletingAuthenticator || !visible}
          onClick={onDismiss}
        >
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

const PasswordAuthenticatorCell: React.FC<PasswordAuthenticatorCellProps> = function PasswordAuthenticatorCell(
  props: PasswordAuthenticatorCellProps
) {
  const { id, kind, lastUpdated, showConfirmationDialog } = props;
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const labelId = getLocaleKeyWithAuthenticatorType("PASSWORD", kind);

  const onResetPasswordClicked = useCallback(() => {
    navigate("./reset-password");
  }, [navigate]);

  const onRemoveClicked = useCallback(() => {
    showConfirmationDialog(id, renderToString(labelId!));
  }, [labelId, id, renderToString, showConfirmationDialog]);

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

const TOTPAuthenticatorCell: React.FC<TOTPAuthenticatorCellProps> = function TOTPAuthenticatorCell(
  props: TOTPAuthenticatorCellProps
) {
  const { id, kind, label, addedOn, showConfirmationDialog } = props;

  const onRemoveClicked = useCallback(() => {
    showConfirmationDialog(id, label);
  }, [id, label, showConfirmationDialog]);

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

const OOBOTPAuthenticatorCell: React.FC<OOBOTPAuthenticatorCellProps> = function (
  props: OOBOTPAuthenticatorCellProps
) {
  const { id, label, iconName, kind, addedOn, showConfirmationDialog } = props;

  const onRemoveClicked = useCallback(() => {
    showConfirmationDialog(id, label);
  }, [id, label, showConfirmationDialog]);

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

  const {
    deleteAuthenticator,
    loading: deletingAuthenticator,
    error: deleteAuthenticatorError,
  } = useDeleteAuthenticatorMutation();

  const [
    isConfirmationDialogVisible,
    setIsConfirmationDialogVisible,
  ] = useState(false);
  const [
    confirmationDialogData,
    setConfirmationDialogData,
  ] = useState<RemoveConfirmationDialogData | null>(null);

  const primaryAuthenticatorLists = useMemo(() => {
    return constructAuthenticatorLists(authenticators, "PRIMARY", locale);
  }, [locale, authenticators]);

  const secondaryAuthenticatorLists = useMemo(() => {
    return constructAuthenticatorLists(authenticators, "SECONDARY", locale);
  }, [locale, authenticators]);

  const showConfirmationDialog = useCallback(
    (authenticatorID: string, authenticatorName: string) => {
      setConfirmationDialogData({
        authenticatorID,
        authenticatorName,
      });
      setIsConfirmationDialogVisible(true);
    },
    []
  );

  const dismissConfirmationDialog = useCallback(() => {
    setIsConfirmationDialogVisible(false);
  }, []);

  const onRenderPasswordAuthenticatorDetailCell = useCallback(
    (item?: PasswordAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <PasswordAuthenticatorCell
          {...item}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    },
    [showConfirmationDialog]
  );

  const onRenderOobOtpAuthenticatorDetailCell = useCallback(
    (item?: OOBOTPAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <OOBOTPAuthenticatorCell
          {...item}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    },
    [showConfirmationDialog]
  );

  const onRenderTotpAuthenticatorDetailCell = useCallback(
    (item?: TOTPAuthenticatorData, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }
      return (
        <TOTPAuthenticatorCell
          {...item}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    },
    [showConfirmationDialog]
  );

  const onConfirmDeleteAuthenticator = useCallback(
    (authenticatorID) => {
      deleteAuthenticator(authenticatorID)
        .catch(() => {})
        .finally(() => {
          dismissConfirmationDialog();
        });
    },
    [deleteAuthenticator, dismissConfirmationDialog]
  );

  return (
    <div className={styles.root}>
      <RemoveConfirmationDialog
        visible={isConfirmationDialogVisible}
        authenticatorID={confirmationDialogData?.authenticatorID}
        authenticatorName={confirmationDialogData?.authenticatorName}
        onDismiss={dismissConfirmationDialog}
        deleteAuthenticator={onConfirmDeleteAuthenticator}
        deletingAuthenticator={deletingAuthenticator}
      />
      <ErrorDialog
        rules={[]}
        error={deleteAuthenticatorError}
        fallbackErrorMessageID="UserDetails.account-security.remove-authenticator.generic-error"
      />
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
                <FormattedMessage id="AuthenticatorType.primary.oob-otp" />
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
                  <FormattedMessage id="AuthenticatorType.secondary.totp" />
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
                <FormattedMessage id="AuthenticatorType.secondary.oob-otp" />
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
