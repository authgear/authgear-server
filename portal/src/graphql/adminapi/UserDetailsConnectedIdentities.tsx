import React, {
  useMemo,
  useCallback,
  useContext,
  useState,
  useEffect,
} from "react";
import cn from "classnames";
import { useNavigate } from "react-router-dom";
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

import PrimaryIdentitiesSelectionForm from "./PrimaryIdentitiesSelectionForm";
import ButtonWithLoading from "../../ButtonWithLoading";
import ListCellLayout from "../../ListCellLayout";
import { useDeleteIdentityMutation } from "./mutations/deleteIdentityMutation";
import { formatDatetime } from "../../util/formatDatetime";
import { parseError } from "../../util/error";
import { Violation } from "../../util/validation";
import { OAuthSSOProviderType } from "../../types";
import { destructiveTheme, verifyButtonTheme } from "../../theme";

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

interface UserDetailsConnectedIdentitiesProps {
  identities: Identity[];
  availableLoginIdIdentities: string[];
}

const loginIdIdentityTypes = ["email", "phone", "username"] as const;
type LoginIDIdentityType = typeof loginIdIdentityTypes[number];
type IdentityType = "login_id" | "oauth";

interface OAuthIdentityListItem {
  id: string;
  type: "oauth";
  providerType: OAuthSSOProviderType;
  name: string;
  verified: boolean;
  connectedOn: string;
}

interface LoginIDIdentityListItem {
  id: string;
  type: "login_id";
  key: "email" | "phone" | "username";
  value: string;
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
  identityName?: string;
  connectedOn: string;
  verified?: boolean;
  toggleVerified?: (identityID: string, verified: boolean) => void;
  onRemoveClicked: (identityID: string, identityName: string) => void;
}

interface VerifyButtonProps {
  verified: boolean;
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
  return loginIdIconMap[item.key];
}

function getName(item: LoginIDIdentityListItem | OAuthIdentityListItem) {
  if (item.type === "oauth") {
    return item.name;
  }
  return item.value;
}

function getErrorMessageIdsFromViolation(violations: Violation[]) {
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
  return { errorMessageIds, unknownViolations };
}

const VerifyButton: React.FC<VerifyButtonProps> = function VerifyButton(
  props: VerifyButtonProps
) {
  const { verified, toggleVerified } = props;

  const onClickVerify = useCallback(() => {
    toggleVerified(true);
  }, [toggleVerified]);

  const onClickUnverify = useCallback(() => {
    toggleVerified(false);
  }, [toggleVerified]);

  if (verified) {
    return (
      <DefaultButton
        className={cn(styles.controlButton, styles.unverifyButton)}
        onClick={onClickUnverify}
      >
        <FormattedMessage id={"unverify"} />
      </DefaultButton>
    );
  }

  return (
    <PrimaryButton
      className={cn(styles.controlButton, styles.verifyButton)}
      theme={verifyButtonTheme}
      onClick={onClickVerify}
    >
      <FormattedMessage id={"verify"} />
    </PrimaryButton>
  );
};

