import React, {
  useMemo,
  useCallback,
  useContext,
  useState,
  createContext,
} from "react";
import cn from "classnames";
import { useNavigate, useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  Icon,
  IContextualMenuProps,
  List,
  PrimaryButton,
  Text,
} from "@fluentui/react";

// import PrimaryIdentitiesSelectionForm from "./PrimaryIdentitiesSelectionForm";
import ButtonWithLoading from "../../ButtonWithLoading";
import ListCellLayout from "../../ListCellLayout";
import ErrorDialog from "../../error/ErrorDialog";
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import { useSetVerifiedStatusMutation } from "./mutations/setVerifiedStatusMutation";
import { formatDatetime } from "../../util/formatDatetime";
import { OAuthSSOProviderType } from "../../types";
import { UserQuery_node_User_verifiedClaims } from "./query/__generated__/UserQuery";

import styles from "./UserDetailsConnectedIdentities.module.scss";
import { useSystemConfig } from "../../context/SystemConfigContext";

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
  preferred_username?: string;
  "https://authgear.com/claims/oauth/provider_type"?: OAuthSSOProviderType;
  "https://authgear.com/claims/login_id/type"?: LoginIDIdentityType;
}

interface Identity {
  id: string;
  type: "ANONYMOUS" | "LOGIN_ID" | "OAUTH" | "BIOMETRIC";
  claims: IdentityClaim;
  createdAt: string;
  updatedAt: string;
}

type VerifiedClaim = UserQuery_node_User_verifiedClaims;
interface UserDetailsConnectedIdentitiesProps {
  identities: Identity[];
  verifiedClaims: VerifiedClaim[];
  availableLoginIdIdentities: string[];
}

const loginIdIdentityTypes = ["email", "phone", "username"] as const;
type LoginIDIdentityType = typeof loginIdIdentityTypes[number];
type IdentityType = "login_id" | "oauth" | "biometric";

type IdentityListItem =
  | OAuthIdentityListItem
  | LoginIDIdentityListItem
  | BiometricIdentityListItem;
interface OAuthIdentityListItem {
  id: string;
  type: "oauth";
  providerType: OAuthSSOProviderType;
  claimName: string;
  claimValue: string;
  verified: boolean;
  connectedOn: string;
}

interface LoginIDIdentityListItem {
  id: string;
  type: "login_id";
  loginIDKey: "email" | "phone" | "username";
  claimName: string;
  claimValue: string;
  verified?: boolean;
  connectedOn: string;
}

interface BiometricIdentityListItem {
  id: string;
  type: "biometric";
  connectedOn: string;
  verified: undefined;
}

export interface IdentityLists {
  oauth: OAuthIdentityListItem[];
  email: LoginIDIdentityListItem[];
  phone: LoginIDIdentityListItem[];
  username: LoginIDIdentityListItem[];
  biometric: BiometricIdentityListItem[];
}

