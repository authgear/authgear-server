import React, { useMemo, useCallback, useContext, useState } from "react";
import cn from "classnames";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { FormattedMessage, Context } from "../../intl";
import {
  Button,
  Dialog,
  DropdownMenu,
  IconButton,
  Text,
} from "@radix-ui/themes";
import {
  CheckCircledIcon,
  CrossCircledIcon,
  DotsVerticalIcon,
  EnvelopeClosedIcon,
  IdCardIcon,
  MobileIcon,
  Pencil1Icon,
  PersonIcon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";

// import PrimaryIdentitiesSelectionForm from "./PrimaryIdentitiesSelectionForm";
import ListCellLayout from "../../ListCellLayout";
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import { useSetVerifiedStatusMutation } from "./mutations/setVerifiedStatusMutation";
import { formatDatetime } from "../../util/formatDatetime";
import { formatDateOnly } from "../../util/formatDateOnly";
import { LoginIDKeyType, OAuthSSOProviderType } from "../../types";
import { UserQueryNodeFragment } from "./query/userQuery.generated";

import styles from "./UserDetailsConnectedIdentities.module.css";
import listStyles from "./UserDetailsListTable.module.css";
import { useIsLoading, useLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";
import ExternalLink, { ExternalLinkProps } from "../../ExternalLink";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { TextField } from "../../components/v2/TextField/TextField";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import PhoneTextField from "../../PhoneTextField";
import phoneDialogStyles from "../../PhoneDialog.module.css";
import { useUpdateLoginIDIdentityMutation } from "./mutations/updateIdentityMutation";

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
  preferred_username?: string;
  "https://authgear.com/claims/oauth/provider_type"?: OAuthSSOProviderType;
  "https://authgear.com/claims/oauth/subject_id"?: string;
  "https://authgear.com/claims/login_id/type"?: LoginIDIdentityType;
  "https://authgear.com/claims/ldap/last_login_username"?: string | null;
  "https://authgear.com/claims/ldap/user_id_attribute_name"?: string;
  "https://authgear.com/claims/ldap/user_id_attribute_value"?: string;
}

interface Identity {
  id: string;
  type:
    | "ANONYMOUS"
    | "LOGIN_ID"
    | "OAUTH"
    | "BIOMETRIC"
    | "PASSKEY"
    | "SIWE"
    | "LDAP";
  claims: IdentityClaim;
  createdAt: string;
  updatedAt: string;
}

type VerifiedClaims = UserQueryNodeFragment["verifiedClaims"];
interface UserDetailsConnectedIdentitiesProps {
  identities: Identity[];
  verifiedClaims: VerifiedClaims;
  availableLoginIdIdentities: string[];
  phoneInputAllowlist?: string[];
  phoneInputPinnedList?: string[];
  onIdentityCreated?: () => unknown;
}

const loginIdIdentityTypes = ["email", "phone", "username"] as const;
type LoginIDIdentityType = (typeof loginIdIdentityTypes)[number];
type IdentityType =
  | "login_id"
  | "oauth"
  | "biometric"
  | "anonymous"
  | "siwe"
  | "ldap";

type IdentityListItem =
  | OAuthIdentityListItem
  | LoginIDIdentityListItem
  | BiometricIdentityListItem
  | AnonymousIdentityListItem
  | LDAPIdentityListItem;
interface OAuthIdentityListItem {
  id: string;
  type: "oauth";
  providerType: OAuthSSOProviderType;
  subjectID?: string;
  claimName?: string;
  claimValue?: string;
  verified?: boolean;
  connectedOn: string;
}

interface LoginIDIdentityListItem {
  id: string;
  type: "login_id";
  loginIDKey: LoginIDKeyType;
  claimName: string;
  claimValue: string;
  verified?: boolean;
  connectedOn: string;
  connectedOnDateOnly: string;
}

interface BiometricIdentityListItem {
  id: string;
  type: "biometric";
  connectedOn: string;
  verified: undefined;
  formattedDeviceInfo: string;
}

interface AnonymousIdentityListItem {
  id: string;
  type: "anonymous";
  verified: undefined;
  connectedOn: string;
}

interface LDAPIdentityListItem {
  id: string;
  type: "ldap";
  verified: undefined;
  connectedOn: string;
  lastLoginUserName?: string | null;
  userIDAttributeName?: string;
  userIDAttributeValue?: string;
}

export interface IdentityLists {
  oauth: OAuthIdentityListItem[];
  email: LoginIDIdentityListItem[];
  phone: LoginIDIdentityListItem[];
  username: LoginIDIdentityListItem[];
  biometric: BiometricIdentityListItem[];
  anonymous: AnonymousIdentityListItem[];
  ldap: LDAPIdentityListItem[];
}

interface VerifyButtonProps {
  disabled?: boolean;
  verified: boolean;
  verifying: boolean;
  toggleVerified: (verified: boolean) => void;
}

interface ConfirmationDialogData {
  identityID: string;
  identityName: string;
}

interface AddEmailDialogProps {
  open: boolean;
  userID: string;
  identityToEdit?: {
    id: string;
    value: string;
  };
  onOpenChange: (open: boolean) => void;
  onCreated?: () => unknown;
}

function AddEmailDialog({
  open,
  userID,
  identityToEdit,
  onOpenChange,
  onCreated,
}: AddEmailDialogProps): React.ReactElement {
  const [email, setEmail] = useState(() =>
    open ? identityToEdit?.value ?? "" : ""
  );
  const { createIdentity, loading, error } =
    useCreateLoginIDIdentityMutation(userID);
  const {
    updateIdentity,
    loading: updating,
    error: updateError,
  } = useUpdateLoginIDIdentityMutation(userID);
  const isLoading = loading || updating;
  useLoading(isLoading);
  useProvideError(error ?? updateError);

  const [prevOpen, setPrevOpen] = useState(open);
  const [prevIdentityToEdit, setPrevIdentityToEdit] = useState(identityToEdit);
  if (prevOpen !== open || prevIdentityToEdit !== identityToEdit) {
    setPrevOpen(open);
    setPrevIdentityToEdit(identityToEdit);
    setEmail(open ? identityToEdit?.value ?? "" : "");
  }

  const onSubmit = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      const value = email.trim();
      if (value === "" || isLoading) {
        return;
      }

      try {
        const identity =
          identityToEdit == null
            ? await createIdentity({ key: "email", value })
            : await updateIdentity(identityToEdit.id, {
                key: "email",
                value,
              });
        if (identity != null) {
          await onCreated?.();
          onOpenChange(false);
        }
      } catch {
        // The mutation error is surfaced by useProvideError.
      }
    },
    [
      createIdentity,
      email,
      identityToEdit,
      isLoading,
      onCreated,
      onOpenChange,
      updateIdentity,
    ]
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage
            id={
              identityToEdit == null
                ? "EmailScreen.add.title"
                : "EmailScreen.edit.title"
            }
          />
        </Dialog.Title>
        <Dialog.Description size="2">
          {identityToEdit == null ? (
            <FormattedMessage id="EmailScreen.add.description" />
          ) : (
            <FormattedMessage
              id="EmailScreen.edit.current-value"
              values={{ value: identityToEdit.value }}
            />
          )}
        </Dialog.Description>
        <form
          className={styles.addIdentityForm}
          onSubmit={(event) => {
            onSubmit(event).finally(() => {});
          }}
        >
          <TextField
            size="2"
            type="email"
            label={<FormattedMessage id="EmailScreen.email.label" />}
            value={email}
            onChange={(event) => {
              setEmail(event.currentTarget.value);
            }}
          />
          <div className={styles.addIdentityDialogActions}>
            <SecondaryButton
              size="2"
              disabled={isLoading}
              text={<FormattedMessage id="cancel" />}
              onClick={() => {
                onOpenChange(false);
              }}
            />
            <Button
              type="submit"
              size="2"
              loading={isLoading}
              disabled={email.trim() === ""}
            >
              <FormattedMessage id={identityToEdit == null ? "add" : "save"} />
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}