const IdentityListCell: React.FC<IdentityListCellProps> = function IdentityListCell(
  props: IdentityListCellProps
) {
  const {
    identityID,
    identityType,
    icon,
    identityName,
    connectedOn,
    verified,
    toggleVerified,
    onRemoveClicked: _onRemoveClicked,
  } = props;

  const onRemoveClicked = useCallback(() => {
    _onRemoveClicked(identityID, identityName ?? "");
  }, [identityID, identityName, _onRemoveClicked]);

  const onVerifyClicked = useCallback(
    (verified: boolean) => {
      toggleVerified?.(identityID, verified);
    },
    [toggleVerified, identityID]
  );

  return (
    <ListCellLayout className={styles.cellContainer}>
      <div className={styles.cellIcon}>{icon}</div>
      <Text className={styles.cellName}>{identityName ?? ""}</Text>
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
      {verified != null && toggleVerified != null && (
        <VerifyButton verified={verified} toggleVerified={onVerifyClicked} />
      )}
      <DefaultButton
        className={cn(styles.controlButton, styles.removeButton)}
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
  const { identities, availableLoginIdIdentities } = props;
  const { locale, renderToString } = useContext(Context);
  const navigate = useNavigate();
  const {
    deleteIdentity,
    loading: deletingIdentity,
    error: deleteIdentityError,
  } = useDeleteIdentityMutation();

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

    // TODO: get actual verified state
    for (const identity of identities) {
      const createdAtStr = formatDatetime(locale, identity.createdAt) ?? "";
      if (identity.type === "OAUTH") {
        const providerType = identity.claims[
          "https://authgear.com/claims/oauth/provider_type"
        ]!;
        oauthIdentityList.push({
          id: identity.id,
          type: "oauth",
          name: identity.claims.email!,
          providerType: providerType,
          verified: false,
          connectedOn: createdAtStr,
        });
      }

      if (identity.type === "LOGIN_ID") {
        if (
          identity.claims["https://authgear.com/claims/login_id/type"] ===
          "email"
        ) {
          emailIdentityList.push({
            id: identity.id,
            type: "login_id",
            key: "email",
            value: identity.claims.email!,
            verified: true,
            connectedOn: createdAtStr,
          });
        }

        if (
          identity.claims["https://authgear.com/claims/login_id/type"] ===
          "phone"
        ) {
          phoneIdentityList.push({
            id: identity.id,
            type: "login_id",
            key: "phone",
            value: identity.claims.phone_number!,
            verified: false,
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
            key: "username",
            value: identity.claims.preferred_username!,
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
  }, [locale, identities]);

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

  useEffect(() => {
    const fallbackErrorMessageId =
      "UserDetails.connected-identities.remove-identity-error.generic";
    const violations = parseError(deleteIdentityError);
    const {
      errorMessageIds,
      unknownViolations,
    } = getErrorMessageIdsFromViolation(violations);

    let errorMessage = null;
    if (errorMessageIds.length > 0) {
      errorMessage = errorMessageIds.map((id) => renderToString(id)).join("\n");
    } else if (unknownViolations.length > 0) {
      errorMessage = renderToString(fallbackErrorMessageId);
    }

    if (errorMessage != null) {
      setErrorDialogData({
        message: errorMessage,
      });
      setIsErrorDialogVisible(true);
    }
  }, [deleteIdentityError, renderToString]);

  const onDismissErrorDialog = useCallback(() => {
    setIsErrorDialogVisible(false);
  }, []);

  const onRenderIdentityCell = useCallback(
    (
      item?: OAuthIdentityListItem | LoginIDIdentityListItem,
      _index?: number
    ): React.ReactNode => {
      if (item == null) {
        return null;
      }

      const icon = getIcon(item);
      const name = getName(item);
      return (
        <IdentityListCell
          identityID={item.id}
          identityType={item.type}
          icon={icon}
          identityName={name}
          verified={item.verified}
          connectedOn={item.connectedOn}
          onRemoveClicked={onRemoveClicked}
          toggleVerified={() => {}}
        />
      );
    },
    [onRemoveClicked]
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

  return (
    <div className={styles.root}>
      <Dialog
        hidden={!isConfirmationDialogVisible}
        title={
          <FormattedMessage id="UserDetails.connected-identities.confirm-remove-identity-title" />
        }
        subText={renderToString(
          "UserDetails.connected-identities.confirm-remove-identity-message",
          { identityName: confirmationDialogData.identityName }
        )}
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
        title={
          <FormattedMessage id="UserDetails.connected-identities.error-dialog-title" />
        }
        subText={errorDialogData.message}
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
      <Text as="h2" className={styles.primaryIdentitiesTitle}>
        <FormattedMessage id="UserDetails.connected-identities.primary-identities.title" />
      </Text>
      <PrimaryIdentitiesSelectionForm
        className={styles.primaryIdentitiesForm}
        identityLists={identityLists}
      />
    </div>
  );
};

export default UserDetailsConnectedIdentities;
