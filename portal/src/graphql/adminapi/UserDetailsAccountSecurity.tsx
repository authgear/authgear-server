import React, { useMemo, useCallback, useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import cn from "classnames";
import {
  Dialog,
  DialogFooter,
  IContextualMenuItem,
  IContextualMenuProps,
  Icon,
  List,
  Text,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import { useDeleteAuthenticatorMutation } from "./mutations/deleteAuthenticatorMutation";
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ListCellLayout from "../../ListCellLayout";
import ButtonWithLoading from "../../ButtonWithLoading";
import { formatDatetime } from "../../util/formatDatetime";
import {
  Identity,
  Authenticator,
  AuthenticatorType,
  AuthenticatorKind,
  IdentityType,
} from "./globalTypes.generated";
import { useProvideError } from "../../hook/error";
import styles from "./UserDetailsAccountSecurity.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { PortalAPIAppConfig, SecondaryAuthenticatorType } from "../../types";

type OOBOTPVerificationMethod = "email" | "phone" | "unknown";

interface UserDetailsAccountSecurityProps {
  authenticationConfig: PortalAPIAppConfig["authentication"];
  identities: Identity[];
  authenticators: Authenticator[];
}

interface PasskeyIdentityData {
  id: string;
  displayName: string;
  addedOn: string;
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

interface PasskeyIdentityCellProps extends PasskeyIdentityData {
  withTopSpacing: boolean;
  showConfirmationDialog: (options: RemoveConfirmationDialogData) => void;
}

interface PasswordAuthenticatorCellProps extends PasswordAuthenticatorData {
  withTopSpacing: boolean;
  showConfirmationDialog: (options: RemoveConfirmationDialogData) => void;
}

interface TOTPAuthenticatorCellProps extends TOTPAuthenticatorData {
  withTopSpacing: boolean;
  showConfirmationDialog: (options: RemoveConfirmationDialogData) => void;
}

interface OOBOTPAuthenticatorCellProps extends OOBOTPAuthenticatorData {
  withTopSpacing: boolean;
  showConfirmationDialog: (options: RemoveConfirmationDialogData) => void;
}

interface RemoveConfirmationDialogData {
  id: string;
  displayName: string;
  type: "identity" | "authenticator";
}

interface RemoveConfirmationDialogProps
  extends Partial<RemoveConfirmationDialogData> {
  visible: boolean;
  onDismiss: () => void;
  remove?: (id: string) => void;
  loading?: boolean;
}

interface Add2FAMenuItem extends IContextualMenuItem {
  key: SecondaryAuthenticatorType;
}

const LABEL_PLACEHOLDER = "---";

const primaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "AuthenticatorType.primary.password",
  OOB_OTP_EMAIL: "AuthenticatorType.primary.oob-otp-email",
  OOB_OTP_SMS: "AuthenticatorType.primary.oob-otp-phone",
};

const secondaryAuthenticatorTypeLocaleKeyMap: {
  [key in AuthenticatorType]?: string;
} = {
  PASSWORD: "AuthenticatorType.secondary.password",
  TOTP: "AuthenticatorType.secondary.totp",
  OOB_OTP_EMAIL: "AuthenticatorType.secondary.oob-otp-email",
  OOB_OTP_SMS: "AuthenticatorType.secondary.oob-otp-phone",
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

function constructPasskeyIdentityData(
  identity: Identity,
  locale: string
): PasskeyIdentityData {
  const addedOn = formatDatetime(locale, identity.createdAt) ?? "";

  return {
    id: identity.id,
    displayName: (identity.claims[
      "https://authgear.com/claims/passkey/display_name"
    ] ?? "") as string,
    addedOn,
  };
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
  totpAuthenticatorClaims: Authenticator["claims"]
): string {
  return (totpAuthenticatorClaims[
    "https://authgear.com/claims/totp/display_name"
  ] ?? LABEL_PLACEHOLDER) as string;
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
  switch (authenticator.type) {
    case "OOB_OTP_EMAIL":
      return "email";
    case "OOB_OTP_SMS":
      return "phone";
    default:
      return "unknown";
  }
}

const oobOtpVerificationMethodIconName: Partial<
  Record<OOBOTPVerificationMethod, string>
> = {
  email: "Mail",
  phone: "CellPhone",
};

function getOobOtpAuthenticatorLabel(
  authenticator: Authenticator,
  verificationMethod: OOBOTPVerificationMethod
): string {
  switch (verificationMethod) {
    case "email":
      return (authenticator.claims[
        "https://authgear.com/claims/oob_otp/email"
      ] ?? "") as string;
    case "phone":
      return (authenticator.claims[
        "https://authgear.com/claims/oob_otp/phone"
      ] ?? "") as string;
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

function constructSecondaryAuthenticatorList(
  config: PortalAPIAppConfig["authentication"],
  authenticators: Authenticator[],
  locale: string
) {
  const passwordAuthenticatorList: PasswordAuthenticatorData[] = [];
  const oobOtpEmailAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const oobOtpSMSAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const totpAuthenticatorList: TOTPAuthenticatorData[] = [];
  const isAnySecondaryAuthenticatorEnabled =
    (config?.secondary_authenticators?.length ?? 0) >= 1;
  const isSecondaryPasswordEnabled =
    config?.secondary_authenticators?.includes("password") ?? false;
  const isSecondaryOOBOTPEmailEnabled =
    config?.secondary_authenticators?.includes("oob_otp_email") ?? false;
  const isSecondaryOOBOTPSMSEnabled =
    config?.secondary_authenticators?.includes("oob_otp_sms") ?? false;

  const filteredAuthenticators = authenticators.filter(
    (a) => a.kind === AuthenticatorKind.Secondary
  );

  for (const authenticator of filteredAuthenticators) {
    switch (authenticator.type) {
      case "PASSWORD":
        passwordAuthenticatorList.push(
          constructPasswordAuthenticatorData(authenticator, locale)
        );
        break;
      case "OOB_OTP_EMAIL":
        oobOtpEmailAuthenticatorList.push(
          constructOobOtpAuthenticatorData(authenticator, locale)
        );
        break;
      case "OOB_OTP_SMS":
        oobOtpSMSAuthenticatorList.push(
          constructOobOtpAuthenticatorData(authenticator, locale)
        );
        break;
      case "TOTP":
        totpAuthenticatorList.push(
          constructTotpAuthenticatorData(authenticator, locale)
        );
        break;
      default:
        break;
    }
  }

  return {
    password: passwordAuthenticatorList,
    oobOtpEmail: oobOtpEmailAuthenticatorList,
    oobOtpSMS: oobOtpSMSAuthenticatorList,
    totp: totpAuthenticatorList,
    hasVisibleList: [
      passwordAuthenticatorList,
      oobOtpEmailAuthenticatorList,
      oobOtpSMSAuthenticatorList,
      totpAuthenticatorList,
    ].some((list) => list.length > 0),
    isAnySecondaryAuthenticatorEnabled,
    isSecondaryOOBOTPEmailEnabled,
    isSecondaryOOBOTPSMSEnabled,
    isSecondaryPasswordEnabled,
  };
}

function constructPrimaryAuthenticatorLists(
  identities: Identity[],
  authenticators: Authenticator[],
  locale: string
) {
  const passkeyIdentityList: PasskeyIdentityData[] = [];
  const passwordAuthenticatorList: PasswordAuthenticatorData[] = [];
  const oobOtpEmailAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const oobOtpSMSAuthenticatorList: OOBOTPAuthenticatorData[] = [];

  const filteredAuthenticators = authenticators.filter(
    (a) => a.kind === AuthenticatorKind.Primary
  );

  for (const identity of identities) {
    switch (identity.type) {
      case IdentityType.Passkey:
        passkeyIdentityList.push(
          constructPasskeyIdentityData(identity, locale)
        );
        break;
      default:
        break;
    }
  }

  for (const authenticator of filteredAuthenticators) {
    switch (authenticator.type) {
      case "PASSWORD":
        passwordAuthenticatorList.push(
          constructPasswordAuthenticatorData(authenticator, locale)
        );
        break;
      case "OOB_OTP_EMAIL":
        oobOtpEmailAuthenticatorList.push(
          constructOobOtpAuthenticatorData(authenticator, locale)
        );
        break;
      case "OOB_OTP_SMS":
        oobOtpSMSAuthenticatorList.push(
          constructOobOtpAuthenticatorData(authenticator, locale)
        );
        break;
      case "TOTP":
        break;
      default:
        break;
    }
  }

  return {
    passkey: passkeyIdentityList,
    password: passwordAuthenticatorList,
    oobOtpEmail: oobOtpEmailAuthenticatorList,
    oobOtpSMS: oobOtpSMSAuthenticatorList,
    hasVisibleList: [
      passkeyIdentityList,
      passwordAuthenticatorList,
      oobOtpEmailAuthenticatorList,
      oobOtpSMSAuthenticatorList,
    ].some((list) => list.length > 0),
  };
}

const RemoveConfirmationDialog: React.VFC<RemoveConfirmationDialogProps> =
  function RemoveConfirmationDialog(props: RemoveConfirmationDialogProps) {
    const {
      visible,
      remove,
      loading,
      id,
      displayName,
      onDismiss: onDismissProps,
    } = props;

    const { renderToString } = useContext(Context);

    const onConfirmClicked = useCallback(() => {
      remove?.(id!);
    }, [remove, id]);

    const onDismiss = useCallback(() => {
      if (!loading) {
        onDismissProps();
      }
    }, [onDismissProps, loading]);

    const dialogMessage = useMemo(() => {
      return renderToString(
        "UserDetails.account-security.remove-confirm-dialog.message",
        { displayName: displayName ?? "" }
      );
    }, [renderToString, displayName]);

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
        modalProps={{ isBlocking: loading }}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <ButtonWithLoading
            onClick={onConfirmClicked}
            labelId="confirm"
            loading={loading ?? false}
            disabled={!visible}
          />
          <DefaultButton
            disabled={(loading ?? false) || !visible}
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };

const PasskeyIdentityCell: React.VFC<PasskeyIdentityCellProps> =
  function PasskeyIdentityCell(props: PasskeyIdentityCellProps) {
    const { id, displayName, addedOn, showConfirmationDialog, withTopSpacing } =
      props;
    const { themes } = useSystemConfig();
    const onRemoveClicked = useCallback(() => {
      showConfirmationDialog({
        id,
        displayName,
        type: "identity",
      });
    }, [id, displayName, showConfirmationDialog]);
    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.passkeyCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
        <i
          className={cn(
            styles.passkeyCellIcon,
            "authgear-portal-icons authgear-portal-icons-passkey"
          )}
        ></i>
        <Text className={cn(styles.cellLabel, styles.passkeyCellLabel)}>
          {displayName}
        </Text>
        <Text className={cn(styles.cellDesc, styles.passkeyCellDesc)}>
          <FormattedMessage
            id="UserDetails.account-security.added-on"
            values={{ datetime: addedOn }}
          />
        </Text>
        <DefaultButton
          className={cn(styles.button, styles.passkeyCellRemoveButton)}
          onClick={onRemoveClicked}
          theme={themes.destructive}
          text={<FormattedMessage id="remove" />}
        />
      </ListCellLayout>
    );
  };

const PasswordAuthenticatorCell: React.VFC<PasswordAuthenticatorCellProps> =
  function PasswordAuthenticatorCell(props: PasswordAuthenticatorCellProps) {
    const { id, kind, lastUpdated, showConfirmationDialog, withTopSpacing } =
      props;
    const navigate = useNavigate();
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const labelId = getLocaleKeyWithAuthenticatorType(
      AuthenticatorType.Password,
      kind
    );

    const onResetPasswordClicked = useCallback(() => {
      navigate("./reset-password");
    }, [navigate]);

    const onRemoveClicked = useCallback(() => {
      showConfirmationDialog({
        id,
        displayName: renderToString(labelId!),
        type: "authenticator",
      });
    }, [labelId, id, renderToString, showConfirmationDialog]);

    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.passwordCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
        <Text className={cn(styles.cellLabel, styles.passwordCellLabel)}>
          <FormattedMessage id={labelId!} />
        </Text>
        <Text className={cn(styles.cellDesc, styles.passwordCellDesc)}>
          <FormattedMessage
            id="UserDetails.account-security.last-updated"
            values={{ datetime: lastUpdated }}
          />
        </Text>
        {kind === "PRIMARY" ? (
          <PrimaryButton
            className={cn(styles.button, styles.resetPasswordButton)}
            onClick={onResetPasswordClicked}
            text={
              <FormattedMessage id="UserDetails.account-security.reset-password" />
            }
          />
        ) : null}
        {kind === "SECONDARY" ? (
          <DefaultButton
            className={cn(styles.button, styles.removePasswordButton)}
            onClick={onRemoveClicked}
            theme={themes.destructive}
            text={<FormattedMessage id="remove" />}
          />
        ) : null}
      </ListCellLayout>
    );
  };

const TOTPAuthenticatorCell: React.VFC<TOTPAuthenticatorCellProps> =
  function TOTPAuthenticatorCell(props: TOTPAuthenticatorCellProps) {
    const { id, kind, label, addedOn, showConfirmationDialog, withTopSpacing } =
      props;
    const { themes } = useSystemConfig();

    const onRemoveClicked = useCallback(() => {
      showConfirmationDialog({
        id,
        displayName: label,
        type: "authenticator",
      });
    }, [id, label, showConfirmationDialog]);

    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.totpCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
        <Text className={cn(styles.cellLabel, styles.totpCellLabel)}>
          {label}
        </Text>
        <Text className={cn(styles.cellDesc, styles.totpCellDesc)}>
          <FormattedMessage
            id="UserDetails.account-security.added-on"
            values={{ datetime: addedOn }}
          />
        </Text>
        {kind === "SECONDARY" ? (
          <DefaultButton
            className={cn(styles.button, styles.totpRemoveButton)}
            onClick={onRemoveClicked}
            theme={themes.destructive}
            text={<FormattedMessage id="remove" />}
          />
        ) : null}
      </ListCellLayout>
    );
  };

const OOBOTPAuthenticatorCell: React.VFC<OOBOTPAuthenticatorCellProps> =
  function (props: OOBOTPAuthenticatorCellProps) {
    const {
      id,
      label,
      iconName,
      kind,
      addedOn,
      showConfirmationDialog,
      withTopSpacing,
    } = props;
    const { themes } = useSystemConfig();

    const onRemoveClicked = useCallback(() => {
      showConfirmationDialog({
        id,
        displayName: label,
        type: "authenticator",
      });
    }, [id, label, showConfirmationDialog]);

    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.oobOtpCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
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

        {kind === "SECONDARY" ? (
          <DefaultButton
            className={cn(styles.button, styles.oobOtpRemoveButton)}
            onClick={onRemoveClicked}
            theme={themes.destructive}
            text={<FormattedMessage id="remove" />}
          />
        ) : null}
      </ListCellLayout>
    );
  };

const UserDetailsAccountSecurity: React.VFC<UserDetailsAccountSecurityProps> =
  // eslint-disable-next-line complexity
  function UserDetailsAccountSecurity(props: UserDetailsAccountSecurityProps) {
    const { authenticationConfig, identities, authenticators } = props;
    const { locale, renderToString } = useContext(Context);
    const navigate = useNavigate();

    const {
      deleteAuthenticator,
      loading: deletingAuthenticator,
      error: deleteAuthenticatorError,
    } = useDeleteAuthenticatorMutation();
    useProvideError(deleteAuthenticatorError);

    const {
      deleteIdentity,
      loading: deletingIdentity,
      error: deleteIdentityError,
    } = useDeleteIdentityMutation();
    useProvideError(deleteIdentityError);

    const [isConfirmationDialogVisible, setIsConfirmationDialogVisible] =
      useState(false);
    const [confirmationDialogData, setConfirmationDialogData] =
      useState<RemoveConfirmationDialogData | null>(null);

    const primaryAuthenticatorLists = useMemo(() => {
      return constructPrimaryAuthenticatorLists(
        identities,
        authenticators,
        locale
      );
    }, [locale, identities, authenticators]);

    const secondaryAuthenticatorLists = useMemo(() => {
      return constructSecondaryAuthenticatorList(
        authenticationConfig,
        authenticators,
        locale
      );
    }, [authenticationConfig, authenticators, locale]);

    const showConfirmationDialog = useCallback(
      (options: RemoveConfirmationDialogData) => {
        setConfirmationDialogData(options);
        setIsConfirmationDialogVisible(true);
      },
      []
    );

    const dismissConfirmationDialog = useCallback(() => {
      setIsConfirmationDialogVisible(false);
    }, []);

    const onRenderPasskeyIdentityDetailCell = useCallback(
      (item?: PasskeyIdentityData, index?: number): React.ReactNode => {
        if (item == null) {
          return null;
        }
        return (
          <PasskeyIdentityCell
            {...item}
            withTopSpacing={index !== 0}
            showConfirmationDialog={showConfirmationDialog}
          />
        );
      },
      [showConfirmationDialog]
    );

    const onRenderPasswordAuthenticatorDetailCell = useCallback(
      (item?: PasswordAuthenticatorData, index?: number): React.ReactNode => {
        if (item == null) {
          return null;
        }
        return (
          <PasswordAuthenticatorCell
            {...item}
            withTopSpacing={index !== 0}
            showConfirmationDialog={showConfirmationDialog}
          />
        );
      },
      [showConfirmationDialog]
    );

    const onRenderOobOtpAuthenticatorDetailCell = useCallback(
      (item?: OOBOTPAuthenticatorData, index?: number): React.ReactNode => {
        if (item == null) {
          return null;
        }
        return (
          <OOBOTPAuthenticatorCell
            {...item}
            withTopSpacing={index !== 0}
            showConfirmationDialog={showConfirmationDialog}
          />
        );
      },
      [showConfirmationDialog]
    );

    const onRenderTotpAuthenticatorDetailCell = useCallback(
      (item?: TOTPAuthenticatorData, index?: number): React.ReactNode => {
        if (item == null) {
          return null;
        }
        return (
          <TOTPAuthenticatorCell
            {...item}
            withTopSpacing={index !== 0}
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

    const onConfirmDeleteIdentity = useCallback(
      (identityID) => {
        deleteIdentity(identityID)
          .catch(() => {})
          .finally(() => {
            dismissConfirmationDialog();
          });
      },
      [deleteIdentity, dismissConfirmationDialog]
    );

    const add2FAMenuProps: IContextualMenuProps = useMemo(() => {
      const availableMenuItem: Add2FAMenuItem[] = [
        {
          key: "password",
          text: renderToString("AuthenticatorType.secondary.password"),
          iconProps: { iconName: "Accounts" },
          onClick: () => navigate("./add-2fa-password"),
        },
        {
          key: "oob_otp_email",
          text: renderToString("AuthenticatorType.secondary.oob-otp-email"),
          iconProps: { iconName: "Mail" },
          onClick: () => navigate("./add-2fa-email"),
        },
        {
          key: "oob_otp_sms",
          text: renderToString("AuthenticatorType.secondary.oob-otp-phone"),
          iconProps: { iconName: "CellPhone" },
          onClick: () => navigate("./add-2fa-phone"),
        },
      ];
      const enabledItems = availableMenuItem.filter((item) => {
        if (
          !authenticationConfig?.secondary_authenticators?.includes(item.key)
        ) {
          return false;
        }

        if (item.key === "password") {
          // Multiple additinal password is not allowed
          if (
            authenticators.findIndex(
              (authn) =>
                authn.kind === AuthenticatorKind.Secondary &&
                authn.type === AuthenticatorType.Password
            ) !== -1
          ) {
            return false;
          }
        }

        return true;
      });
      return {
        items: enabledItems,
        directionalHintFixed: true,
      };
    }, [
      renderToString,
      navigate,
      authenticationConfig?.secondary_authenticators,
      authenticators,
    ]);

    return (
      <div className={styles.root}>
        <RemoveConfirmationDialog
          visible={isConfirmationDialogVisible}
          id={confirmationDialogData?.id}
          displayName={confirmationDialogData?.displayName}
          remove={
            confirmationDialogData?.type === "authenticator"
              ? onConfirmDeleteAuthenticator
              : confirmationDialogData?.type === "identity"
              ? onConfirmDeleteIdentity
              : undefined
          }
          loading={
            confirmationDialogData?.type === "authenticator"
              ? deletingAuthenticator
              : confirmationDialogData?.type === "identity"
              ? deletingIdentity
              : undefined
          }
          onDismiss={dismissConfirmationDialog}
        />
        {primaryAuthenticatorLists.hasVisibleList ? (
          <div className={styles.authenticatorContainer}>
            <Text
              as="h2"
              variant="medium"
              className={cn(styles.header, styles.authenticatorKindHeader)}
            >
              <FormattedMessage id="UserDetails.account-security.primary" />
            </Text>
            {primaryAuthenticatorLists.password.length > 0 ? (
              <List
                className={cn(
                  styles.authenticatorTypeSection,
                  styles["authenticatorTypeSection--password"]
                )}
                items={primaryAuthenticatorLists.password}
                onRenderCell={onRenderPasswordAuthenticatorDetailCell}
              />
            ) : null}
            {primaryAuthenticatorLists.passkey.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.primary.passkey" />
                </Text>
                <List
                  items={primaryAuthenticatorLists.passkey}
                  onRenderCell={onRenderPasskeyIdentityDetailCell}
                />
              </div>
            ) : null}
            {primaryAuthenticatorLists.oobOtpEmail.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.primary.oob-otp-email" />
                </Text>
                <List
                  items={primaryAuthenticatorLists.oobOtpEmail}
                  onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
                />
              </div>
            ) : null}
            {primaryAuthenticatorLists.oobOtpSMS.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.primary.oob-otp-phone" />
                </Text>
                <List
                  items={primaryAuthenticatorLists.oobOtpSMS}
                  onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
                />
              </div>
            ) : null}
          </div>
        ) : null}
        {secondaryAuthenticatorLists.hasVisibleList ||
        secondaryAuthenticatorLists.isAnySecondaryAuthenticatorEnabled ? (
          <div className={styles.authenticatorContainer}>
            <div
              className={cn(
                "flex justify-between",
                styles.authenticatorKindHeader
              )}
            >
              <Text as="h2" className={cn(styles.header)}>
                <FormattedMessage id="UserDetails.account-security.secondary" />
              </Text>
              <PrimaryButton
                disabled={add2FAMenuProps.items.length === 0}
                iconProps={{ iconName: "CirclePlus" }}
                menuProps={add2FAMenuProps}
                styles={{
                  menuIcon: { paddingLeft: "3px" },
                  icon: { paddingRight: "3px" },
                }}
                text={
                  <FormattedMessage id="UserDetails.account-security.secondary.add" />
                }
              />
            </div>
            {!secondaryAuthenticatorLists.hasVisibleList ? (
              <Text as="h3" className={cn(styles.authenticatorEmpty)}>
                <FormattedMessage id="UserDetails.account-security.secondary.empty" />
              </Text>
            ) : null}
            {secondaryAuthenticatorLists.totp.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.secondary.totp" />
                </Text>
                <List
                  items={secondaryAuthenticatorLists.totp}
                  onRenderCell={onRenderTotpAuthenticatorDetailCell}
                />
              </div>
            ) : null}
            {secondaryAuthenticatorLists.oobOtpEmail.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.secondary.oob-otp-email" />
                </Text>
                <List
                  items={secondaryAuthenticatorLists.oobOtpEmail}
                  onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
                />
              </div>
            ) : null}
            {secondaryAuthenticatorLists.oobOtpSMS.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="h3"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.secondary.oob-otp-phone" />
                </Text>
                <List
                  items={secondaryAuthenticatorLists.oobOtpSMS}
                  onRenderCell={onRenderOobOtpAuthenticatorDetailCell}
                />
              </div>
            ) : null}
            {secondaryAuthenticatorLists.password.length > 0 ? (
              <List
                className={cn(
                  styles.authenticatorTypeSection,
                  styles["authenticatorTypeSection--password"]
                )}
                items={secondaryAuthenticatorLists.password}
                onRenderCell={onRenderPasswordAuthenticatorDetailCell}
              />
            ) : null}
          </div>
        ) : null}
      </div>
    );
  };

export default UserDetailsAccountSecurity;