interface AddUsernameDialogProps {
  open: boolean;
  userID: string;
  identityToEdit?: {
    id: string;
    value: string;
  };
  onOpenChange: (open: boolean) => void;
  onCreated?: () => unknown;
}

function AddUsernameDialog({
  open,
  userID,
  identityToEdit,
  onOpenChange,
  onCreated,
}: AddUsernameDialogProps): React.ReactElement {
  const [username, setUsername] = useState(() =>
    open ? identityToEdit?.value ?? "" : ""
  );
  const { createIdentity, loading, error } =
    useCreateLoginIDIdentityMutation(userID);
  const {
    updateIdentity,
    loading: updating,
    error: updateError,
  } = useUpdateLoginIDIdentityMutation(userID);
  const isLoading = loading || updating;
  useLoading(isLoading);
  useProvideError(error ?? updateError);

  const [prevOpen, setPrevOpen] = useState(open);
  const [prevIdentityToEdit, setPrevIdentityToEdit] = useState(identityToEdit);
  if (prevOpen !== open || prevIdentityToEdit !== identityToEdit) {
    setPrevOpen(open);
    setPrevIdentityToEdit(identityToEdit);
    setUsername(open ? identityToEdit?.value ?? "" : "");
  }

  const onSubmit = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      const value = username.trim();
      if (value === "" || isLoading) {
        return;
      }

      try {
        const identity =
          identityToEdit == null
            ? await createIdentity({ key: "username", value })
            : await updateIdentity(identityToEdit.id, {
                key: "username",
                value,
              });
        if (identity != null) {
          await onCreated?.();
          onOpenChange(false);
        }
      } catch {
        // The mutation error is surfaced by useProvideError.
      }
    },
    [
      createIdentity,
      username,
      identityToEdit,
      isLoading,
      onCreated,
      onOpenChange,
      updateIdentity,
    ]
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage
            id={
              identityToEdit == null
                ? "UsernameScreen.add.title"
                : "UsernameScreen.edit.title"
            }
          />
        </Dialog.Title>
        <Dialog.Description size="2">
          {identityToEdit == null ? (
            <FormattedMessage id="UsernameScreen.add.description" />
          ) : (
            <FormattedMessage
              id="UsernameScreen.edit.current-value"
              values={{ value: identityToEdit.value }}
            />
          )}
        </Dialog.Description>
        <form
          className={styles.addIdentityForm}
          onSubmit={(event) => {
            onSubmit(event).finally(() => {});
          }}
        >
          <TextField
            size="2"
            type="text"
            label={<FormattedMessage id="UsernameScreen.username.label" />}
            value={username}
            onChange={(event) => {
              setUsername(event.currentTarget.value);
            }}
          />
          <div className={styles.addIdentityDialogActions}>
            <SecondaryButton
              size="2"
              disabled={isLoading}
              text={<FormattedMessage id="cancel" />}
              onClick={() => {
                onOpenChange(false);
              }}
            />
            <Button
              type="submit"
              size="2"
              loading={isLoading}
              disabled={username.trim() === ""}
            >
              <FormattedMessage id={identityToEdit == null ? "add" : "save"} />
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}

interface AddPhoneDialogProps {
  open: boolean;
  userID: string;
  phoneInputAllowlist?: string[];
  phoneInputPinnedList?: string[];
  onOpenChange: (open: boolean) => void;
  onCreated?: () => unknown;
}

