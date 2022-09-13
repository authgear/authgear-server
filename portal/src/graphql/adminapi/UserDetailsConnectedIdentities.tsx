import React, { useMemo, useCallback, useContext, useState } from "react";
import cn from "classnames";
import { useNavigate, useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  Dialog,
  DialogFooter,
  Icon,
  IContextualMenuProps,
  List,
  Text,
} from "@fluentui/react";

// import PrimaryIdentitiesSelectionForm from "./PrimaryIdentitiesSelectionForm";
import ButtonWithLoading from "../../ButtonWithLoading";
import ListCellLayout from "../../ListCellLayout";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import { useSetVerifiedStatusMutation } from "./mutations/setVerifiedStatusMutation";
import { formatDatetime } from "../../util/formatDatetime";
import { OAuthSSOProviderType } from "../../types";
import { UserQueryNodeFragment } from "./query/userQuery.generated";

import styles from "./UserDetailsConnectedIdentities.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useIsLoading, useLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";

// Always disable virtualization for List component, as it wont work properly with mobile view
const onShouldVirtualize = () => {
  return false;
};

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
  preferred_username?: string;
  "https://authgear.com/claims/oauth/provider_type"?: OAuthSSOProviderType;
  "https://authgear.com/claims/login_id/type"?: LoginIDIdentityType;
}

interface Identity {
  id: string;
  type: "ANONYMOUS" | "LOGIN_ID" | "OAUTH" | "BIOMETRIC" | "PASSKEY" | "SIWE";
  claims: IdentityClaim;
  createdAt: string;
  updatedAt: string;
}

type VerifiedClaims = UserQueryNodeFragment["verifiedClaims"];
interface UserDetailsConnectedIdentitiesProps {
  identities: Identity[];
  verifiedClaims: VerifiedClaims;
  availableLoginIdIdentities: string[];
}

const loginIdIdentityTypes = ["email", "phone", "username"] as const;
type LoginIDIdentityType = typeof loginIdIdentityTypes[number];
type IdentityType = "login_id" | "oauth" | "biometric" | "anonymous" | "siwe";

type IdentityListItem =
  | OAuthIdentityListItem
  | LoginIDIdentityListItem
  | BiometricIdentityListItem
  | AnonymousIdentityListItem
  | SIWEIdentityListItem;
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
  formattedDeviceInfo: string;
}

interface AnonymousIdentityListItem {
  id: string;
  type: "anonymous";
  verified: undefined;
  connectedOn: string;
}

interface SIWEIdentityListItem {
  id: string;
  type: "siwe";
  address: string;
  chainId: number;
  verified: undefined;
  connectedOn: string;
}

export interface IdentityLists {
  oauth: OAuthIdentityListItem[];
  email: LoginIDIdentityListItem[];
  phone: LoginIDIdentityListItem[];
  username: LoginIDIdentityListItem[];
  biometric: BiometricIdentityListItem[];
  anonymous: AnonymousIdentityListItem[];
  siwe: SIWEIdentityListItem[];
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
  email: <Icon iconName="Mail" />,
  phone: <Icon iconName="CellPhone" />,
  username: <Icon iconName="Accounts" />,
};

const biometricIcon: React.ReactNode = <Icon iconName="Fingerprint" />;
const anonymousIcon: React.ReactNode = <Icon iconName="People" />;
const siweIcon: React.ReactNode = <i className={cn("fab", "fa-ethereum")} />;

const removeButtonTextId: Record<IdentityType, "remove" | "disconnect" | ""> = {
  oauth: "disconnect",
  login_id: "remove",
  biometric: "remove",
  anonymous: "",
  siwe: "",
};

function getIcon(item: IdentityListItem) {
  if (item.type === "oauth") {
    return oauthIconMap[item.providerType];
  }
  if (item.type === "biometric") {
    return biometricIcon;
  }
  if (item.type === "anonymous") {
    return anonymousIcon;
  }
  if (item.type === "siwe") {
    return siweIcon;
  }
  return loginIdIconMap[item.loginIDKey];
}

function getClaimName(item: IdentityListItem): string | undefined {
  if (item.type === "biometric") {
    return undefined;
  }
  if (item.type === "anonymous") {
    return undefined;
  }
  if (item.type === "siwe") {
    return undefined;
  }
  return item.claimName;
}

function getClaimValue(item: IdentityListItem): string | undefined {
  if (item.type === "biometric") {
    return undefined;
  }
  if (item.type === "anonymous") {
    return undefined;
  }
  if (item.type === "siwe") {
    return undefined;
  }
  return item.claimValue;
}

function getIdentityName(
  item: IdentityListItem,
  renderToString: (id: string) => string
): string {
  if (item.type === "biometric") {
    return item.formattedDeviceInfo
      ? item.formattedDeviceInfo
      : renderToString(
          "UserDetails.connected-identities.biometric.unknown-device"
        );
  }
  if (item.type === "anonymous") {
    return renderToString(
      "UserDetails.connected-identities.anonymous.anonymous-user"
    );
  }
  if (item.type === "siwe") {
    return item.address;
  }
  return item.claimValue;
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
  const { themes } = useSystemConfig();
  const loading = useIsLoading();

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
        disabled={loading}
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
      disabled={loading}
      theme={themes.verifyButton}
      onClick={onClickVerify}
      loading={verifying}
      labelId="make-as-verified"
    />
  );
};

