import React, { useMemo, useCallback, useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import cn from "classnames";
import {
  Badge,
  Button,
  DropdownMenu,
  IconButton,
  Text,
} from "@radix-ui/themes";
import {
  CheckCircledIcon,
  DotsVerticalIcon,
  EnvelopeClosedIcon,
  MobileIcon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { FormattedMessage, Context } from "../../intl";

import { useDeleteAuthenticatorMutation } from "./mutations/deleteAuthenticatorMutation";
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import ListCellLayout from "../../ListCellLayout";
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
import { PortalAPIAppConfig } from "../../types";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { DateTime } from "luxon";
import { parseDuration } from "../../util/duration";
import { SetPasswordExpiredConfirmationDialog } from "../../components/users/SetPasswordExpiredConfirmationDialog";
import { useConfirmationDialog } from "../../hook/useConfirmationDialog";
import { useSetPasswordExpiredMutation } from "./mutations/setPasswordExpiredMutation";
import LinkButton from "../../LinkButton";
import { useUserQuery } from "./query/userQuery";
import { parseDate } from "../../util/date";
import {
  MFAGracePeriodAction,
  SetMFAGracePeriodConfirmationDialog,
} from "../../components/users/SetMFAGracePeriodConfirmationDialog";
import { useSetMFAGracePeriodMutation } from "./mutations/setMFAGracePeriodMutation";
import { useRemoveMFAGracePeriodMutation } from "./mutations/removeMFAGracePeriodMutation";
import { CancelMFAGracePeriodConfirmationDialog } from "../../components/users/CancelMFAGracePeriodConfirmationDialog";
import { Add2FAPhoneDialog } from "../../components/users/Add2FAPhoneDialog";
import { Add2FAEmailDialog } from "../../components/users/Add2FAEmailDialog";
import { Add2FAPasswordDialog } from "../../components/users/Add2FAPasswordDialog";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Callout } from "../../components/v2/Callout/Callout";

type OOBOTPVerificationMethod = "email" | "phone" | "unknown";

interface UserDetailsAccountSecurityProps {
  userID: string;
  authenticationConfig: PortalAPIAppConfig["authentication"];
  authenticatorConfig: PortalAPIAppConfig["authenticator"];
  identities: Identity[];
  authenticators: Authenticator[];
  phoneInputAllowlist?: string[];
  phoneInputPinnedList?: string[];
  onAuthenticatorCreated?: () => unknown;
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
  lastUpdatedInDays: number;
  manualChangeOnLogin: boolean;
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
  forceChangeDaysSinceLastUpdate: number | null;
  showRemoveConfirmationDialog: (options: RemoveConfirmationDialogData) => void;
  showMarkAsExpiredConfirmationDialog: () => void;
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

const LABEL_PLACEHOLDER = "---";

const primaryAuthenticatorTypeLocaleKeyMap: Partial<
  Record<AuthenticatorType, string>
> = {
  PASSWORD: "AuthenticatorType.primary.password",
  OOB_OTP_EMAIL: "AuthenticatorType.primary.oob-otp-email",
  OOB_OTP_SMS: "AuthenticatorType.primary.oob-otp-phone",
};

const secondaryAuthenticatorTypeLocaleKeyMap: Partial<
  Record<AuthenticatorType, string>
> = {
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
  const manualChangeOnLogin = authenticator.expireAfter
    ? DateTime.fromISO(authenticator.expireAfter) <= DateTime.utc()
    : false;
  const lastUpdatedInDays = Math.round(
    DateTime.now().diff(DateTime.fromISO(authenticator.updatedAt), "days").days
  );

  return {
    id: authenticator.id,
    kind: authenticator.kind,
    lastUpdated,
    lastUpdatedInDays,
    manualChangeOnLogin,
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

  const enabledPasswordAuthenticatorList: PasswordAuthenticatorData[] = [];
  const enabledOobOtpEmailAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const enabledOobOtpSMSAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const enabledTotpAuthenticatorList: TOTPAuthenticatorData[] = [];

  const isAnySecondaryAuthenticatorEnabled =
    (config?.secondary_authenticators?.length ?? 0) >= 1;
  const isSecondaryPasswordEnabled =
    config?.secondary_authenticators?.includes("password") ?? false;
  const isSecondaryOOBOTPEmailEnabled =
    config?.secondary_authenticators?.includes("oob_otp_email") ?? false;
  const isSecondaryOOBOTPSMSEnabled =
    config?.secondary_authenticators?.includes("oob_otp_sms") ?? false;
  const isSecondaryTOTPEnabled =
    config?.secondary_authenticators?.includes("totp") ?? false;

  const filteredAuthenticators = authenticators.filter(
    (a) => a.kind === AuthenticatorKind.Secondary
  );

  for (const authenticator of filteredAuthenticators) {
    switch (authenticator.type) {
      case "PASSWORD": {
        const passwordData = constructPasswordAuthenticatorData(
          authenticator,
          locale
        );
        passwordAuthenticatorList.push(passwordData);
        if (isSecondaryPasswordEnabled) {
          enabledPasswordAuthenticatorList.push(passwordData);
        }
        break;
      }
      case "OOB_OTP_EMAIL": {
        const oobOtpEmailData = constructOobOtpAuthenticatorData(
          authenticator,
          locale
        );
        oobOtpEmailAuthenticatorList.push(oobOtpEmailData);
        if (isSecondaryOOBOTPEmailEnabled) {
          enabledOobOtpEmailAuthenticatorList.push(oobOtpEmailData);
        }
        break;
      }
      case "OOB_OTP_SMS": {
        const oobOtpSmsData = constructOobOtpAuthenticatorData(
          authenticator,
          locale
        );
        oobOtpSMSAuthenticatorList.push(oobOtpSmsData);
        if (isSecondaryOOBOTPSMSEnabled) {
          enabledOobOtpSMSAuthenticatorList.push(oobOtpSmsData);
        }
        break;
      }
      case "TOTP": {
        const totpData = constructTotpAuthenticatorData(authenticator, locale);
        totpAuthenticatorList.push(totpData);
        if (isSecondaryTOTPEnabled) {
          enabledTotpAuthenticatorList.push(totpData);
        }
        break;
      }
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
    isSecondaryTOTPEnabled,
    hasEnabledAuthenticator: [
      enabledPasswordAuthenticatorList,
      enabledOobOtpEmailAuthenticatorList,
      enabledOobOtpSMSAuthenticatorList,
      enabledTotpAuthenticatorList,
    ].some((list) => list.length > 0),
  };
}

function constructPrimaryAuthenticatorLists(
  config: PortalAPIAppConfig["authentication"],
  identities: Identity[],
  authenticators: Authenticator[],
  locale: string
) {
  const passkeyIdentityList: PasskeyIdentityData[] = [];
  const passwordAuthenticatorList: PasswordAuthenticatorData[] = [];
  const oobOtpEmailAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const oobOtpSMSAuthenticatorList: OOBOTPAuthenticatorData[] = [];
  const isPrimaryPasswordEnabled =
    config?.primary_authenticators?.includes("password") ?? false;

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
    isPrimaryPasswordEnabled,
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

    return (
      <ConfirmationDialog
        open={visible}
        onOpenChange={(open) => {
          if (!open) {
            onDismiss();
          }
        }}
        title={
          <FormattedMessage id="UserDetails.account-security.remove-confirm-dialog.title" />
        }
        description={dialogMessage}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onDismiss}
        loading={loading}
        confirmColor="red"
      />
    );
  };

interface SecondaryAuthenticatorTableCellProps {
  id: string;
  label: string;
  addedOn: string;
  showConfirmationDialog: (options: RemoveConfirmationDialogData) => void;
}

const SecondaryAuthenticatorTableCell: React.VFC<SecondaryAuthenticatorTableCellProps> =
  function SecondaryAuthenticatorTableCell(
    props: SecondaryAuthenticatorTableCellProps
  ) {
    const { id, label, addedOn, showConfirmationDialog } = props;
    const { renderToString } = useContext(Context);

    const onRemoveClicked = useCallback(() => {
      showConfirmationDialog({
        id,
        displayName: label,
        type: "authenticator",
      });
    }, [id, label, showConfirmationDialog]);

    return (
      <ListCellLayout className={styles.authenticatorTableCell}>
        <div className={styles.authenticatorTableValue}>
          <Text size="2">{label}</Text>
        </div>
        <Text size="2" color="gray" className={styles.authenticatorTableDate}>
          <FormattedMessage
            id="UserDetails.connected-identities.added-on"
            values={{ datetime: addedOn }}
          />
        </Text>
        <div className={styles.actionButton}>
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton
                className={styles.rowActionsButton}
                size="2"
                variant="soft"
                color="gray"
                aria-label={renderToString("action")}
              >
                <DotsVerticalIcon width="1rem" height="1rem" />
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              <DropdownMenu.Item color="red" onSelect={onRemoveClicked}>
                <TrashIcon />
                <FormattedMessage id="remove" />
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>
      </ListCellLayout>
    );
  };

const PasskeyIdentityCell: React.VFC<PasskeyIdentityCellProps> =
  function PasskeyIdentityCell(props: PasskeyIdentityCellProps) {
    const { id, displayName, addedOn, showConfirmationDialog, withTopSpacing } =
      props;
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
        <Text
          size="2"
          weight="medium"
          className={cn(styles.cellLabel, styles.passkeyCellLabel)}
        >
          {displayName}
        </Text>
        <Text
          size="2"
          color="gray"
          className={cn(styles.cellDesc, styles.passkeyCellDesc)}
        >
          <FormattedMessage
            id="UserDetails.account-security.added-on"
            values={{ datetime: addedOn }}
          />
        </Text>
        <Button
          className={cn(styles.button, styles.passkeyCellRemoveButton)}
          size="1"
          variant="outline"
          color="red"
          onClick={onRemoveClicked}
        >
          <FormattedMessage id="remove" />
        </Button>
      </ListCellLayout>
    );
  };

const PasswordAuthenticatorCell: React.VFC<PasswordAuthenticatorCellProps> =
  function PasswordAuthenticatorCell(props: PasswordAuthenticatorCellProps) {
    const {
      id,
      kind,
      lastUpdated,
      lastUpdatedInDays,
      manualChangeOnLogin,
      forceChangeDaysSinceLastUpdate,
      showRemoveConfirmationDialog,
      showMarkAsExpiredConfirmationDialog,
      withTopSpacing,
    } = props;
    const navigate = useNavigate();
    const { renderToString } = useContext(Context);

    const labelId = getLocaleKeyWithAuthenticatorType(
      AuthenticatorType.Password,
      kind
    );

    const passwordExpired =
      forceChangeDaysSinceLastUpdate != null &&
      lastUpdatedInDays > forceChangeDaysSinceLastUpdate;
    const changeOnLogin = manualChangeOnLogin || passwordExpired;

    const expiredInDays = useMemo(() => {
      if (!forceChangeDaysSinceLastUpdate) {
        return 0;
      }

      return lastUpdatedInDays - forceChangeDaysSinceLastUpdate;
    }, [forceChangeDaysSinceLastUpdate, lastUpdatedInDays]);

    const onResetPasswordClicked = useCallback(() => {
      navigate("./change-password");
    }, [navigate]);

    const onRemoveClicked = useCallback(() => {
      showRemoveConfirmationDialog({
        id,
        displayName: renderToString(labelId!),
        type: "authenticator",
      });
    }, [labelId, id, renderToString, showRemoveConfirmationDialog]);

    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.passwordCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
        <Text
          size="3"
          weight="medium"
          className={cn(styles.cellLabel, styles.passwordCellLabel)}
        >
          <FormattedMessage id={labelId!} />
        </Text>
        <div className={styles.passwordCellSummary}>
          <div className={styles.passwordCellMetadata}>
            <Badge size="1" color="green" variant="outline">
              <CheckCircledIcon width={12} height={12} />
              <FormattedMessage id="UserDetails.account-security.password-set" />
            </Badge>
            <Text
              size="2"
              className={cn(styles.cellDesc, styles.passwordCellDesc)}
            >
              <FormattedMessage
                id="UserDetails.account-security.last-updated"
                values={{ datetime: lastUpdated }}
              />
            </Text>
          </div>
          {kind === "PRIMARY" ? (
            <Button
              className={cn(styles.button, styles.changePasswordButton)}
              size="2"
              variant="ghost"
              onClick={onResetPasswordClicked}
            >
              <FormattedMessage id="UserDetails.account-security.change-password" />
            </Button>
          ) : null}
        </div>
        {kind === "PRIMARY" ? (
          <div className={styles.passwordCellChangeOnLogin}>
            <div className={styles.passwordCellChangeOnLoginText}>
              <Text
                as="p"
                size="2"
                weight="medium"
                className={styles.passwordCellChangeOnLoginTitle}
              >
                <FormattedMessage id="UserDetails.account-security.change-on-login.label" />
              </Text>
              <Text
                as="p"
                size="2"
                color="gray"
                className={styles.passwordCellChangeOnLoginDescription}
              >
                <FormattedMessage id="UserDetails.account-security.change-on-login.description" />
              </Text>
            </div>
            <Toggle
              checked={changeOnLogin}
              disabled={passwordExpired}
              onCheckedChange={showMarkAsExpiredConfirmationDialog}
            />
          </div>
        ) : null}
        {passwordExpired ? (
          <Callout
            className={styles.passwordCellExpired}
            type="warning"
            showCloseButton={false}
            text={
              <FormattedMessage
                id="UserDetails.account-security.expired"
                values={{
                  expiredInDays,
                  // eslint-disable-next-line react/no-unstable-nested-components
                  text: (chunks: React.ReactNode) => (
                    <Text className={styles.passwordCellExpiredPrefix}>
                      {chunks}
                    </Text>
                  ),
                }}
              />
            }
          />
        ) : null}
        {kind === "SECONDARY" ? (
          <Button
            className={cn(styles.button, styles.removePasswordButton)}
            size="1"
            variant="outline"
            color="red"
            onClick={onRemoveClicked}
          >
            <FormattedMessage id="remove" />
          </Button>
        ) : null}
      </ListCellLayout>
    );
  };

const TOTPAuthenticatorCell: React.VFC<TOTPAuthenticatorCellProps> =
  function TOTPAuthenticatorCell(props: TOTPAuthenticatorCellProps) {
    const { id, kind, label, addedOn, showConfirmationDialog, withTopSpacing } =
      props;

    if (kind === AuthenticatorKind.Secondary) {
      return (
        <SecondaryAuthenticatorTableCell
          id={id}
          label={label}
          addedOn={addedOn}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    }

    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.totpCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
        <Text
          size="2"
          weight="medium"
          className={cn(styles.cellLabel, styles.totpCellLabel)}
        >
          {label}
        </Text>
        <Text
          size="2"
          color="gray"
          className={cn(styles.cellDesc, styles.totpCellDesc)}
        >
          <FormattedMessage
            id="UserDetails.account-security.added-on"
            values={{ datetime: addedOn }}
          />
        </Text>
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

    if (kind === AuthenticatorKind.Secondary) {
      return (
        <SecondaryAuthenticatorTableCell
          id={id}
          label={label}
          addedOn={addedOn}
          showConfirmationDialog={showConfirmationDialog}
        />
      );
    }

    return (
      <ListCellLayout
        className={cn(
          styles.cell,
          styles.oobOtpCell,
          withTopSpacing ? styles["cell--not-first"] : ""
        )}
      >
        <span className={styles.oobOtpCellIcon}>
          {iconName === "Mail" ? <EnvelopeClosedIcon /> : <MobileIcon />}
        </span>
        <Text
          size="2"
          weight="medium"
          className={cn(styles.cellLabel, styles.oobOtpCellLabel)}
        >
          {label}
        </Text>
        <Text
          size="2"
          color="gray"
          className={cn(styles.cellDesc, styles.oobOtpCellAddedOn)}
        >
          <FormattedMessage
            id="UserDetails.account-security.added-on"
            values={{ datetime: addedOn }}
          />
        </Text>
      </ListCellLayout>
    );
  };

const UserDetailsAccountSecurity: React.VFC<UserDetailsAccountSecurityProps> =
  function UserDetailsAccountSecurity(props: UserDetailsAccountSecurityProps) {
    const {
      userID,
      authenticationConfig,
      authenticatorConfig,
      identities,
      authenticators,
      phoneInputAllowlist,
      phoneInputPinnedList,
      onAuthenticatorCreated,
    } = props;
    const { locale } = useContext(Context);
    const navigate = useNavigate();
    const [add2FAPhoneDialogOpen, setAdd2FAPhoneDialogOpen] = useState(false);
    const [add2FAEmailDialogOpen, setAdd2FAEmailDialogOpen] = useState(false);
    const [add2FAPasswordDialogOpen, setAdd2FAPasswordDialogOpen] =
      useState(false);

    const { user } = useUserQuery(userID);

    const passwordForceChangeDaysSinceLastUpdate = useMemo(() => {
      const expiryForceChangeConfig =
        authenticatorConfig?.password?.expiry?.force_change;
      if (expiryForceChangeConfig?.enabled !== true) {
        return null;
      }

      const durationString = expiryForceChangeConfig.duration_since_last_update;
      if (durationString == null) {
        return null;
      }

      const secondsPerDay = 24 * 60 * 60;
      return Math.round(
        durationString ? parseDuration(durationString) / secondsPerDay : 1
      );
    }, [authenticatorConfig]);

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
        authenticationConfig,
        identities,
        authenticators,
        locale
      );
    }, [authenticationConfig, locale, identities, authenticators]);

    const secondaryAuthenticatorLists = useMemo(() => {
      return constructSecondaryAuthenticatorList(
        authenticationConfig,
        authenticators,
        locale
      );
    }, [authenticationConfig, authenticators, locale]);

    const secondaryAuthicatorIsRequired =
      authenticationConfig?.secondary_authentication_mode === "required";

    const isWithinPerUserMFAGracePeriod = useMemo(() => {
      if (user?.mfaGracePeriodEndAt == null) {
        return false;
      }

      return parseDate(user.mfaGracePeriodEndAt) >= new Date();
    }, [user]);

    const globalGracePeriodEndAt = useMemo(() => {
      if (
        authenticationConfig?.secondary_authentication_grace_period?.enabled !==
        true
      ) {
        return null;
      }

      const gracePeriodEndAtString =
        authenticationConfig.secondary_authentication_grace_period.end_at;
      if (gracePeriodEndAtString == null) {
        return null;
      }
      return parseDate(gracePeriodEndAtString);
    }, [authenticationConfig]);

    const userGracePeriod = useMemo(() => {
      if (user?.mfaGracePeriodEndAt == null) {
        return null;
      }
      return parseDate(user.mfaGracePeriodEndAt);
    }, [user]);

    const isWithinGlobalMFAGracePeriod = useMemo(() => {
      if (
        authenticationConfig?.secondary_authentication_grace_period?.enabled !==
        true
      ) {
        return false;
      }

      if (globalGracePeriodEndAt == null) {
        return true;
      }

      return globalGracePeriodEndAt >= new Date();
    }, [authenticationConfig, globalGracePeriodEndAt]);

    const isWithinMFAGracePeriod = useMemo(() => {
      return isWithinPerUserMFAGracePeriod || isWithinGlobalMFAGracePeriod;
    }, [isWithinPerUserMFAGracePeriod, isWithinGlobalMFAGracePeriod]);

    const farthestMFAGracePeriodEndAt = useMemo(() => {
      const globalEndAt = globalGracePeriodEndAt;
      const userEndAt = userGracePeriod;
      if (globalEndAt == null) {
        // Two cases:
        // 1. null global end_at means no deadline
        if (
          authenticationConfig?.secondary_authentication_grace_period
            ?.enabled === true
        ) {
          return globalEndAt;
        }
        // 2. gloabl grace period not enabled
        return userEndAt;
      }
      if (userEndAt == null) {
        return globalEndAt;
      }
      return globalEndAt > userEndAt ? globalEndAt : userEndAt;
    }, [globalGracePeriodEndAt, userGracePeriod, authenticationConfig]);

    const userMFAGracePeriodEndAt = user?.mfaGracePeriodEndAt;
    const canExtendMFAGracePeriod = useMemo(() => {
      // Global grace period without deadline, no need to extend
      if (
        authenticationConfig?.secondary_authentication_grace_period?.enabled ===
          true &&
        globalGracePeriodEndAt == null
      ) {
        return false;
      }

      // user grace period not enabled
      if (userMFAGracePeriodEndAt == null) {
        return false;
      }

      // Can extend the deadline if user grace period is after global period
      if (
        authenticationConfig?.secondary_authentication_grace_period?.end_at !=
        null
      ) {
        const gracePeriod = parseDate(
          authenticationConfig.secondary_authentication_grace_period.end_at
        );
        const userGracePeriod = parseDate(userMFAGracePeriodEndAt);
        return userGracePeriod >= gracePeriod;
      }

      return true;
    }, [authenticationConfig, globalGracePeriodEndAt, userMFAGracePeriodEndAt]);

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

    const setPasswordExpiredConfirmDialog = useConfirmationDialog();

    const { setPasswordExpired, error: setPasswordExpiredError } =
      useSetPasswordExpiredMutation();
    useProvideError(setPasswordExpiredError);

    const isExpired = useMemo(
      () =>
        primaryAuthenticatorLists.password.some(
          (authenticator) => authenticator.manualChangeOnLogin
        ),
      [primaryAuthenticatorLists.password]
    );

    const onConfirmSetPasswordExpired = useCallback(async () => {
      await setPasswordExpired(userID, !isExpired);
      setPasswordExpiredConfirmDialog.dismiss();
    }, [
      setPasswordExpired,
      userID,
      isExpired,
      setPasswordExpiredConfirmDialog,
    ]);

    const setMFAGracePeriodConfirmationDialog = useConfirmationDialog();
    const cancelMFAGracePeriodConfirmationDialog = useConfirmationDialog();

    const { setMFAGracePeriod, error: setMFAGracePeriodError } =
      useSetMFAGracePeriodMutation();
    useProvideError(setMFAGracePeriodError);

    const { removeMFAGracePeriod, error: removeMFAGracePeriodError } =
      useRemoveMFAGracePeriodMutation();
    useProvideError(removeMFAGracePeriodError);

    const mfaGracePeriodAction = useMemo(() => {
      if (!isWithinPerUserMFAGracePeriod) {
        return MFAGracePeriodAction.Grant;
      }

      return MFAGracePeriodAction.Extend;
    }, [isWithinPerUserMFAGracePeriod]);

    const onConfirmSetMFAGracePeriod = useCallback(async () => {
      switch (mfaGracePeriodAction) {
        case MFAGracePeriodAction.Grant: {
          const newGracePeriod = DateTime.now().plus({ days: 10 }).toJSDate();
          await setMFAGracePeriod(userID, newGracePeriod);
          break;
        }
        case MFAGracePeriodAction.Extend: {
          const fromDate = DateTime.max(
            DateTime.now(),
            DateTime.fromISO(user?.mfaGracePeriodEndAt)
          );
          const newGracePeriod = fromDate.plus({ days: 10 }).toJSDate();
          await setMFAGracePeriod(userID, newGracePeriod);
          break;
        }
      }
      setMFAGracePeriodConfirmationDialog.dismiss();
    }, [
      mfaGracePeriodAction,
      setMFAGracePeriodConfirmationDialog,
      setMFAGracePeriod,
      userID,
      user?.mfaGracePeriodEndAt,
    ]);

    const onConfirmRemoveMFAGracePeriod = useCallback(async () => {
      await removeMFAGracePeriod(userID);
      cancelMFAGracePeriodConfirmationDialog.dismiss();
    }, [removeMFAGracePeriod, userID, cancelMFAGracePeriodConfirmationDialog]);

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
            forceChangeDaysSinceLastUpdate={
              passwordForceChangeDaysSinceLastUpdate
            }
            showRemoveConfirmationDialog={showConfirmationDialog}
            showMarkAsExpiredConfirmationDialog={
              setPasswordExpiredConfirmDialog.show
            }
          />
        );
      },
      [
        passwordForceChangeDaysSinceLastUpdate,
        showConfirmationDialog,
        setPasswordExpiredConfirmDialog.show,
      ]
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

    const onRenderExtendedMFAGracePeriod = useCallback(() => {
      return (
        <LinkButton
          className={styles.authenticatorGrantGracePeriod}
          onClick={setMFAGracePeriodConfirmationDialog.show}
        >
          <FormattedMessage
            id={"UserDetails.account-security.secondary.extend-grace-period"}
          />
        </LinkButton>
      );
    }, [setMFAGracePeriodConfirmationDialog.show]);

    const onRenderCancelMFAGracePeriod = useCallback(() => {
      return (
        <LinkButton
          className={styles.authenticatorGrantGracePeriod}
          onClick={cancelMFAGracePeriodConfirmationDialog.show}
        >
          <FormattedMessage
            id={"UserDetails.account-security.secondary.cancel-grace-period"}
          />
        </LinkButton>
      );
    }, [cancelMFAGracePeriodConfirmationDialog.show]);

    const addPrimaryPassword = useCallback(() => {
      navigate("./add-password");
    }, [navigate]);

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
        {primaryAuthenticatorLists.hasVisibleList ||
        primaryAuthenticatorLists.isPrimaryPasswordEnabled ? (
          <div className={styles.authenticatorContainer}>
            <div
              className={cn(
                "flex justify-between",
                styles.authenticatorKindHeader
              )}
            >
              <Text
                as="p"
                size="3"
                weight="medium"
                className={cn(styles.header)}
              >
                <FormattedMessage id="UserDetails.account-security.primary" />
              </Text>
              {primaryAuthenticatorLists.password.length === 0 ? (
                <Button size="2" onClick={addPrimaryPassword}>
                  <PlusIcon />
                  <FormattedMessage id="UserDetails.account-security.primary.password.add" />
                </Button>
              ) : null}
            </div>
            {!primaryAuthenticatorLists.hasVisibleList ? (
              <>
                <Text
                  as="p"
                  size="2"
                  color="gray"
                  className={cn(styles.authenticatorEmpty)}
                >
                  <FormattedMessage id="UserDetails.account-security.primary.empty" />
                </Text>
              </>
            ) : null}
            {primaryAuthenticatorLists.password.length > 0 ? (
              <div
                className={cn(
                  styles.authenticatorTypeSection,
                  styles["authenticatorTypeSection--password"]
                )}
              >
                {primaryAuthenticatorLists.password.map((item, index) => (
                  <React.Fragment key={item.id}>
                    {onRenderPasswordAuthenticatorDetailCell(item, index)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
            {primaryAuthenticatorLists.passkey.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.primary.passkey" />
                </Text>
                {primaryAuthenticatorLists.passkey.map((item, index) => (
                  <React.Fragment key={item.id}>
                    {onRenderPasskeyIdentityDetailCell(item, index)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
            {primaryAuthenticatorLists.oobOtpEmail.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.primary.oob-otp-email" />
                </Text>
                {primaryAuthenticatorLists.oobOtpEmail.map((item, index) => (
                  <React.Fragment key={item.id}>
                    {onRenderOobOtpAuthenticatorDetailCell(item, index)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
            {primaryAuthenticatorLists.oobOtpSMS.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={cn(styles.header, styles.authenticatorTypeHeader)}
                >
                  <FormattedMessage id="AuthenticatorType.primary.oob-otp-phone" />
                </Text>
                {primaryAuthenticatorLists.oobOtpSMS.map((item, index) => (
                  <React.Fragment key={item.id}>
                    {onRenderOobOtpAuthenticatorDetailCell(item, index)}
                  </React.Fragment>
                ))}
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
              <Text
                as="p"
                size="3"
                weight="medium"
                className={cn(styles.header)}
              >
                <FormattedMessage id="UserDetails.account-security.secondary" />
              </Text>
            </div>
            {secondaryAuthicatorIsRequired &&
            !secondaryAuthenticatorLists.hasVisibleList ? (
              <div className={styles.secondaryEmpty}>
                <Text
                  as="p"
                  size="3"
                  weight="medium"
                  className={styles.secondaryEmptyTitle}
                >
                  <FormattedMessage id="UserDetails.account-security.secondary.empty" />
                </Text>
                <Text
                  as="p"
                  size="2"
                  color="gray"
                  className={styles.secondaryEmptyDescription}
                >
                  <FormattedMessage
                    id="UserDetails.account-security.secondary.empty-description"
                    values={{
                      // eslint-disable-next-line react/no-unstable-nested-components
                      gracePeriod: (chunks: React.ReactNode) => (
                        <LinkButton
                          className={styles.authenticatorGrantGracePeriod}
                          onClick={setMFAGracePeriodConfirmationDialog.show}
                        >
                          {chunks}
                        </LinkButton>
                      ),
                    }}
                  />
                </Text>
              </div>
            ) : null}
            {secondaryAuthenticatorLists.totp.length > 0 ? (
              <div className={styles.authenticatorTypeSection}>
                <div className={styles.authenticatorTable}>
                  <div className={styles.authenticatorTableHeader}>
                    <Text
                      as="p"
                      size="2"
                      weight="medium"
                      className={styles.authenticatorTableGroupTitle}
                    >
                      <FormattedMessage id="AuthenticatorType.secondary.totp" />
                      {!secondaryAuthenticatorLists.isSecondaryTOTPEnabled ? (
                        <>
                          {" "}
                          <FormattedMessage id="UserDetails.account-security.disabled" />
                        </>
                      ) : null}
                    </Text>
                  </div>
                  {secondaryAuthenticatorLists.totp.map((item, index) => (
                    <React.Fragment key={item.id}>
                      {onRenderTotpAuthenticatorDetailCell(item, index)}
                    </React.Fragment>
                  ))}
                </div>
              </div>
            ) : null}
            {secondaryAuthenticatorLists.isSecondaryOOBOTPEmailEnabled ||
            secondaryAuthenticatorLists.oobOtpEmail.length > 0 ? (
              <div className={styles.authenticatorTypeGroup}>
                <div className={styles.authenticatorTypeSection}>
                  <div className={styles.authenticatorTable}>
                    <div className={styles.authenticatorTableHeader}>
                      <Text
                        as="p"
                        size="2"
                        weight="medium"
                        className={styles.authenticatorTableGroupTitle}
                      >
                        <FormattedMessage id="AuthenticatorType.secondary.oob-otp-email" />
                        {!secondaryAuthenticatorLists.isSecondaryOOBOTPEmailEnabled ? (
                          <>
                            {" "}
                            <FormattedMessage id="UserDetails.account-security.disabled" />
                          </>
                        ) : null}
                      </Text>
                    </div>
                    {secondaryAuthenticatorLists.oobOtpEmail.map(
                      (item, index) => (
                        <React.Fragment key={item.id}>
                          {onRenderOobOtpAuthenticatorDetailCell(item, index)}
                        </React.Fragment>
                      )
                    )}
                  </div>
                </div>
                {secondaryAuthenticatorLists.isSecondaryOOBOTPEmailEnabled ? (
                  <button
                    type="button"
                    className={styles.addAuthenticatorButton}
                    onClick={() => setAdd2FAEmailDialogOpen(true)}
                  >
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="UserDetails.account-security.secondary.add-oob-otp-email" />
                  </button>
                ) : null}
              </div>
            ) : null}
            {secondaryAuthenticatorLists.isSecondaryOOBOTPSMSEnabled ||
            secondaryAuthenticatorLists.oobOtpSMS.length > 0 ? (
              <div className={styles.authenticatorTypeGroup}>
                <div className={styles.authenticatorTypeSection}>
                  <div className={styles.authenticatorTable}>
                    <div className={styles.authenticatorTableHeader}>
                      <Text
                        as="p"
                        size="2"
                        weight="medium"
                        className={styles.authenticatorTableGroupTitle}
                      >
                        <FormattedMessage id="AuthenticatorType.secondary.oob-otp-phone" />
                        {!secondaryAuthenticatorLists.isSecondaryOOBOTPSMSEnabled ? (
                          <>
                            {" "}
                            <FormattedMessage id="UserDetails.account-security.disabled" />
                          </>
                        ) : null}
                      </Text>
                    </div>
                    {secondaryAuthenticatorLists.oobOtpSMS.map(
                      (item, index) => (
                        <React.Fragment key={item.id}>
                          {onRenderOobOtpAuthenticatorDetailCell(item, index)}
                        </React.Fragment>
                      )
                    )}
                  </div>
                </div>
                {secondaryAuthenticatorLists.isSecondaryOOBOTPSMSEnabled ? (
                  <button
                    type="button"
                    className={styles.addAuthenticatorButton}
                    onClick={() => setAdd2FAPhoneDialogOpen(true)}
                  >
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="UserDetails.account-security.secondary.add-oob-otp-phone" />
                  </button>
                ) : null}
              </div>
            ) : null}
            {secondaryAuthenticatorLists.isSecondaryPasswordEnabled ||
            secondaryAuthenticatorLists.password.length > 0 ? (
              <div className={styles.authenticatorTypeGroup}>
                <div className={styles.authenticatorTypeSection}>
                  <Text
                    as="p"
                    size="2"
                    weight="medium"
                    className={cn(
                      styles.header,
                      styles.authenticatorTypeHeader
                    )}
                  >
                    <FormattedMessage id="AuthenticatorType.secondary.password" />
                    {!secondaryAuthenticatorLists.isSecondaryPasswordEnabled ? (
                      <>
                        {" "}
                        <FormattedMessage id="UserDetails.account-security.disabled" />
                      </>
                    ) : null}
                  </Text>
                  <div
                    className={cn(
                      styles.authenticatorTypeSection,
                      styles["authenticatorTypeSection--password"]
                    )}
                  >
                    {secondaryAuthenticatorLists.password.map(
                      (item, index) => (
                        <React.Fragment key={item.id}>
                          {onRenderPasswordAuthenticatorDetailCell(item, index)}
                        </React.Fragment>
                      )
                    )}
                  </div>
                </div>
                {secondaryAuthenticatorLists.isSecondaryPasswordEnabled &&
                secondaryAuthenticatorLists.password.length === 0 ? (
                  <button
                    type="button"
                    className={styles.addAuthenticatorButton}
                    onClick={() => setAdd2FAPasswordDialogOpen(true)}
                  >
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="UserDetails.account-security.secondary.add-password" />
                  </button>
                ) : null}
              </div>
            ) : null}
            {secondaryAuthicatorIsRequired &&
            secondaryAuthenticatorLists.hasVisibleList &&
            !secondaryAuthenticatorLists.hasEnabledAuthenticator ? (
              <>
                <Text
                  as="p"
                  size="2"
                  color="gray"
                  className={cn(styles.authenticatorEmpty)}
                >
                  <FormattedMessage id="UserDetails.account-security.secondary.no-enabled-authenticators" />{" "}
                  {!isWithinMFAGracePeriod ? (
                    <FormattedMessage id="UserDetails.account-security.secondary.cannot-login" />
                  ) : farthestMFAGracePeriodEndAt != null ? (
                    <FormattedMessage
                      id="UserDetails.account-security.secondary.within-grace-period"
                      values={{
                        gracePeriodEndAt:
                          formatDatetime(locale, farthestMFAGracePeriodEndAt) ??
                          "",
                      }}
                    />
                  ) : (
                    <FormattedMessage id="UserDetails.account-security.secondary.within-grace-period.no-deadline" />
                  )}
                </Text>
                {!isWithinMFAGracePeriod ? (
                  <LinkButton
                    className={styles.authenticatorGrantGracePeriod}
                    onClick={setMFAGracePeriodConfirmationDialog.show}
                  >
                    <FormattedMessage
                      id={
                        "UserDetails.account-security.secondary.grant-grace-period"
                      }
                    />
                  </LinkButton>
                ) : null}
                {canExtendMFAGracePeriod ? (
                  <div className={styles.updateMFAGracePeriodContainer}>
                    <FormattedMessage
                      id="UserDetails.account-security.secondary.update-existing-grace-period"
                      values={{
                        extend: onRenderExtendedMFAGracePeriod,
                        cancel: onRenderCancelMFAGracePeriod,
                      }}
                    />
                  </div>
                ) : null}
              </>
            ) : null}
          </div>
        ) : null}
        <SetPasswordExpiredConfirmationDialog
          store={setPasswordExpiredConfirmDialog}
          isExpired={isExpired}
          onConfirm={() => {
            onConfirmSetPasswordExpired().finally(() => {});
          }}
        />
        <SetMFAGracePeriodConfirmationDialog
          store={setMFAGracePeriodConfirmationDialog}
          action={mfaGracePeriodAction}
          onConfirm={() => {
            onConfirmSetMFAGracePeriod().finally(() => {});
          }}
        />
        <CancelMFAGracePeriodConfirmationDialog
          store={cancelMFAGracePeriodConfirmationDialog}
          onConfirm={() => {
            onConfirmRemoveMFAGracePeriod().finally(() => {});
          }}
        />
        <Add2FAPhoneDialog
          open={add2FAPhoneDialogOpen}
          userID={userID}
          phoneInputAllowlist={phoneInputAllowlist}
          phoneInputPinnedList={phoneInputPinnedList}
          onOpenChange={setAdd2FAPhoneDialogOpen}
          onCreated={onAuthenticatorCreated}
        />
        <Add2FAEmailDialog
          open={add2FAEmailDialogOpen}
          userID={userID}
          onOpenChange={setAdd2FAEmailDialogOpen}
          onCreated={onAuthenticatorCreated}
        />
        <Add2FAPasswordDialog
          open={add2FAPasswordDialogOpen}
          userID={userID}
          passwordPolicy={authenticatorConfig?.password?.policy ?? {}}
          onOpenChange={setAdd2FAPasswordDialogOpen}
          onCreated={onAuthenticatorCreated}
        />
      </div>
    );
  };

export default UserDetailsAccountSecurity;