function AddPhoneDialog({
  open,
  userID,
  phoneInputAllowlist,
  phoneInputPinnedList,
  onOpenChange,
  onCreated,
}: AddPhoneDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [e164, setE164] = useState("");
  const [rawInputValue, setRawInputValue] = useState("");
  const [fieldKey, setFieldKey] = useState(0);
  const { createIdentity, loading, error } =
    useCreateLoginIDIdentityMutation(userID);
  useLoading(loading);
  useProvideError(error);

  const [prevOpen, setPrevOpen] = useState(open);
  if (prevOpen !== open) {
    setPrevOpen(open);
    if (!open) {
      setE164("");
      setRawInputValue("");
    } else {
      setFieldKey((key) => key + 1);
    }
  }

  const onSubmit = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      if (e164 === "" || loading) {
        return;
      }

      try {
        const identity = await createIdentity({
          key: "phone",
          value: e164,
        });
        if (identity != null) {
          await onCreated?.();
          onOpenChange(false);
        }
      } catch {
        // The mutation error is surfaced by useProvideError.
      }
    },
    [createIdentity, e164, loading, onCreated, onOpenChange]
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content
        maxWidth="480px"
        size="3"
        className={phoneDialogStyles.phoneDialogContent}
        data-phone-dialog="true"
      >
        <Dialog.Title>
          <FormattedMessage id="PhoneScreen.add.title" />
        </Dialog.Title>
        <Dialog.Description size="2">
          <FormattedMessage id="PhoneScreen.add.description" />
        </Dialog.Description>
        <form
          className={cn(
            styles.addIdentityForm,
            phoneDialogStyles.phoneDialogForm
          )}
          onSubmit={(event) => {
            onSubmit(event).finally(() => {});
          }}
        >
          <PhoneTextField
            key={fieldKey}
            label={renderToString("PhoneScreen.phone.label")}
            allowlist={phoneInputAllowlist}
            pinnedList={phoneInputPinnedList}
            initialInputValue={rawInputValue}
            onChange={(values) => {
              setE164(values.e164 ?? "");
              setRawInputValue(values.rawInputValue);
            }}
          />
          <div
            className={cn(
              styles.addIdentityDialogActions,
              phoneDialogStyles.phoneDialogActions
            )}
          >
            <SecondaryButton
              size="2"
              disabled={loading}
              text={<FormattedMessage id="cancel" />}
              onClick={() => {
                onOpenChange(false);
              }}
            />
            <Button
              type="submit"
              size="2"
              loading={loading}
              disabled={e164 === ""}
            >
              <FormattedMessage id="add" />
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}

const oauthIconMap: Record<OAuthSSOProviderType, React.ReactNode> = {
  apple: <i className={cn("fab", "fa-apple")} />,
  google: <i className={cn("fab", "fa-google")} />,
  facebook: <i className={cn("fab", "fa-facebook")} />,
  github: <i className={cn("fab", "fa-github")} />,
  linkedin: <i className={cn("fab", "fa-linkedin")} />,
  azureadv2: <i className={cn("fab", "fa-microsoft")} />,
  azureadb2c: <i className={cn("fab", "fa-microsoft")} />,
  adfs: <i className={cn("fab", "fa-microsoft")} />,
  wechat: <i className={cn("fab", "fa-weixin")} />,
};

const loginIdIconMap: Record<LoginIDIdentityType, React.ReactNode> = {
  email: <EnvelopeClosedIcon />,
  phone: <MobileIcon />,
  username: <PersonIcon />,
};

const biometricIcon: React.ReactNode = <IdCardIcon />;
const anonymousIcon: React.ReactNode = <PersonIcon />;
const ldapIcon: React.ReactNode = <IdCardIcon />;

const removeButtonTextId: Record<IdentityType, "remove" | "disconnect" | ""> = {
  oauth: "disconnect",
  login_id: "remove",
  biometric: "remove",
  anonymous: "",
  siwe: "",
  ldap: "",
};

function getIdentityName(
  item: IdentityListItem,
  renderToString: (id: string) => string
): string {
  switch (item.type) {
    case "oauth":
      return (
        item.claimValue ??
        item.subjectID ??
        renderToString("oauth-provider." + item.providerType)
      );
    case "login_id":
      return item.claimValue;
    case "biometric":
      return item.formattedDeviceInfo
        ? item.formattedDeviceInfo
        : renderToString(
            "UserDetails.connected-identities.biometric.unknown-device"
          );
    case "anonymous":
      return renderToString(
        "UserDetails.connected-identities.anonymous.anonymous-user"
      );
    case "ldap":
      return item.lastLoginUserName ?? "";
    default:
      return "";
  }
}

function checkIsClaimVerified(
  verifiedClaims: VerifiedClaims,
  claimName: string,
  claimValue: string
) {
  const matchedClaim = verifiedClaims.find((claim) => {
    return claim.name === claimName && claim.value === claimValue;
  });

  return matchedClaim != null;
}

const VerifyButton: React.VFC<VerifyButtonProps> = function VerifyButton(
  props: VerifyButtonProps
) {
  const { verified, verifying, toggleVerified } = props;
  const loading = useIsLoading();

  const onClickVerify = useCallback(() => {
    toggleVerified(true);
  }, [toggleVerified]);

  const onClickUnverify = useCallback(() => {
    toggleVerified(false);
  }, [toggleVerified]);

  if (verified) {
    return (
      <Button
        className={cn(styles.controlButton, styles.unverifyButton)}
        size="1"
        variant="outline"
        disabled={loading || verifying}
        onClick={onClickUnverify}
        loading={verifying}
      >
        <FormattedMessage id="make-as-unverified" />
      </Button>
    );
  }

  return (
    <Button
      className={cn(styles.controlButton, styles.verifyButton)}
      size="1"
      variant="solid"
      disabled={loading || verifying}
      onClick={onClickVerify}
      loading={verifying}
    >
      <FormattedMessage id="make-as-verified" />
    </Button>
  );
};

interface BaseIdentityListCellTitleProps {
  icon?: React.ReactNode;
  as?: "ExternalLink" | "Text";
  externalLinkProps?: ExternalLinkProps;
  children?: React.ReactNode;
}

