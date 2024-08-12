import React, { useMemo, useCallback, useContext, useState } from "react";
import cn from "classnames";
import { generatePath, useNavigate, useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  Dialog,
  DialogFooter,
  Icon,
  IContextualMenuItem,
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
import {
  LoginIDKeyType,
  NFT,
  NFTContract,
  NFTToken,
  OAuthSSOProviderType,
  Web3Claims,
} from "../../types";
import { UserQueryNodeFragment } from "./query/userQuery.generated";

import styles from "./UserDetailsConnectedIdentities.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useIsLoading, useLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";
import {
  createEIP681URL,
  explorerAddress,
  parseEIP681,
} from "../../util/eip681";
import ExternalLink, { ExternalLinkProps } from "../../ExternalLink";
import { truncateAddress } from "../../util/hex";
import LinkButton from "../../LinkButton";
import NFTCollectionDetailDialog from "./NFTCollectionDetailDialog";

// Always disable virtualization for List component, as it wont work properly with mobile view
const onShouldVirtualize = () => {
  return false;
};

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
  preferred_username?: string;
  "https://authgear.com/claims/oauth/provider_type"?: OAuthSSOProviderType;
  "https://authgear.com/claims/oauth/subject_id"?: string;
  "https://authgear.com/claims/login_id/type"?: LoginIDIdentityType;
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
  web3Claims: Web3Claims;
}

const loginIdIdentityTypes = ["email", "phone", "username"] as const;
type LoginIDIdentityType = typeof loginIdIdentityTypes[number];
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
  | SIWEIdentityListItem
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
  nfts: NFT[] | undefined;
}

interface LDAPIdentityListItem {
  id: string;
  type: "ldap";
  verified: undefined;
  connectedOn: string;
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
  siwe: SIWEIdentityListItem[];
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
const ldapIcon: React.ReactNode = <Icon iconName="Mail" />;

const removeButtonTextId: Record<IdentityType, "remove" | "disconnect" | ""> = {
  oauth: "disconnect",
  login_id: "remove",
  biometric: "remove",
  anonymous: "",
  siwe: "",
  ldap: "",
};

// eslint-disable-next-line complexity
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
    case "siwe":
      return createEIP681URL({ chainId: item.chainId, address: item.address });
    case "ldap":
      if (item.userIDAttributeName && item.userIDAttributeValue) {
        return `${item.userIDAttributeName}=${item.userIDAttributeValue}`;
      }
      return "";
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

interface NFTCollectionListCellProps {
  contract: NFTContract;
  tokens: NFTToken[];
  eip681String: string;
}

const NFTCollectionListCell: React.VFC<NFTCollectionListCellProps> = (
  props
) => {
  const { contract, tokens, eip681String } = props;
  const [isDetailDialogVisible, setIsDetailDialogVisible] = useState(false);

  const eip681 = useMemo(() => parseEIP681(eip681String), [eip681String]);

  const contractEIP681 = useMemo(
    () =>
      createEIP681URL({ address: contract.address, chainId: eip681.chainId }),
    [contract.address, eip681.chainId]
  );

  const openDetailDialog = useCallback(() => {
    setIsDetailDialogVisible(true);
  }, []);

  const onDismissDetailDialog = useCallback(() => {
    setIsDetailDialogVisible(false);
  }, []);

  return (
    <div className={styles.NFTListCell}>
      <ExternalLink href={explorerAddress(contractEIP681)}>
        <Text
          className={cn(styles.cellName, styles.cellNameExternalLink)}
          variant="small"
        >
          <FormattedMessage
            id="UserDetails.connected-identities.siwe.nft-collections.name"
            values={{
              name: contract.name,
              address: truncateAddress(contract.address),
            }}
          />
        </Text>
      </ExternalLink>
      <LinkButton className={styles.NFTListCellBtn} onClick={openDetailDialog}>
        <Text className={styles.NFTListCellBtnLabel} variant="small">
          <FormattedMessage id="UserDetails.connected-identities.siwe.nft-collections.view-tokens" />
        </Text>
      </LinkButton>
      <NFTCollectionDetailDialog
        contract={contract}
        tokens={tokens}
        isVisible={isDetailDialogVisible}
        onDismiss={onDismissDetailDialog}
        eip681String={eip681String}
      />
    </div>
  );
};

interface NFTCollectionListProps {
  nfts?: NFT[];
  eip681String: string;
}

const NFTCollectionList: React.VFC<NFTCollectionListProps> = (props) => {
  const { nfts, eip681String } = props;

  const onRenderCollectionCell = useCallback(
    (item?: NFT, _index?: number): React.ReactNode => {
      if (item == null) {
        return null;
      }

      return (
        <NFTCollectionListCell
          contract={item.contract}
          tokens={item.tokens}
          eip681String={eip681String}
        />
      );
    },
    [eip681String]
  );

  if (nfts == null || nfts.length === 0) {
    return null;
  }

  return (
    <div>
      <Text as="h3" className={styles.NFTListHeader}>
        <FormattedMessage id="UserDetails.connected-identities.siwe.nft-collections.title" />
      </Text>
      <List
        items={nfts}
        onRenderCell={onRenderCollectionCell}
        onShouldVirtualize={onShouldVirtualize}
      />
    </div>
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
          <Text className={cn(styles.cellName, styles.cellNameExternalLink)}>
            {children}
          </Text>
        </ExternalLink>
      ) : (
        <Text className={styles.cellName}>{children}</Text>
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

  const { themes } = useSystemConfig();
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
        <DefaultButton
          className={cn(styles.controlButton, styles.removeButton)}
          disabled={loading}
          theme={themes.destructive}
          onClick={onRemoveClicked}
          text={<FormattedMessage id={removeButtonTextId[identityType]} />}
        />
      ) : null}
    </div>
  );
};