interface IdentityListCellProps {
  identityID: string;
  identityType: IdentityType;
  icon: React.ReactNode;
  claimName?: string;
  claimValue?: string;
  identityName: string;
  connectedOn: string;
  verified?: boolean;
  setVerifiedStatus?: (
    claimName: string,
    claimValue: string,
    verified: boolean
  ) => Promise<boolean>;
  onRemoveClicked: (identityID: string, identityName: string) => void;
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

const oauthIconMap: Record<OAuthSSOProviderType, React.ReactNode> = {
  apple: <i className={cn("fab", "fa-apple", styles.widgetLabelIcon)} />,
  google: <i className={cn("fab", "fa-google", styles.widgetLabelIcon)} />,
  facebook: <i className={cn("fab", "fa-facebook", styles.widgetLabelIcon)} />,
  linkedin: <i className={cn("fab", "fa-linkedin", styles.widgetLabelIcon)} />,
  azureadv2: (
    <i className={cn("fab", "fa-microsoft", styles.widgetLabelIcon)} />
  ),
  wechat: <i className={cn("fab", "fa-weixin", styles.widgetLabelIcon)} />,
};

const loginIdIconMap: Record<LoginIDIdentityType, React.ReactNode> = {
  email: <Icon iconName="Mail" />,
  phone: <Icon iconName="CellPhone" />,
  username: <Icon iconName="Accounts" />,
};

const biometricIcon: React.ReactNode = <Icon iconName="Fingerprint" />;

const removeButtonTextId: Record<IdentityType, "remove" | "disconnect"> = {
  oauth: "disconnect",
  login_id: "remove",
  biometric: "remove",
};

function getIcon(item: IdentityListItem) {
  if (item.type === "oauth") {
    return oauthIconMap[item.providerType];
  }
  if (item.type === "biometric") {
    return biometricIcon;
  }
  return loginIdIconMap[item.loginIDKey];
}

function getClaimName(item: IdentityListItem): string | undefined {
  if (item.type === "biometric") {
    return undefined;
  }

  return item.claimName;
}

function getClaimValue(item: IdentityListItem): string | undefined {
  if (item.type === "biometric") {
    return undefined;
  }
  return item.claimValue;
}

function getIdentityName(
  item: IdentityListItem,
  renderToString: (id: string) => string
): string {
  if (item.type === "biometric") {
    return renderToString("UserDetails.connected-identities.biometric");
  }
  return item.claimValue;
}

function checkIsClaimVerified(
  verifiedClaims: VerifiedClaim[],
  claimName: string,
  claimValue: string
) {
  const matchedClaim = verifiedClaims.find((claim) => {
    return claim.name === claimName && claim.value === claimValue;
  });

  return matchedClaim != null;
}

const ConnectedIdentitiesMutationLoadingContext = createContext({
  settingVerifiedStatus: false,
  deletingIdentity: false,
});

const VerifyButton: React.FC<VerifyButtonProps> = function VerifyButton(
  props: VerifyButtonProps
) {
  const { verified, verifying, toggleVerified } = props;
  const { themes } = useSystemConfig();
  const { settingVerifiedStatus } = useContext(
    ConnectedIdentitiesMutationLoadingContext
  );

  const onClickVerify = useCallback(() => {
    toggleVerified(true);
  }, [toggleVerified]);

  const onClickUnverify = useCallback(() => {
    toggleVerified(false);
  }, [toggleVerified]);

  if (verified) {
    return (
      <ButtonWithLoading
        className={cn(styles.controlButton, styles.unverifyButton)}
        disabled={settingVerifiedStatus}
        theme={themes.defaultButton}
        onClick={onClickUnverify}
        labelId="make-as-unverified"
        loading={verifying}
      />
    );
  }

  return (
    <ButtonWithLoading
      className={cn(styles.controlButton, styles.verifyButton)}
      disabled={settingVerifiedStatus}
      theme={themes.verifyButton}
      onClick={onClickVerify}
      loading={verifying}
      labelId="make-as-verified"
    />
  );
};

const IdentityListCell: React.FC<IdentityListCellProps> = function IdentityListCell(
  props: IdentityListCellProps
) {
  const {
    identityID,
    identityType,
    icon,
    claimName,
    claimValue,
    identityName,
    connectedOn,
    verified,
    setVerifiedStatus,
    onRemoveClicked: _onRemoveClicked,
  } = props;

  const { themes } = useSystemConfig();
  const { settingVerifiedStatus } = useContext(
    ConnectedIdentitiesMutationLoadingContext
  );
  const [verifying, setVerifying] = useState(false);
  const onRemoveClicked = useCallback(() => {
    _onRemoveClicked(identityID, identityName);
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

  return (
    <ListCellLayout className={styles.cellContainer}>
      <div className={styles.cellIcon}>{icon}</div>
      <Text className={styles.cellName}>{identityName}</Text>
      {verified != null && (
        <>
          {verified ? (
            <Text className={styles.cellDescVerified}>
              <FormattedMessage id="verified" />
            </Text>
          ) : (
            <Text className={styles.cellDescUnverified}>
              <FormattedMessage id="unverified" />
            </Text>
          )}
          <Text className={styles.cellDescSeparator}>{" | "}</Text>
        </>
      )}
      <Text className={styles.cellDesc}>
        {identityType === "oauth" && (
          <FormattedMessage
            id="UserDetails.connected-identities.connected-on"
            values={{ datetime: connectedOn }}
          />
        )}
        {identityType === "login_id" && (
          <FormattedMessage
            id="UserDetails.connected-identities.added-on"
            values={{ datetime: connectedOn }}
          />
        )}
        {identityType === "biometric" && (
          <FormattedMessage
            id="UserDetails.connected-identities.added-on"
            values={{ datetime: connectedOn }}
          />
        )}
      </Text>
      {verified != null && setVerifiedStatus != null && (
        <VerifyButton
          verified={verified}
          verifying={verifying}
          toggleVerified={onVerifyClicked}
        />
      )}
      <DefaultButton
        className={cn(styles.controlButton, styles.removeButton)}
        disabled={settingVerifiedStatus}
        theme={themes.destructive}
        onClick={onRemoveClicked}
      >
        <FormattedMessage id={removeButtonTextId[identityType]} />
      </DefaultButton>
    </ListCellLayout>
  );
};

const UserDetailsConnectedIdentities: React.FC<UserDetailsConnectedIdentitiesProps> = function UserDetailsConnectedIdentities(
  props: UserDetailsConnectedIdentitiesProps
) {
  const { identities, verifiedClaims, availableLoginIdIdentities } = props;
  const { locale, renderToString } = useContext(Context);

  const { userID } = useParams();
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
  const {
    setVerifiedStatus,
    loading: settingVerifiedStatus,
    error: setVerifiedStatusError,
  } = useSetVerifiedStatusMutation(userID);

  const [
    isConfirmationDialogVisible,
    setIsConfirmationDialogVisible,
  ] = useState(false);

  const [
    confirmationDialogData,
    setConfirmationDialogData,
  ] = useState<ConfirmationDialogData>({
    identityID: "",
    identityName: "",
  });

  const identityLists: IdentityLists = useMemo(() => {
    const oauthIdentityList: OAuthIdentityListItem[] = [];
    const emailIdentityList: LoginIDIdentityListItem[] = [];
    const phoneIdentityList: LoginIDIdentityListItem[] = [];
    const usernameIdentityList: LoginIDIdentityListItem[] = [];
    const biometricIdentityList: BiometricIdentityListItem[] = [];

    for (const identity of identities) {
      const createdAtStr = formatDatetime(locale, identity.createdAt) ?? "";
      if (identity.type === "OAUTH") {
        const providerType = identity.claims[
          "https://authgear.com/claims/oauth/provider_type"
        ]!;

        const claimName = "email";
        const claimValue = identity.claims.email!;

        oauthIdentityList.push({
          id: identity.id,
          type: "oauth",
          claimName,
          claimValue,
          providerType: providerType,
          verified: checkIsClaimVerified(verifiedClaims, claimName, claimValue),
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
          });
        }
      }

      if (identity.type === "BIOMETRIC") {
        biometricIdentityList.push({
          id: identity.id,
          type: "biometric",
          connectedOn: createdAtStr,
          verified: undefined,
        });
      }
    }
    return {
      oauth: oauthIdentityList,
      email: emailIdentityList,
      phone: phoneIdentityList,
      username: usernameIdentityList,
      biometric: biometricIdentityList,
    };
  }, [locale, identities, verifiedClaims]);

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

      const icon = getIcon(item);
      return (
        <IdentityListCell
          identityID={item.id}
          identityType={item.type}
          icon={icon}
          claimName={getClaimName(item)}
          claimValue={getClaimValue(item)}
          identityName={getIdentityName(item, renderToString)}
          verified={item.verified}
          connectedOn={item.connectedOn}
          onRemoveClicked={onRemoveClicked}
          setVerifiedStatus={setVerifiedStatus}
        />
      );
    },
    [onRemoveClicked, setVerifiedStatus, renderToString]
  );