const BaseIdentityListCellTitle: React.VFC<BaseIdentityListCellTitleProps> = (
  props
) => {
  const { icon, externalLinkProps, children, as = "Text" } = props;

  return (
    <>
      <div className={styles.cellIcon}>{icon}</div>
      {as === "ExternalLink" ? (
        <ExternalLink {...externalLinkProps}>
          <Text
            size="2"
            weight="medium"
            className={cn(styles.cellName, styles.cellNameExternalLink)}
          >
            {children}
          </Text>
        </ExternalLink>
      ) : (
        <Text size="2" weight="medium" className={styles.cellName}>
          {children}
        </Text>
      )}
    </>
  );
};

interface BaseIdentityListCellDescriptionProps {
  verified?: boolean;
  children: React.ReactNode;
}

const BaseIdentityListCellDescription: React.VFC<
  BaseIdentityListCellDescriptionProps
> = (props) => {
  const { verified, children } = props;

  return (
    <Text as="div" size="2" color="gray" className={styles.cellDesc}>
      {verified != null ? (
        <>
          {verified ? (
            <Text color="green" className={styles.cellDescVerified}>
              <FormattedMessage id="verified" />
            </Text>
          ) : (
            <Text color="amber" className={styles.cellDescUnverified}>
              <FormattedMessage id="unverified" />
            </Text>
          )}
          <Text className={styles.cellDescSeparator}>{" | "}</Text>
        </>
      ) : null}
      {children}
    </Text>
  );
};

interface BaseIdentityListCellButtonGroupProps {
  identityID?: string;
  identityType: IdentityType;
  identityName?: string;
  claimName?: string;
  claimValue?: string;
  verified?: boolean;
  setVerifiedStatus?: (
    claimName: string,
    claimValue: string,
    verified: boolean
  ) => Promise<boolean>;
  onRemoveClicked?: (identityID: string, identityName: string) => void;
}

const BaseIdentityListCellButtonGroup: React.VFC<
  BaseIdentityListCellButtonGroupProps
> = (props) => {
  const {
    identityID,
    identityType,
    identityName,
    claimName,
    claimValue,
    verified,
    setVerifiedStatus,
    onRemoveClicked: _onRemoveClicked,
  } = props;

  const loading = useIsLoading();
  const [verifying, setVerifying] = useState(false);
  const onRemoveClicked = useCallback(() => {
    if (identityID == null || identityName == null) {
      return;
    }

    _onRemoveClicked?.(identityID, identityName);
  }, [identityID, identityName, _onRemoveClicked]);

  const onVerifyClicked = useCallback(
    (verified: boolean) => {
      if (claimName === undefined || claimValue === undefined) {
        return;
      }
      setVerifying(true);
      setVerifiedStatus?.(claimName, claimValue, verified).finally(() => {
        setVerifying(false);
      });
    },
    [setVerifiedStatus, claimName, claimValue]
  );

  const shouldShowVerifyButton = verified != null && setVerifiedStatus != null;

  return (
    <div className={styles.buttonGroup}>
      {shouldShowVerifyButton ? (
        <VerifyButton
          verified={verified}
          verifying={verifying}
          toggleVerified={onVerifyClicked}
        />
      ) : null}
      {removeButtonTextId[identityType] !== "" ? (
        <Button
          className={cn(styles.controlButton, styles.removeButton)}
          disabled={loading}
          size="1"
          variant="outline"
          color="red"
          onClick={onRemoveClicked}
        >
          <FormattedMessage id={removeButtonTextId[identityType]} />
        </Button>
      ) : null}
    </div>
  );
};

interface BaseIdentityListCellActionButtonProps {
  className?: string;
  identityID?: string;
  identityType: IdentityType;
  identityName?: string;
  claimName?: string;
  claimValue?: string;
  verified?: boolean;
  setVerifiedStatus?: (
    claimName: string,
    claimValue: string,
    verified: boolean
  ) => Promise<boolean>;
  onRemoveClicked?: (identityID: string, identityName: string) => void;
  onEditClicked?: (identityID: string, identityName: string) => void;
}

const BaseIdentityListCellActionButton: React.VFC<
  BaseIdentityListCellActionButtonProps
> = (props) => {
  const {
    className,
    identityID,
    identityType,
    identityName,
    claimName,
    claimValue,
    verified,
    setVerifiedStatus,
    onRemoveClicked: _onRemoveClicked,
    onEditClicked: _onEditClicked,
  } = props;

  const { renderToString } = useContext(Context);
  const loading = useIsLoading();
  const [verifying, setVerifying] = useState(false);
  const onRemoveClicked = useCallback(() => {
    if (identityID == null || identityName == null) {
      return;
    }

    _onRemoveClicked?.(identityID, identityName);
  }, [identityID, identityName, _onRemoveClicked]);

  const onEditClicked = useCallback(() => {
    if (identityID == null || identityName == null) {
      return;
    }

    _onEditClicked?.(identityID, identityName);
  }, [identityID, identityName, _onEditClicked]);

  const onVerifyClicked = useCallback(
    (verified: boolean) => {
      if (claimName === undefined || claimValue === undefined) {
        return;
      }
      setVerifying(true);
      setVerifiedStatus?.(claimName, claimValue, verified).finally(() => {
        setVerifying(false);
      });
    },
    [setVerifiedStatus, claimName, claimValue]
  );

  const shouldShowEditButton = identityType === "login_id";
  const shouldShowVerifyButton = verified != null && setVerifiedStatus != null;
  const shouldShowRemoveButton = removeButtonTextId[identityType] !== "";

  const hasActions =
    shouldShowVerifyButton || shouldShowEditButton || shouldShowRemoveButton;

  return (
    <div className={className ?? styles.actionButton}>
      {hasActions ? (
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            <IconButton
              className={listStyles.rowActionsButton}
              size="2"
              variant="soft"
              color="gray"
              disabled={loading}
              aria-label={renderToString("action")}
            >
              <DotsVerticalIcon width="1rem" height="1rem" />
            </IconButton>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            {shouldShowVerifyButton ? (
              <DropdownMenu.Item
                disabled={verifying}
                onSelect={() => onVerifyClicked(!verified)}
              >
                <FormattedMessage
                  id={verified ? "make-as-unverified" : "make-as-verified"}
                />
              </DropdownMenu.Item>
            ) : null}
            {shouldShowEditButton ? (
              <DropdownMenu.Item onSelect={onEditClicked}>
                <Pencil1Icon />
                <FormattedMessage id="edit" />
              </DropdownMenu.Item>
            ) : null}
            {shouldShowRemoveButton &&
            (shouldShowVerifyButton || shouldShowEditButton) ? (
              <DropdownMenu.Separator />
            ) : null}
            {shouldShowRemoveButton ? (
              <DropdownMenu.Item color="red" onSelect={onRemoveClicked}>
                <TrashIcon />
                <FormattedMessage id={removeButtonTextId[identityType]} />
              </DropdownMenu.Item>
            ) : null}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      ) : null}
    </div>
  );
};