interface BaseIdentityListCellActionButtonProps {
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
  const { themes } = useSystemConfig();
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

  const menuItems = useMemo<IContextualMenuItem[]>(() => {
    const items: IContextualMenuItem[] = [];
    if (shouldShowVerifyButton) {
      items.push({
        key: "verify",
        text: renderToString(
          verified ? "make-as-unverified" : "make-as-verified"
        ),
        onClick: () => onVerifyClicked(!verified),
        disabled: verifying,
      });
    }
    if (shouldShowEditButton) {
      items.push({
        key: "edit",
        text: renderToString("edit"),
        onClick: () => onEditClicked(),
      });
    }
    if (removeButtonTextId[identityType] !== "") {
      items.push({
        key: "remove",
        text: renderToString(removeButtonTextId[identityType]),
        onClick: () => onRemoveClicked(),
      });
    }
    return items;
  }, [
    identityType,
    onEditClicked,
    onRemoveClicked,
    onVerifyClicked,
    renderToString,
    shouldShowEditButton,
    shouldShowVerifyButton,
    verified,
    verifying,
  ]);

  const menuProps = useMemo<IContextualMenuProps>(() => {
    return {
      shouldFocusOnMount: true,
      items: menuItems,
    };
  }, [menuItems]);

  return (
    <div className={styles.actionButton}>
      {menuItems.length > 0 ? (
        <DefaultButton
          disabled={loading}
          theme={themes.main}
          text={<FormattedMessage id="action" />}
          menuProps={menuProps}
        />
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
  onEditClicked: (identityID: string, loginIDKey: LoginIDKeyType) => void;
}

const LoginIDIdentityListCell: React.VFC<LoginIDIdentityListCellProps> = (
  props
) => {
  const {
    icon,
    identityID,
    identityType,
    loginIDKey,
    identityName,
    claimName,
    claimValue,
    verified,
    connectedOn,
    setVerifiedStatus,
    onRemoveClicked,
    onEditClicked: _onEditClicked,
  } = props;

  const onEditClicked = useCallback(() => {
    _onEditClicked(identityID, loginIDKey);
  }, [_onEditClicked, identityID, loginIDKey]);

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

interface SIWEIdentityListCellProps extends BaseIdentityListCellProps {
  nfts?: NFT[];
}

const SIWEIdentityListCell: React.VFC<SIWEIdentityListCellProps> = (props) => {
  const {
    icon,
    identityID,
    identityType,
    identityName,
    verified,
    connectedOn,
    nfts,
    setVerifiedStatus,
    onRemoveClicked,
  } = props;

  const externalLinkProps: ExternalLinkProps = useMemo(() => {
    return {
      href: explorerAddress(identityName),
    };
  }, [identityName]);

  return (
    <ListCellLayout className={styles.cellContainer}>
      <BaseIdentityListCellTitle
        as="ExternalLink"
        icon={icon}
        externalLinkProps={externalLinkProps}
      >
        {identityName}
      </BaseIdentityListCellTitle>
      <BaseIdentityListCellDescription verified={verified}>
        <div>
          <FormattedMessage
            id="UserDetails.connected-identities.added-on"
            values={{ datetime: connectedOn }}
          />
          <NFTCollectionList nfts={nfts} eip681String={identityName} />
        </div>
      </BaseIdentityListCellDescription>
      <BaseIdentityListCellButtonGroup
        verified={verified}
        identityID={identityID}
        identityName={identityName}
        identityType={identityType}
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
  verified?: boolean;
  connectedOn: string;
}

const LDAPIdentityListCell: React.VFC<LDAPIdentityListCellProps> = (props) => {
  const {
    icon,
    identityName,
    verified,
    connectedOn,
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
    </ListCellLayout>
  );
};

const UserDetailsConnectedIdentities: React.VFC<UserDetailsConnectedIdentitiesProps> =
  // eslint-disable-next-line complexity
  function UserDetailsConnectedIdentities(
    props: UserDetailsConnectedIdentitiesProps
  ) {
    const {
      identities,
      verifiedClaims,
      availableLoginIdIdentities,
      web3Claims,
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
      const ldapIdentityList: LDAPIdentityListItem[] = [];

      for (const identity of identities) {
        const createdAtStr = formatDatetime(locale, identity.createdAt) ?? "";
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
          const nfts = web3Claims.accounts?.find(
            (account) =>
              account.account_identifier?.address === formattedAddress
          )?.nfts;
          siweIdentityList.push({
            id: identity.id,
            type: "siwe",
            verified: undefined,
            address: formattedAddress,
            chainId: formattedChainId,
            connectedOn: createdAtStr,
            nfts: nfts,
          });
        }
        if (identity.type === "LDAP") {
          ldapIdentityList.push({
            id: identity.id,
            type: "ldap",
            verified: undefined,
            connectedOn: createdAtStr,
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
        siwe: siweIdentityList,
        ldap: ldapIdentityList,
      };
    }, [identities, locale, verifiedClaims, web3Claims.accounts]);

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
      (identityID: string, loginIDKey: LoginIDKeyType) => {
        switch (loginIDKey) {
          case "username":
            navigate(
              generatePath("./edit-username/:identityID", { identityID })
            );
            break;
          case "phone":
            navigate(generatePath("./edit-phone/:identityID", { identityID }));
            break;
          case "email":
            navigate(generatePath("./edit-email/:identityID", { identityID }));
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
          case "siwe":
            return (
              <SIWEIdentityListCell
                icon={siweIcon}
                identityID={identityID}
                identityType={identityType}
                identityName={identityName}
                verified={verified}
                connectedOn={connectedOn}
                nfts={item.nfts}
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
          {identityLists.siwe.length === 0 ? (
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
          ) : null}
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
          {identityLists.ldap.length > 0 ? (
            <div>
              <Text as="h3" className={styles.subHeader}>
                <FormattedMessage id="UserDetails.connected-identities.ldap" />
              </Text>
              <List
                items={identityLists.ldap}
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