  const addIdentitiesMenuProps: IContextualMenuProps = useMemo(() => {
    const availableMenuItem = [
      {
        key: "email",
        text: renderToString("UserDetails.connected-identities.email"),
        iconProps: { iconName: "Mail" },
        onClick: () => navigate("./add-email"),
      },
      {
        key: "phone",
        text: renderToString("UserDetails.connected-identities.phone"),
        iconProps: { iconName: "CellPhone" },
        onClick: () => navigate("./add-phone"),
      },
      {
        key: "username",
        text: renderToString("UserDetails.connected-identities.username"),
        iconProps: { iconName: "Accounts" },
        onClick: () => navigate("./add-username"),
      },
    ];
    const enabledItems = availableMenuItem.filter((item) => {
      return availableLoginIdIdentities.includes(item.key);
    });
    return {
      items: enabledItems,
      directionalHintFixed: true,
    };
  }, [renderToString, navigate, availableLoginIdIdentities]);

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
    <ConnectedIdentitiesMutationLoadingContext.Provider
      value={{ settingVerifiedStatus, deletingIdentity }}
    >
      <div className={styles.root}>
        <Dialog
          hidden={!isConfirmationDialogVisible}
          dialogContentProps={confirmationDialogContentProps}
          modalProps={{ isBlocking: deletingIdentity }}
          onDismiss={onDismissConfirmationDialog}
        >
          <DialogFooter>
            <ButtonWithLoading
              labelId="confirm"
              onClick={onConfirmRemoveIdentity}
              loading={deletingIdentity}
              disabled={!isConfirmationDialogVisible}
            />
            <DefaultButton
              disabled={deletingIdentity || !isConfirmationDialogVisible}
              onClick={onDismissConfirmationDialog}
            >
              <FormattedMessage id="cancel" />
            </DefaultButton>
          </DialogFooter>
        </Dialog>
        <ErrorDialog
          error={deleteIdentityError}
          rules={[
            {
              reason: "InvariantViolated",
              kind: "RemoveLastIdentity",
              errorMessageID:
                "UserDetails.connected-identities.remove-identity-error.connot-remove-last",
            },
          ]}
          fallbackErrorMessageID="UserDetails.connected-identities.remove-identity-error.generic"
        />
        <ErrorDialog
          error={setVerifiedStatusError}
          rules={[]}
          fallbackErrorMessageID="UserDetails.connected-identities.verify-identity-error.generic"
        />
        <section className={styles.headerSection}>
          <Text as="h2" className={styles.header}>
            <FormattedMessage id="UserDetails.connected-identities.title" />
          </Text>
          <PrimaryButton
            disabled={addIdentitiesMenuProps.items.length === 0}
            iconProps={{ iconName: "CirclePlus" }}
            menuProps={addIdentitiesMenuProps}
            styles={{
              menuIcon: { paddingLeft: "3px" },
              icon: { paddingRight: "3px" },
            }}
          >
            <FormattedMessage id="UserDetails.connected-identities.add-identity" />
          </PrimaryButton>
        </section>
        <section className={styles.identityLists}>
          {identityLists.oauth.length > 0 && (
            <>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.oauth" />
              </Text>
              <List
                className={styles.list}
                items={identityLists.oauth}
                onRenderCell={onRenderIdentityCell}
              />
            </>
          )}
          {identityLists.email.length > 0 && (
            <>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.email" />
              </Text>
              <List
                className={styles.list}
                items={identityLists.email}
                onRenderCell={onRenderIdentityCell}
              />
            </>
          )}
          {identityLists.phone.length > 0 && (
            <>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.phone" />
              </Text>
              <List
                className={styles.list}
                items={identityLists.phone}
                onRenderCell={onRenderIdentityCell}
              />
            </>
          )}
          {identityLists.username.length > 0 && (
            <>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.username" />
              </Text>
              <List
                className={styles.list}
                items={identityLists.username}
                onRenderCell={onRenderIdentityCell}
              />
            </>
          )}
          {identityLists.biometric.length > 0 && (
            <>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.biometric" />
              </Text>
              <List
                className={styles.list}
                items={identityLists.biometric}
                onRenderCell={onRenderIdentityCell}
              />
            </>
          )}
        </section>
        {/* TODO: implement primary identities mutation
        <Text as="h2" className={styles.primaryIdentitiesTitle}>
          <FormattedMessage id="UserDetails.connected-identities.primary-identities.title" />
        </Text>
        <PrimaryIdentitiesSelectionForm
          key={remountIdentifier}
          className={styles.primaryIdentitiesForm}
          identityLists={identityLists}
          resetForm={resetForm}
        />
        */}
      </div>
    </ConnectedIdentitiesMutationLoadingContext.Provider>
  );
};

export default UserDetailsConnectedIdentities;