interface BaseIdentityListCellProps {
  icon: React.ReactNode;
  identityID: string;
  identityType: IdentityType;
  identityName: string;
  verified?: boolean;
  connectedOn: string;
  claimName?: string;
  claimValue?: string;

  setVerifiedStatus: (
    claimName: string,
    claimValue: string,
    verified: boolean
  ) => Promise<boolean>;
  onRemoveClicked: (identityID: string, identityName: string) => void;
}

const BaseIdentityListCell: React.VFC<BaseIdentityListCellProps> = (props) => {
  const {
    icon,
    identityID,
    identityType,
    identityName,
    claimName,
    claimValue,
    verified,
    connectedOn,
    setVerifiedStatus,
    onRemoveClicked,
  } = props;

  return (
    <ListCellLayout className={styles.cellContainer}>
      <BaseIdentityListCellTitle as="Text" icon={icon}>
        {identityName}
      </BaseIdentityListCellTitle>
      <BaseIdentityListCellDescription verified={verified}>
        <FormattedMessage
          id="UserDetails.connected-identities.added-on"
          values={{ datetime: connectedOn }}
        />
      </BaseIdentityListCellDescription>
      <BaseIdentityListCellActionButton
        verified={verified}
        identityID={identityID}
        identityName={identityName}
        identityType={identityType}
        claimName={claimName}
        claimValue={claimValue}
        setVerifiedStatus={setVerifiedStatus}
        onRemoveClicked={onRemoveClicked}
      />
    </ListCellLayout>
  );
};

interface LoginIDIdentityListCellProps extends BaseIdentityListCellProps {
  loginIDKey: LoginIDKeyType;
  connectedOnDateOnly: string;
  onEditClicked: (
    identityID: string,
    loginIDKey: LoginIDKeyType,
    identityValue: string
  ) => void;
}

const LoginIDIdentityListCell: React.VFC<LoginIDIdentityListCellProps> = (
  props
) => {
  const {
    identityID,
    identityType,
    loginIDKey,
    identityName,
    claimName,
    claimValue,
    verified,
    connectedOn,
    connectedOnDateOnly,
    setVerifiedStatus,
    onRemoveClicked,
    onEditClicked: _onEditClicked,
  } = props;

  const onEditClicked = useCallback(() => {
    _onEditClicked(identityID, loginIDKey, identityName);
  }, [_onEditClicked, identityID, identityName, loginIDKey]);

  return (
    <ListCellLayout className={listStyles.row}>
      <div className={listStyles.rowValue}>
        <Text size="2" className={listStyles.rowValueText}>
          {identityName}
        </Text>
        {verified != null ? (
          <span
            className={cn(
              styles.verificationStatus,
              verified ? styles.verifiedStatus : styles.unverifiedStatus
            )}
          >
            {verified ? (
              <CheckCircledIcon className={styles.verificationIcon} />
            ) : (
              <CrossCircledIcon className={styles.verificationIcon} />
            )}
            <FormattedMessage id={verified ? "verified" : "unverified"} />
          </span>
        ) : null}
      </div>
      <Tooltip content={connectedOn}>
        <Text size="2" color="gray" className={listStyles.rowDate}>
          <FormattedMessage
            id="UserDetails.connected-identities.added-on"
            values={{ datetime: connectedOnDateOnly }}
          />
        </Text>
      </Tooltip>
      <BaseIdentityListCellActionButton
        className={listStyles.rowAction}
        verified={verified}
        identityID={identityID}
        identityName={identityName}
        identityType={identityType}
        claimName={claimName}
        claimValue={claimValue}
        setVerifiedStatus={setVerifiedStatus}
        onRemoveClicked={onRemoveClicked}
        onEditClicked={onEditClicked}
      />
    </ListCellLayout>
  );
};

interface OAuthIdentityListCellProps extends BaseIdentityListCellProps {}

const OAuthIdentityListCell: React.VFC<OAuthIdentityListCellProps> = (
  props
) => {
  const {
    icon,
    identityID,
    identityType,
    identityName,
    claimName,
    claimValue,
    verified,
    connectedOn,
    setVerifiedStatus,
    onRemoveClicked,
  } = props;

  return (
    <ListCellLayout className={styles.cellContainer}>
      <BaseIdentityListCellTitle as="Text" icon={icon}>
        {identityName}
      </BaseIdentityListCellTitle>
      <BaseIdentityListCellDescription verified={verified}>
        <FormattedMessage
          id="UserDetails.connected-identities.connected-on"
          values={{ datetime: connectedOn }}
        />
      </BaseIdentityListCellDescription>
      <BaseIdentityListCellButtonGroup
        verified={verified}
        identityID={identityID}
        identityName={identityName}
        identityType={identityType}
        claimName={claimName}
        claimValue={claimValue}
        setVerifiedStatus={setVerifiedStatus}
        onRemoveClicked={onRemoveClicked}
      />
    </ListCellLayout>
  );
};