const IdentityListCell: React.VFC<IdentityListCellProps> =
  // eslint-disable-next-line complexity
  function IdentityListCell(props: IdentityListCellProps) {
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
    const loading = useIsLoading();
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

    const shouldShowVerifyButton =
      verified != null && setVerifiedStatus != null;

    return (
      <ListCellLayout className={styles.cellContainer}>
        <div className={styles.cellIcon}>{icon}</div>
        <Text className={styles.cellName}>{identityName}</Text>
        <Text className={styles.cellDesc}>
          {verified != null ? (
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
              <Text block={true} className={styles.cellDescSeparator}>
                {" | "}
              </Text>
            </>
          ) : null}
          {identityType === "oauth" ? (
            <FormattedMessage
              id="UserDetails.connected-identities.connected-on"
              values={{ datetime: connectedOn }}
            />
          ) : null}
          {identityType === "login_id" ? (
            <FormattedMessage
              id="UserDetails.connected-identities.added-on"
              values={{ datetime: connectedOn }}
            />
          ) : null}
          {identityType === "biometric" ? (
            <FormattedMessage
              id="UserDetails.connected-identities.added-on"
              values={{ datetime: connectedOn }}
            />
          ) : null}
          {identityType === "anonymous" ? (
            <FormattedMessage
              id="UserDetails.connected-identities.added-on"
              values={{ datetime: connectedOn }}
            />
          ) : null}
          {identityType === "siwe" ? (
            <FormattedMessage
              id="UserDetails.connected-identities.added-on"
              values={{ datetime: connectedOn }}
            />
          ) : null}
        </Text>
        <div className={styles.buttonGroup}>
          {shouldShowVerifyButton ? (
            <VerifyButton
              verified={verified}
              verifying={verifying}
              toggleVerified={onVerifyClicked}
            />
          ) : null}
          {removeButtonTextId[identityType] !== "" ? (
            <DefaultButton
              className={cn(
                styles.controlButton,
                styles.removeButton,
                shouldShowVerifyButton ? "" : styles.removeButtonFull
              )}
              disabled={loading}
              theme={themes.destructive}
              onClick={onRemoveClicked}
              text={<FormattedMessage id={removeButtonTextId[identityType]} />}
            />
          ) : null}
        </div>
      </ListCellLayout>
    );
  };

const UserDetailsConnectedIdentities: React.VFC<UserDetailsConnectedIdentitiesProps> =
  function UserDetailsConnectedIdentities(
    props: UserDetailsConnectedIdentitiesProps
  ) {
    const { identities, verifiedClaims, availableLoginIdIdentities } = props;
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

    const [confirmationDialogData, setConfirmationDialogData] =
      useState<ConfirmationDialogData>({
        identityID: "",
        identityName: "",
      });

    // eslint-disable-next-line complexity
    const identityLists: IdentityLists = useMemo(() => {
      const oauthIdentityList: OAuthIdentityListItem[] = [];
      const emailIdentityList: LoginIDIdentityListItem[] = [];
      const phoneIdentityList: LoginIDIdentityListItem[] = [];
      const usernameIdentityList: LoginIDIdentityListItem[] = [];
      const biometricIdentityList: BiometricIdentityListItem[] = [];
      const anonymousIdentityList: AnonymousIdentityListItem[] = [];
      const siweIdentityList: SIWEIdentityListItem[] = [];

      for (const identity of identities) {
        const createdAtStr = formatDatetime(locale, identity.createdAt) ?? "";
        if (identity.type === "OAUTH") {
          const providerType =
            identity.claims["https://authgear.com/claims/oauth/provider_type"]!;

          const claimName = "email";
          const claimValue = identity.claims.email!;

          oauthIdentityList.push({
            id: identity.id,
            type: "oauth",
            claimName,
            claimValue,
            providerType: providerType,
            verified: checkIsClaimVerified(
              verifiedClaims,
              claimName,
              claimValue
            ),
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
        if (identity.type === "SIWE") {
          const address =
            identity.claims["https://authgear.com/claims/siwe/address"];
          const chainId =
            identity.claims["https://authgear.com/claims/siwe/chain_id"];
          const formattedAddress = typeof address === "string" ? address : "";
          const formattedChainId = typeof chainId === "number" ? chainId : -1;
          siweIdentityList.push({
            id: identity.id,
            type: "siwe",
            verified: undefined,
            address: formattedAddress,
            chainId: formattedChainId,
            connectedOn: createdAtStr,
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
        siwe: siweIdentityList,
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
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
        <section className={styles.headerSection}>
          <Text as="h2" variant="medium" className={styles.header}>
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
            text={
              <FormattedMessage id="UserDetails.connected-identities.add-identity" />
            }
          />
        </section>
        <section className={styles.identityLists}>
          {identityLists.oauth.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.oauth" />
              </Text>
              <List
                items={identityLists.oauth}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
          {identityLists.email.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.email" />
              </Text>
              <List
                items={identityLists.email}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
          {identityLists.phone.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.phone" />
              </Text>
              <List
                items={identityLists.phone}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
          {identityLists.username.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.username" />
              </Text>
              <List
                items={identityLists.username}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
          {identityLists.biometric.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.biometric" />
              </Text>
              <List
                items={identityLists.biometric}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
          {identityLists.anonymous.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.anonymous" />
              </Text>
              <List
                items={identityLists.anonymous}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
          {identityLists.siwe.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.siwe" />
              </Text>
              <List
                items={identityLists.siwe}
                onRenderCell={onRenderIdentityCell}
                onShouldVirtualize={onShouldVirtualize}
              />
            </div>
          ) : null}
        </section>
      </div>
    );
  };

export default UserDetailsConnectedIdentities;
