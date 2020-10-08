import React, {
  useMemo,
  useCallback,
  useContext,
  useState,
  useEffect,
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
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import { useSetVerifiedStatusMutation } from "./mutations/setVerifiedStatusMutation";
import { formatDatetime } from "../../util/formatDatetime";
import { parseError } from "../../util/error";
import { Violation } from "../../util/validation";
import { OAuthSSOProviderType } from "../../types";
import {
  destructiveTheme,
  verifyButtonTheme,
  defaultButtonTheme,
} from "../../theme";
import { UserQuery_node_User_verifiedClaims } from "./query/__generated__/UserQuery";

import styles from "./UserDetailsConnectedIdentities.module.scss";

interface IdentityClaim extends Record<string, unknown> {
  email?: string;
  phone_number?: string;
  preferred_username?: string;
  "https://authgear.com/claims/oauth/provider_type"?: OAuthSSOProviderType;
  "https://authgear.com/claims/login_id/type"?: LoginIDIdentityType;
}

interface Identity {
  id: string;
  type: "ANONYMOUS" | "LOGIN_ID" | "OAUTH";
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
type IdentityType = "login_id" | "oauth";

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

export interface IdentityLists {
  oauth: OAuthIdentityListItem[];
  email: LoginIDIdentityListItem[];
  phone: LoginIDIdentityListItem[];
  username: LoginIDIdentityListItem[];
}

interface IdentityListCellProps {
  identityID: string;
  identityType: IdentityType;
  icon: React.ReactNode;
  claimName: string;
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

interface ErrorDialogData {
  message: string;
}

const oauthIconMap: Record<OAuthSSOProviderType, React.ReactNode> = {
  apple: <i className={cn("fab", "fa-apple", styles.widgetLabelIcon)} />,
  google: <i className={cn("fab", "fa-google", styles.widgetLabelIcon)} />,
  facebook: <i className={cn("fab", "fa-facebook", styles.widgetLabelIcon)} />,
  linkedin: <i className={cn("fab", "fa-linkedin", styles.widgetLabelIcon)} />,
  azureadv2: (
    <i className={cn("fab", "fa-microsoft", styles.widgetLabelIcon)} />
  ),
};

const loginIdIconMap: Record<LoginIDIdentityType, React.ReactNode> = {
  email: <Icon iconName="Mail" />,
  phone: <Icon iconName="CellPhone" />,
  username: <Icon iconName="Accounts" />,
};

const removeButtonTextId: Record<IdentityType, "remove" | "disconnect"> = {
  oauth: "disconnect",
  login_id: "remove",
};

function getIcon(item: LoginIDIdentityListItem | OAuthIdentityListItem) {
  if (item.type === "oauth") {
    return oauthIconMap[item.providerType];
  }
  return loginIdIconMap[item.loginIDKey];
}

function getErrorMessageFromViolation(
  violations: Violation[],
  fallbackErrorMessageId: string,
  renderToString: (messageId: string) => string
) {
  const errorMessageIds: string[] = [];
  const unknownViolations: Violation[] = [];
  for (const violation of violations) {
    switch (violation.kind) {
      case "RemoveLastIdentity":
        errorMessageIds.push(
          "UserDetails.connected-identities.remove-identity-error.connot-remove-last"
        );
        break;
      default:
        unknownViolations.push(violation);
        break;
    }
  }

  let errorMessage = null;
  if (errorMessageIds.length > 0) {
    errorMessage = errorMessageIds.map((id) => renderToString(id)).join("\n");
  } else if (unknownViolations.length > 0) {
    errorMessage = renderToString(fallbackErrorMessageId);
  }

  return errorMessage;
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
        theme={defaultButtonTheme}
        onClick={onClickUnverify}
        labelId="unverify"
        loading={verifying}
      />
    );
  }

  return (
    <ButtonWithLoading
      className={cn(styles.controlButton, styles.verifyButton)}
      disabled={settingVerifiedStatus}
      theme={verifyButtonTheme}
      onClick={onClickVerify}
      loading={verifying}
      labelId="verify"
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
    identityName,
    connectedOn,
    verified,
    setVerifiedStatus,
    onRemoveClicked: _onRemoveClicked,
  } = props;

  const { settingVerifiedStatus } = useContext(
    ConnectedIdentitiesMutationLoadingContext
  );
  const [verifying, setVerifying] = useState(false);
  const onRemoveClicked = useCallback(() => {
    _onRemoveClicked(identityID, identityName);
  }, [identityID, identityName, _onRemoveClicked]);

  const onVerifyClicked = useCallback(
    (verified: boolean) => {
      setVerifying(true);
      setVerifiedStatus?.(claimName, identityName, verified).finally(() => {
        setVerifying(false);
      });
    },
    [setVerifiedStatus, claimName, identityName]
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
        theme={destructiveTheme}
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
  const [isErrorDialogVisible, setIsErrorDialogVisible] = useState(false);

  const [confirmationDialogData, setConfirmationDialogData] = useState<
    ConfirmationDialogData
  >({
    identityID: "",
    identityName: "",
  });
  const [errorDialogData, setErrorDialogData] = useState<ErrorDialogData>({
    message: "",
  });

  const identityLists: IdentityLists = useMemo(() => {
    const oauthIdentityList: OAuthIdentityListItem[] = [];
    const emailIdentityList: LoginIDIdentityListItem[] = [];
    const phoneIdentityList: LoginIDIdentityListItem[] = [];
    const usernameIdentityList: LoginIDIdentityListItem[] = [];

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
    }
    return {
      oauth: oauthIdentityList,
      email: emailIdentityList,
      phone: phoneIdentityList,
      username: usernameIdentityList,
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
    setIsConfirmationDialogVisible(false);
  }, []);

  const onConfirmRemoveIdentity = useCallback(() => {
    const { identityID } = confirmationDialogData;
    deleteIdentity(identityID).finally(() => {
      onDismissConfirmationDialog();
    });
  }, [confirmationDialogData, deleteIdentity, onDismissConfirmationDialog]);

  const showErrorDialog = useCallback((errorMessage: string) => {
    setErrorDialogData({
      message: errorMessage,
    });
    setIsErrorDialogVisible(true);
  }, []);

  const onDismissErrorDialog = useCallback(() => {
    setIsErrorDialogVisible(false);
  }, []);

  const handleError = useCallback(
    (error: unknown, fallbackErrorMessageId: string) => {
      const violations = parseError(error);
      const errorMessage = getErrorMessageFromViolation(
        violations,
        fallbackErrorMessageId,
        renderToString
      );

      if (errorMessage != null) {
        showErrorDialog(errorMessage);
      }
    },
    [renderToString, showErrorDialog]
  );

  useEffect(() => {
    handleError(
      deleteIdentityError,
      "UserDetails.connected-identities.remove-identity-error.generic"
    );
  }, [deleteIdentityError, handleError]);

  useEffect(() => {
    handleError(
      setVerifiedStatusError,
      "UserDetails.connected-identities.verify-identity-error.generic"
    );
  }, [setVerifiedStatusError, handleError]);

  const onRenderIdentityCell = useCallback(
    (
      item?: OAuthIdentityListItem | LoginIDIdentityListItem,
      _index?: number
    ): React.ReactNode => {
      if (item == null) {
        return null;
      }

      const icon = getIcon(item);
      return (
        <IdentityListCell
          identityID={item.id}
          identityType={item.type}
          icon={icon}
          claimName={item.claimName}
          identityName={item.claimValue}
          verified={item.verified}
          connectedOn={item.connectedOn}
          onRemoveClicked={onRemoveClicked}
          setVerifiedStatus={setVerifiedStatus}
        />
      );
    },
    [onRemoveClicked, setVerifiedStatus]
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

  const errorDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="UserDetails.connected-identities.error-dialog-title" />
      ),
      subText: errorDialogData.message,
    };
  }, [errorDialogData]);

  return (
    <ConnectedIdentitiesMutationLoadingContext.Provider
      value={{ settingVerifiedStatus, deletingIdentity }}
    >
      <div className={styles.root}>
        <Dialog
          hidden={!isConfirmationDialogVisible}
          dialogContentProps={confirmationDialogContentProps}
          onDismiss={onDismissConfirmationDialog}
        >
          <DialogFooter>
            <ButtonWithLoading
              labelId="confirm"
              onClick={onConfirmRemoveIdentity}
              loading={deletingIdentity}
            />
          </DialogFooter>
        </Dialog>
        <Dialog
          hidden={!isErrorDialogVisible}
          dialogContentProps={errorDialogContentProps}
          onDismiss={onDismissErrorDialog}
        >
          <DialogFooter>
            <PrimaryButton onClick={onDismissErrorDialog}>
              <FormattedMessage id="ok" />
            </PrimaryButton>
          </DialogFooter>
        </Dialog>
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
        </section>
        {/* TODO: implement primary identities mutation
        <Text as="h2" className={styles.primaryIdentitiesTitle}>
          <FormattedMessage id="UserDetails.connected-identities.primary-identities.title" />
        </Text>
        <PrimaryIdentitiesSelectionForm
          className={styles.primaryIdentitiesForm}
          identityLists={identityLists}
        />
        */}
      </div>
    </ConnectedIdentitiesMutationLoadingContext.Provider>
  );
};

export default UserDetailsConnectedIdentities;