interface LDAPIdentityListCellProps {
  icon: React.ReactNode;
  identityID: string;
  identityType: IdentityType;
  identityName: string;
  userIDAttributeName?: string;
  userIDAttributeValue?: string;
  verified?: boolean;
  connectedOn: string;
}

const LDAPIdentityListCell: React.VFC<LDAPIdentityListCellProps> = (props) => {
  const {
    icon,
    identityName,
    userIDAttributeName,
    userIDAttributeValue,
    verified,
    connectedOn,
  } = props;

  return (
    <ListCellLayout className={cn(styles.cellContainer, styles.ldap)}>
      <BaseIdentityListCellTitle as="Text" icon={icon}>
        {identityName ? identityName : "-"}
      </BaseIdentityListCellTitle>
      <Text size="2" className={styles.cellLDAPInfo}>
        {userIDAttributeName && userIDAttributeValue
          ? `${userIDAttributeName}=${userIDAttributeValue}`
          : "-"}
      </Text>
      <BaseIdentityListCellDescription verified={verified}>
        <FormattedMessage
          id="UserDetails.connected-identities.added-on"
          values={{ datetime: connectedOn }}
        />
      </BaseIdentityListCellDescription>
    </ListCellLayout>
  );
};

const UserDetailsConnectedIdentities: React.VFC<UserDetailsConnectedIdentitiesProps> =
  function UserDetailsConnectedIdentities(
    props: UserDetailsConnectedIdentitiesProps
  ) {
    const {
      identities,
      verifiedClaims,
      availableLoginIdIdentities,
      phoneInputAllowlist,
      phoneInputPinnedList,
      onIdentityCreated,
    } = props;
    const { locale, renderToString } = useContext(Context);

    const { userID } = useParams() as { userID: string };
    const navigate = useNavigate();

    /* TODO: implement save primary identities
  const [remountIdentifier, setRemountIdentifier] = useState(0);
  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);
  */

    const {
      deleteIdentity,
      loading: deletingIdentity,
      error: deleteIdentityError,
    } = useDeleteIdentityMutation();
    useLoading(deletingIdentity);
    useProvideError(deleteIdentityError);

    const {
      setVerifiedStatus,
      loading: settingVerifiedStatus,
      error: setVerifiedStatusError,
    } = useSetVerifiedStatusMutation(userID);
    useLoading(settingVerifiedStatus);
    useProvideError(setVerifiedStatusError);

    const [isConfirmationDialogVisible, setIsConfirmationDialogVisible] =
      useState(false);
    const [isAddEmailDialogOpen, setIsAddEmailDialogOpen] = useState(false);
    const [emailIdentityToEdit, setEmailIdentityToEdit] = useState<{
      id: string;
      value: string;
    }>();
    const [isAddPhoneDialogOpen, setIsAddPhoneDialogOpen] = useState(false);
    const [isAddUsernameDialogOpen, setIsAddUsernameDialogOpen] =
      useState(false);
    const [usernameIdentityToEdit, setUsernameIdentityToEdit] = useState<{
      id: string;
      value: string;
    }>();

    const [confirmationDialogData, setConfirmationDialogData] =
      useState<ConfirmationDialogData>({
        identityID: "",
        identityName: "",
      });

    const identityLists: IdentityLists = useMemo(() => {
      const oauthIdentityList: OAuthIdentityListItem[] = [];
      const emailIdentityList: LoginIDIdentityListItem[] = [];
      const phoneIdentityList: LoginIDIdentityListItem[] = [];
      const usernameIdentityList: LoginIDIdentityListItem[] = [];
      const biometricIdentityList: BiometricIdentityListItem[] = [];
      const anonymousIdentityList: AnonymousIdentityListItem[] = [];
      const ldapIdentityList: LDAPIdentityListItem[] = [];

      for (const identity of identities) {
        const createdAtStr = formatDatetime(locale, identity.createdAt) ?? "";
        const createdAtDateOnlyStr =
          formatDateOnly(locale, identity.createdAt) ?? "";
        if (identity.type === "OAUTH") {
          const providerType =
            identity.claims["https://authgear.com/claims/oauth/provider_type"]!;
          const subjectID =
            identity.claims["https://authgear.com/claims/oauth/subject_id"];

          const claimName = "email";
          const claimValue = identity.claims.email;

          oauthIdentityList.push({
            id: identity.id,
            type: "oauth",
            providerType,
            subjectID,
            claimName,
            claimValue,
            verified:
              claimValue == null
                ? undefined
                : checkIsClaimVerified(verifiedClaims, claimName, claimValue),
            connectedOn: createdAtStr,
          });
        }

        if (identity.type === "LOGIN_ID") {
          if (
            identity.claims["https://authgear.com/claims/login_id/type"] ===
            "email"
          ) {
            const claimName = "email";
            const claimValue = identity.claims.email!;

            emailIdentityList.push({
              id: identity.id,
              type: "login_id",
              loginIDKey: "email",
              claimName,
              claimValue,
              verified: checkIsClaimVerified(
                verifiedClaims,
                claimName,
                claimValue
              ),
              connectedOn: createdAtStr,
              connectedOnDateOnly: createdAtDateOnlyStr,
            });
          }

          if (
            identity.claims["https://authgear.com/claims/login_id/type"] ===
            "phone"
          ) {
            const claimName = "phone_number";
            const claimValue = identity.claims.phone_number!;

            phoneIdentityList.push({
              id: identity.id,
              type: "login_id",
              loginIDKey: "phone",
              claimName,
              claimValue,
              verified: checkIsClaimVerified(
                verifiedClaims,
                claimName,
                claimValue
              ),
              connectedOn: createdAtStr,
              connectedOnDateOnly: createdAtDateOnlyStr,
            });
          }

          if (
            identity.claims["https://authgear.com/claims/login_id/type"] ===
            "username"
          ) {
            usernameIdentityList.push({
              id: identity.id,
              type: "login_id",
              loginIDKey: "username",
              claimName: "preferred_username",
              claimValue: identity.claims.preferred_username!,
              connectedOn: createdAtStr,
              connectedOnDateOnly: createdAtDateOnlyStr,
            });
          }
        }

        if (identity.type === "BIOMETRIC") {
          const info =
            identity.claims[
              "https://authgear.com/claims/biometric/formatted_device_info"
            ];
          const formattedDeviceInfo = typeof info === "string" ? info : "";
          biometricIdentityList.push({
            id: identity.id,
            type: "biometric",
            connectedOn: createdAtStr,
            verified: undefined,
            formattedDeviceInfo: formattedDeviceInfo,
          });
        }
        if (identity.type === "ANONYMOUS") {
          anonymousIdentityList.push({
            id: identity.id,
            type: "anonymous",
            verified: undefined,
            connectedOn: createdAtStr,
          });
        }
        if (identity.type === "LDAP") {
          ldapIdentityList.push({
            id: identity.id,
            type: "ldap",
            verified: undefined,
            connectedOn: createdAtStr,
            lastLoginUserName:
              identity.claims[
                "https://authgear.com/claims/ldap/last_login_username"
              ],
            userIDAttributeName:
              identity.claims[
                "https://authgear.com/claims/ldap/user_id_attribute_name"
              ] ?? "",
            userIDAttributeValue:
              identity.claims[
                "https://authgear.com/claims/ldap/user_id_attribute_value"
              ] ?? "",
          });
        }
      }
      return {
        oauth: oauthIdentityList,
        email: emailIdentityList,
        phone: phoneIdentityList,
        username: usernameIdentityList,
        biometric: biometricIdentityList,
        anonymous: anonymousIdentityList,
        ldap: ldapIdentityList,
      };
    }, [identities, locale, verifiedClaims]);

    const onRemoveClicked = useCallback(
      (identityID: string, identityName: string) => {
        setConfirmationDialogData({
          identityID,
          identityName,
        });
        setIsConfirmationDialogVisible(true);
      },
      [setConfirmationDialogData]
    );

    const onEditLoginIDClicked = useCallback(
      (
        identityID: string,
        loginIDKey: LoginIDKeyType,
        identityValue: string
      ) => {
        switch (loginIDKey) {
          case "username":
            setUsernameIdentityToEdit({
              id: identityID,
              value: identityValue,
            });
            setIsAddUsernameDialogOpen(true);
            break;
          case "phone":
            navigate(generatePath("./edit-phone/:identityID", { identityID }));
            break;
          case "email":
            setEmailIdentityToEdit({ id: identityID, value: identityValue });
            setIsAddEmailDialogOpen(true);
            break;
          default:
            console.error(
              new Error(`Unexpected loginIDKey ${loginIDKey as string}`)
            );
            break;
        }
      },
      [navigate]
    );

    const onDismissConfirmationDialog = useCallback(() => {
      if (!deletingIdentity) {
        setIsConfirmationDialogVisible(false);
      }
    }, [deletingIdentity]);

    const onConfirmRemoveIdentity = useCallback(() => {
      const { identityID } = confirmationDialogData;
      deleteIdentity(identityID).finally(() => {
        onDismissConfirmationDialog();
      });
    }, [confirmationDialogData, deleteIdentity, onDismissConfirmationDialog]);

    const onRenderIdentityCell = useCallback(
      (item?: IdentityListItem, _index?: number): React.ReactNode => {
        if (item == null) {
          return null;
        }

        const identityID = item.id;
        const identityType = item.type;
        const identityName = getIdentityName(item, renderToString);
        const verified = item.verified;
        const connectedOn = item.connectedOn;

        switch (item.type) {
          case "login_id":
            return (
              <LoginIDIdentityListCell
                icon={loginIdIconMap[item.loginIDKey]}
                loginIDKey={item.loginIDKey}
                identityID={identityID}
                identityType={identityType}
                identityName={identityName}
                claimName={item.claimName}
                claimValue={item.claimValue}
                verified={verified}
                connectedOn={connectedOn}
                connectedOnDateOnly={item.connectedOnDateOnly}
                setVerifiedStatus={setVerifiedStatus}
                onRemoveClicked={onRemoveClicked}
                onEditClicked={onEditLoginIDClicked}
              />
            );
          case "oauth":
            return (
              <OAuthIdentityListCell
                icon={oauthIconMap[item.providerType]}
                identityID={identityID}
                identityType={identityType}
                identityName={identityName}
                claimName={item.claimName}
                claimValue={item.claimValue}
                verified={verified}
                connectedOn={connectedOn}
                setVerifiedStatus={setVerifiedStatus}
                onRemoveClicked={onRemoveClicked}
              />
            );
          case "biometric":
            return (
              <BaseIdentityListCell
                icon={biometricIcon}
                identityID={identityID}
                identityType={identityType}
                identityName={identityName}
                verified={verified}
                connectedOn={connectedOn}
                setVerifiedStatus={setVerifiedStatus}
                onRemoveClicked={onRemoveClicked}
              />
            );
          case "anonymous":
            return (
              <BaseIdentityListCell
                icon={anonymousIcon}
                identityID={identityID}
                identityType={identityType}
                identityName={identityName}
                verified={verified}
                connectedOn={connectedOn}
                setVerifiedStatus={setVerifiedStatus}
                onRemoveClicked={onRemoveClicked}
              />
            );
          case "ldap":
            return (
              <LDAPIdentityListCell
                icon={ldapIcon}
                identityID={identityID}
                identityType={identityType}
                identityName={identityName}
                userIDAttributeName={item.userIDAttributeName}
                userIDAttributeValue={item.userIDAttributeValue}
                verified={verified}
                connectedOn={connectedOn}
              />
            );
          default:
            return null;
        }
      },
      [renderToString, setVerifiedStatus, onRemoveClicked, onEditLoginIDClicked]
    );

    // Sections appear in Email → Phone number → User name order.
    // Show a section when the login ID type is enabled in the project, or
    // when the user already has such an identity (it may have been added
    // before the configuration changed).
    const loginIdentityTypesToShow = useMemo(
      () =>
        loginIdIdentityTypes.filter(
          (type) =>
            availableLoginIdIdentities.includes(type) ||
            identityLists[type].length > 0
        ),
      [availableLoginIdIdentities, identityLists]
    );

    const confirmationDialogContentProps = useMemo(() => {
      return {
        title: (
          <FormattedMessage id="UserDetails.connected-identities.confirm-remove-identity-title" />
        ),
        subText: renderToString(
          "UserDetails.connected-identities.confirm-remove-identity-message",
          { identityName: confirmationDialogData.identityName }
        ),
      };
    }, [confirmationDialogData, renderToString]);

    return (
      <div className={styles.root}>
        <ConfirmationDialog
          open={isConfirmationDialogVisible}
          onOpenChange={(open) => {
            if (!open) {
              onDismissConfirmationDialog();
            }
          }}
          title={confirmationDialogContentProps.title}
          description={confirmationDialogContentProps.subText}
          confirmText={<FormattedMessage id="confirm" />}
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={onConfirmRemoveIdentity}
          onCancel={onDismissConfirmationDialog}
          loading={deletingIdentity}
          confirmColor="red"
        />
        <AddEmailDialog
          open={isAddEmailDialogOpen}
          userID={userID}
          identityToEdit={emailIdentityToEdit}
          onOpenChange={(open) => {
            setIsAddEmailDialogOpen(open);
            if (!open) {
              setEmailIdentityToEdit(undefined);
            }
          }}
          onCreated={onIdentityCreated}
        />
        <AddPhoneDialog
          open={isAddPhoneDialogOpen}
          userID={userID}
          phoneInputAllowlist={phoneInputAllowlist}
          phoneInputPinnedList={phoneInputPinnedList}
          onOpenChange={setIsAddPhoneDialogOpen}
          onCreated={onIdentityCreated}
        />
        <AddUsernameDialog
          open={isAddUsernameDialogOpen}
          userID={userID}
          identityToEdit={usernameIdentityToEdit}
          onOpenChange={(open) => {
            setIsAddUsernameDialogOpen(open);
            if (!open) {
              setUsernameIdentityToEdit(undefined);
            }
          }}
          onCreated={onIdentityCreated}
        />
        <section className={styles.headerSection}>
          <Text as="p" size="3" weight="medium" className={styles.header}>
            <FormattedMessage id="UserDetails.connected-identities.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.description}>
            <FormattedMessage id="UserDetails.connected-identities.description" />
          </Text>
        </section>
        <section className={styles.content}>
          <section className={styles.identityLists}>
            {loginIdentityTypesToShow.map((type) => (
              <div className={styles.loginIdentityGroup} key={type}>
                <div className={listStyles.table}>
                  <div className={listStyles.tableHeader}>
                    <Text
                      as="p"
                      size="2"
                      weight="medium"
                      className={listStyles.tableTitle}
                    >
                      <FormattedMessage
                        id={`UserDetails.connected-identities.${type}`}
                      />
                    </Text>
                  </div>
                  {identityLists[type].map((item) => (
                    <React.Fragment key={item.id}>
                      {onRenderIdentityCell(item)}
                    </React.Fragment>
                  ))}
                </div>
                {availableLoginIdIdentities.includes(type) ? (
                  <button
                    type="button"
                    className={listStyles.addButton}
                    onClick={() => {
                      if (type === "email") {
                        setEmailIdentityToEdit(undefined);
                        setIsAddEmailDialogOpen(true);
                      } else if (type === "phone") {
                        setIsAddPhoneDialogOpen(true);
                      } else {
                        setUsernameIdentityToEdit(undefined);
                        setIsAddUsernameDialogOpen(true);
                      }
                    }}
                  >
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage
                      id={`UserDetails.connected-identities.add-${type}`}
                    />
                  </button>
                ) : null}
              </div>
            ))}
            {identityLists.oauth.length > 0 ? (
              <div>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.subHeader}
                >
                  <FormattedMessage id="UserDetails.connected-identities.oauth" />
                </Text>
                {identityLists.oauth.map((item) => (
                  <React.Fragment key={item.id}>
                    {onRenderIdentityCell(item)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
            {identityLists.biometric.length > 0 ? (
              <div>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.subHeader}
                >
                  <FormattedMessage id="UserDetails.connected-identities.biometric" />
                </Text>
                {identityLists.biometric.map((item) => (
                  <React.Fragment key={item.id}>
                    {onRenderIdentityCell(item)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
            {identityLists.anonymous.length > 0 ? (
              <div>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.subHeader}
                >
                  <FormattedMessage id="UserDetails.connected-identities.anonymous" />
                </Text>
                {identityLists.anonymous.map((item) => (
                  <React.Fragment key={item.id}>
                    {onRenderIdentityCell(item)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
            {identityLists.ldap.length > 0 ? (
              <div>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.subHeader}
                >
                  <FormattedMessage id="UserDetails.connected-identities.ldap" />
                </Text>
                {identityLists.ldap.map((item) => (
                  <React.Fragment key={item.id}>
                    {onRenderIdentityCell(item)}
                  </React.Fragment>
                ))}
              </div>
            ) : null}
          </section>
        </section>
      </div>
    );
  };

export default UserDetailsConnectedIdentities;
