import React, { useMemo, useState, useCallback, useContext } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Pivot,
  PivotItem,
  MessageBar,
  MessageBarType,
  IStyle,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { produce } from "immer";

import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import ScreenContent from "../../ScreenContent";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import DeleteUserDialog from "./DeleteUserDialog";
import SetUserDisabledDialog from "./SetUserDisabledDialog";
import UserDetailSummary from "./UserDetailSummary";
import UserProfileForm, {
  CustomAttributesState,
  StandardAttributesState,
} from "./UserProfileForm";
import UserDetailsAccountSecurity from "./UserDetailsAccountSecurity";
import UserDetailsConnectedIdentities from "./UserDetailsConnectedIdentities";
import UserDetailsSession from "./UserDetailsSession";
import UserDetailsAuthorization from "./UserDetailsAuthorization";

import { useUpdateUserMutation } from "./mutations/updateUserMutation";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import { useUserQuery } from "./query/userQuery";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import { usePivotNavigation } from "../../hook/usePivot";
import { nonNullable } from "../../util/types";
import {
  PortalAPIAppConfig,
  StandardAttributes,
  CustomAttributes,
  AccessControlLevelString,
  CustomAttributesAttributeConfig,
  OAuthClientConfig,
} from "../../types";
import { jsonPointerToString, parseJSONPointer } from "../../util/jsonpointer";
import { formatDateOnly } from "../../util/formatDateOnly";
import { extractRawID } from "../../util/graphql";

import styles from "./UserDetailsScreen.module.css";
import { makeInvariantViolatedErrorParseRule } from "../../error/parse";
import { IdentityType } from "./globalTypes.generated";
import AnonymizeUserDialog from "./AnonymizeUserDialog";
import UserDetailsScreenGroupListContainer from "../../components/roles-and-groups/list/UserDetailsScreenGroupListContainer";
import UserDetailsScreenRoleListContainer from "../../components/roles-and-groups/list/UserDetailsScreenRoleListContainer";
import UserDetailsAdminActions from "./UserDetailsAdminActions";

interface UserDetailsProps {
  form: SimpleFormModel<FormState>;
  data: UserQueryNodeFragment;
  appConfig: PortalAPIAppConfig;
  onRemoveData: () => void;
  onAnonymizeData: () => void;
  handleDataStatusChange: () => void;
}

const USER_PROFILE_KEY = "user-profile";
const ACCOUNT_SECURITY_PIVOT_KEY = "account-security";
const CONNECTED_IDENTITIES_PIVOT_KEY = "connected-identities";
const SESSION_PIVOT_KEY = "session";
const ROLES_KEY = "roles";
const GROUPS_KEY = "groups";
const DISABLE_DELELE_KEY = "disable-delete";

const pivotItemContainerStyle: IStyle = {
  flex: "1 0 auto",
  display: "flex",
  flexDirection: "column",
};
interface FormState {
  userID: string;
  standardAttributes: StandardAttributesState;
  customAttributes: CustomAttributesState;
}

const ERROR_RULES = [
  makeInvariantViolatedErrorParseRule(
    "RemoveLastIdentity",
    "errors.invariant.remove-last-identity"
  ),
  makeInvariantViolatedErrorParseRule(
    "RemoveLastPrimaryAuthenticator",
    "errors.invariant.remove-last-primary-authenticator"
  ),
];

// eslint-disable-next-line complexity
function makeStandardAttributesState(
  attrs: StandardAttributes
): StandardAttributesState {
  return {
    email: attrs.email ?? "",
    phone_number: attrs.phone_number ?? "",
    preferred_username: attrs.preferred_username ?? "",
    family_name: attrs.family_name ?? "",
    given_name: attrs.given_name ?? "",
    middle_name: attrs.middle_name ?? "",
    name: attrs.name ?? "",
    nickname: attrs.nickname ?? "",
    picture: attrs.picture ?? "",
    profile: attrs.profile ?? "",
    website: attrs.website ?? "",
    gender: attrs.gender ?? "",
    birthdate: attrs.birthdate,
    zoneinfo: attrs.zoneinfo ?? "",
    locale: attrs.locale ?? "",
    address: {
      street_address: attrs.address?.street_address ?? "",
      locality: attrs.address?.locality ?? "",
      region: attrs.address?.region ?? "",
      postal_code: attrs.address?.postal_code ?? "",
      country: attrs.address?.country ?? "",
    },
    updated_at: attrs.updated_at,
  };
}

function makeCustomAttributesState(
  attrs: CustomAttributes,
  config: CustomAttributesAttributeConfig[]
): CustomAttributesState {
  const state: CustomAttributesState = {};
  for (const c of config) {
    const ptr = parseJSONPointer(c.pointer);
    // FIXME(custom-attributes): support any-level jsonpointer.
    const unknownValue = attrs[ptr[0]];
    if (unknownValue == null) {
      state[c.pointer] = "";
    } else {
      state[c.pointer] = String(unknownValue);
    }

    if (c.type === "phone_number") {
      if (unknownValue == null) {
        state["phone_number" + c.pointer] = "";
      } else {
        state["phone_number" + c.pointer] = String(unknownValue);
      }
    }
  }
  return state;
}

function makeStandardAttributesFromState(
  state: StandardAttributesState
): StandardAttributes {
  return produce(state, (state) => {
    delete state.updated_at;

    for (const key of Object.keys(state)) {
      // @ts-expect-error
      const value = state[key];
      if (value === "") {
        // @ts-expect-error
        delete state[key];
      }
    }

    for (const key of Object.keys(state.address)) {
      // @ts-expect-error
      const value = state.address[key];
      if (value === "") {
        // @ts-expect-error
        delete state.address[key];
      }
    }
    if (Object.keys(state.address).length === 0) {
      // @ts-expect-error
      delete state.address;
    }
  });
}

// eslint-disable-next-line complexity
function makeCustomAttributesFromState(
  state: CustomAttributesState,
  config: CustomAttributesAttributeConfig[]
): CustomAttributes {
  const out: CustomAttributes = {};
  for (const c of config) {
    const value = state[c.pointer];

    if (value === "") {
      continue;
    }

    // FIXME(custom-attributes): support any-level jsonpointer.
    const ptr = parseJSONPointer(c.pointer);
    const fieldName = ptr[0];

    switch (c.type) {
      case "string":
        out[fieldName] = value;
        break;
      case "number":
        out[fieldName] = parseFloat(value);
        break;
      case "integer":
        out[fieldName] = parseInt(value, 10);
        break;
      case "enum":
        out[fieldName] = value;
        break;
      case "phone_number":
        out[fieldName] = value;
        break;
      case "email":
        out[fieldName] = value;
        break;
      case "url":
        out[fieldName] = value;
        break;
      case "country_code":
        out[fieldName] = value;
        break;
    }
  }

  return out;
}

// eslint-disable-next-line complexity
const UserDetails: React.VFC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { selectedKey, onLinkClick } = usePivotNavigation([
    USER_PROFILE_KEY,
    ACCOUNT_SECURITY_PIVOT_KEY,
    CONNECTED_IDENTITIES_PIVOT_KEY,
    SESSION_PIVOT_KEY,
    ROLES_KEY,
    GROUPS_KEY,
    DISABLE_DELELE_KEY,
  ]);
  const {
    form,
    data,
    appConfig,
    onRemoveData,
    onAnonymizeData,
    handleDataStatusChange,
  } = props;
  const { state, setState } = form;
  const { renderToString } = React.useContext(Context);

  const availableLoginIdIdentities = useMemo(() => {
    const authenticationIdentities = appConfig.authentication?.identities ?? [];
    const loginIdIdentityEnabled =
      authenticationIdentities.includes("login_id");
    if (!loginIdIdentityEnabled) {
      return [];
    }
    const rawLoginIdKeys = appConfig.identity?.login_id?.keys ?? [];
    return rawLoginIdKeys.map((loginIdKey) => loginIdKey.type);
  }, [appConfig]);

  const standardAttributeAccessControl = useMemo(() => {
    const record: Record<string, AccessControlLevelString> = {};
    for (const item of appConfig.user_profile?.standard_attributes
      ?.access_control ?? []) {
      record[item.pointer] = item.access_control.portal_ui;
    }
    return record;
  }, [appConfig]);

  const customAttributesConfig: CustomAttributesAttributeConfig[] =
    useMemo(() => {
      return appConfig.user_profile?.custom_attributes?.attributes ?? [];
    }, [appConfig]);

  const oauthClientConfig: OAuthClientConfig[] = useMemo(() => {
    return appConfig.oauth?.clients ?? [];
  }, [appConfig]);

  const onChangeStandardAttributes = useCallback(
    (attrs: StandardAttributesState) => {
      setState((state) => {
        return {
          ...state,
          standardAttributes: attrs,
        };
      });
    },
    [setState]
  );

  const onChangeCustomAttributes = useCallback(
    (attrs: CustomAttributesState) => {
      setState((state) => {
        return {
          ...state,
          customAttributes: attrs,
        };
      });
    },
    [setState]
  );

  const web3Claims = data.web3;

  const verifiedClaims = data.verifiedClaims;

  const identities = useMemo(
    () =>
      data.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
      [],
    [data.identities]
  );

  const authenticators = useMemo(
    () =>
      data.authenticators?.edges
        ?.map((edge) => edge?.node)
        .filter(nonNullable) ?? [],
    [data.authenticators]
  );

  const sessions =
    data.sessions?.edges?.map((edge) => edge?.node).filter(nonNullable) ?? [];

  const authorizations =
    data.authorizations?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
    [];

  const profileImageEditable = useMemo(() => {
    const ptr = jsonPointerToString(["picture"]);
    const level = standardAttributeAccessControl[ptr];
    return level === "readwrite";
  }, [standardAttributeAccessControl]);

  const isSIWEIdentity = useMemo(
    () => identities.some((i) => i.type === IdentityType.Siwe),
    [identities]
  );

  const dataStatusBadgeTextId = React.useMemo(() => {
    let badgeTextId = null;
    if (data.isDisabled) {
      badgeTextId = "UserDetails.disabled.badge";
    }
    if (data.anonymizeAt) {
      badgeTextId = "UserDetails.scheduled-anonymization.badge";
    }
    if (data.isAnonymized) {
      badgeTextId = "UserDetails.anonymized.badge";
    }
    if (data.deleteAt) {
      badgeTextId = "UserDetails.scheduled-removal.badge";
    }
    return badgeTextId;
  }, [data.isDisabled, data.anonymizeAt, data.isAnonymized, data.deleteAt]);

  if (data.isAnonymized) {
    return (
      <div className={styles.widget}>
        <UserDetailSummary
          isAnonymous={data.isAnonymous}
          isAnonymized={data.isAnonymized}
          profileImageURL={data.standardAttributes.picture}
          profileImageEditable={profileImageEditable}
          rawUserID={extractRawID(data.id)}
          formattedName={data.formattedName ?? undefined}
          endUserAccountIdentifier={data.endUserAccountID ?? undefined}
          createdAtISO={data.createdAt ?? null}
          lastLoginAtISO={data.lastLoginAt ?? null}
          badgeTextId={dataStatusBadgeTextId}
        />
        <MessageBar messageBarType={MessageBarType.info}>
          <FormattedMessage id="UserDetailsScreen.user-anonymized.message" />
        </MessageBar>
        <Pivot
          styles={{ itemContainer: pivotItemContainerStyle }}
          className={styles.pivot}
          overflowBehavior="menu"
          selectedKey={selectedKey}
          onLinkClick={onLinkClick}
        >
          <PivotItem
            className={"flex-1 pt-8"}
            itemKey={DISABLE_DELELE_KEY}
            headerText={renderToString("UserDetails.disable-delete.header")}
          >
            <UserDetailsAdminActions
              data={data}
              onAnonymizeData={onAnonymizeData}
              handleDataStatusChange={handleDataStatusChange}
              onRemoveData={onRemoveData}
            />
          </PivotItem>
        </Pivot>
      </div>
    );
  }

  return (
    <div className={styles.widget}>
      <UserDetailSummary
        isAnonymous={data.isAnonymous}
        isAnonymized={data.isAnonymized}
        profileImageURL={data.standardAttributes.picture}
        profileImageEditable={profileImageEditable}
        rawUserID={extractRawID(data.id)}
        formattedName={data.formattedName ?? undefined}
        endUserAccountIdentifier={data.endUserAccountID ?? undefined}
        createdAtISO={data.createdAt ?? null}
        lastLoginAtISO={data.lastLoginAt ?? null}
        badgeTextId={dataStatusBadgeTextId}
      />
      <Pivot
        styles={{
          itemContainer: {
            flex: "1 0 auto",
            display: "flex",
            flexDirection: "column",
          },
        }}
        className={styles.pivot}
        overflowBehavior="menu"
        selectedKey={selectedKey}
        onLinkClick={onLinkClick}
      >
        <PivotItem
          itemKey={USER_PROFILE_KEY}
          headerText={renderToString("UserDetails.user-profile.header")}
        >
          <UserProfileForm
            identities={identities}
            standardAttributes={state.standardAttributes}
            onChangeStandardAttributes={onChangeStandardAttributes}
            standardAttributeAccessControl={standardAttributeAccessControl}
            customAttributesConfig={customAttributesConfig}
            customAttributes={state.customAttributes}
            onChangeCustomAttributes={onChangeCustomAttributes}
          />
        </PivotItem>
        <PivotItem
          itemKey={ACCOUNT_SECURITY_PIVOT_KEY}
          headerText={renderToString("UserDetails.account-security.header")}
        >
          {isSIWEIdentity ? (
            <MessageBar className={styles.siweEnabledTabWarningMessageBar}>
              <FormattedMessage id="UserDetailsScreen.user-account-security.siwe-enabled" />
            </MessageBar>
          ) : (
            <UserDetailsAccountSecurity
              userID={data.id}
              authenticationConfig={appConfig.authentication}
              authenticatorConfig={appConfig.authenticator}
              identities={identities}
              authenticators={authenticators}
            />
          )}
        </PivotItem>
        <PivotItem
          itemKey={CONNECTED_IDENTITIES_PIVOT_KEY}
          headerText={renderToString("UserDetails.connected-identities.header")}
        >
          <UserDetailsConnectedIdentities
            identities={identities}
            verifiedClaims={verifiedClaims}
            availableLoginIdIdentities={availableLoginIdIdentities}
            web3Claims={web3Claims}
          />
        </PivotItem>
        <PivotItem
          itemKey={SESSION_PIVOT_KEY}
          headerText={renderToString("UserDetails.session.header")}
        >
          <UserDetailsSession
            sessions={sessions}
            oauthClients={oauthClientConfig}
          />
          <UserDetailsAuthorization
            authorizations={authorizations}
            oauthClientConfig={oauthClientConfig}
          />
        </PivotItem>
        <PivotItem
          className={"flex-1 pt-8"}
          itemKey={ROLES_KEY}
          headerText={renderToString("UserDetails.roles.header")}
        >
          <UserDetailsScreenRoleListContainer user={data} />
        </PivotItem>
        <PivotItem
          className={"flex-1 pt-8"}
          itemKey={GROUPS_KEY}
          headerText={renderToString("UserDetails.groups.header")}
        >
          <UserDetailsScreenGroupListContainer user={data} />
        </PivotItem>
        <PivotItem
          className={"flex-1 pt-8"}
          itemKey={DISABLE_DELELE_KEY}
          headerText={renderToString("UserDetails.disable-delete.header")}
        >
          <UserDetailsAdminActions
            data={data}
            onAnonymizeData={onAnonymizeData}
            handleDataStatusChange={handleDataStatusChange}
            onRemoveData={onRemoveData}
          />
        </PivotItem>
      </Pivot>
    </div>
  );
};

interface WarnScheduledDeletionProps {
  user: UserQueryNodeFragment;
}

function WarnScheduledDeletion(props: WarnScheduledDeletionProps) {
  const { user } = props;
  const { locale } = useContext(Context);
  if (user.deleteAt == null) {
    return null;
  }

  return (
    <MessageBar messageBarType={MessageBarType.warning}>
      <FormattedMessage
        id="UserDetailsScreen.scheduled-deletion"
        values={{
          date: formatDateOnly(locale, user.deleteAt) ?? "",
        }}
      />
    </MessageBar>
  );
}

interface WarnScheduledAnonymizationProps {
  user: UserQueryNodeFragment;
}

function WarnScheduledAnonymization(props: WarnScheduledAnonymizationProps) {
  const { user } = props;
  const { locale } = useContext(Context);
  if (user.anonymizeAt == null) {
    return null;
  }

  return (
    <MessageBar messageBarType={MessageBarType.warning}>
      <FormattedMessage
        id="UserDetailsScreen.scheduled-anonymization"
        values={{
          date: formatDateOnly(locale, user.anonymizeAt) ?? "",
        }}
      />
    </MessageBar>
  );
}

interface UserDetailsScreenContentProps {
  user: UserQueryNodeFragment;
  refreshUser?: () => void;
  effectiveAppConfig: PortalAPIAppConfig;
}

const UserDetailsScreenContent: React.VFC<UserDetailsScreenContentProps> =
  // eslint-disable-next-line complexity
  function UserDetailsScreenContent(props: UserDetailsScreenContentProps) {
    const { user, refreshUser, effectiveAppConfig } = props;
    const navigate = useNavigate();
    const customAttributesConfig = useMemo(() => {
      return (
        effectiveAppConfig.user_profile?.custom_attributes?.attributes ?? []
      );
    }, [effectiveAppConfig.user_profile?.custom_attributes?.attributes]);

    const navBreadcrumbItems = React.useMemo(() => {
      return [
        { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
        { to: ".", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      ];
    }, []);

    const [deleteUserDialogIsHidden, setDeleteUserDialogIsHidden] =
      useState(true);
    const onDismissDeleteUserDialog = useCallback(
      (deletedUser: boolean) => {
        setDeleteUserDialogIsHidden(true);
        if (deletedUser) {
          setTimeout(() => navigate("./../.."), 0);
        }
      },
      [navigate]
    );
    const onClickDeleteUser = useCallback(() => {
      setDeleteUserDialogIsHidden(false);
    }, []);

    const [anonymizeUserDialogIsHidden, setAnonymizeUserDialogIsHidden] =
      useState(true);
    const onDismissAnonymizeUserDialog = useCallback(() => {
      setAnonymizeUserDialogIsHidden(true);
    }, []);
    const onClickAnonymizeUser = useCallback(() => {
      setAnonymizeUserDialogIsHidden(false);
    }, []);

    const [setUserDisabledDialogIsHidden, setSetUserDisabledDialogIsHidden] =
      useState(true);
    const [userIsDisabled, setUserIsDisabled] = useState(user.isDisabled);
    const onDismissSetUserDisabledDialog = useCallback(() => {
      setSetUserDisabledDialogIsHidden(true);
    }, []);
    const onClickSetUserDisabled = useCallback(() => {
      setSetUserDisabledDialogIsHidden(false);
      setUserIsDisabled(user.isDisabled);
    }, [user.isDisabled]);

    const defaultState = useMemo(() => {
      return {
        userID: user.id,
        standardAttributes: makeStandardAttributesState(
          user.standardAttributes
        ),
        customAttributes: makeCustomAttributesState(
          user.customAttributes,
          customAttributesConfig
        ),
      };
    }, [
      user.id,
      user.standardAttributes,
      user.customAttributes,
      customAttributesConfig,
    ]);

    const { updateUser } = useUpdateUserMutation();

    const submit = useCallback(
      async (state: FormState) => {
        await updateUser(
          state.userID,
          makeStandardAttributesFromState(state.standardAttributes),
          makeCustomAttributesFromState(
            state.customAttributes,
            customAttributesConfig
          )
        );
        refreshUser?.();
      },
      [updateUser, customAttributesConfig, refreshUser]
    );

    const form = useSimpleForm({
      stateMode: "UpdateInitialStateWithUseEffect",
      defaultState,
      submit,
    });

    return (
      <FormContainer
        className={styles.formContainer}
        errorRules={ERROR_RULES}
        form={form}
        hideFooterComponent={true}
        messageBar={
          <>
            <WarnScheduledDeletion user={user} />
            <WarnScheduledAnonymization user={user} />
          </>
        }
      >
        <ScreenContent className={styles.screenContent}>
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <UserDetails
            form={form}
            data={user}
            appConfig={effectiveAppConfig}
            onRemoveData={onClickDeleteUser}
            onAnonymizeData={onClickAnonymizeUser}
            handleDataStatusChange={onClickSetUserDisabled}
          />
        </ScreenContent>
        <DeleteUserDialog
          isHidden={deleteUserDialogIsHidden}
          onDismiss={onDismissDeleteUserDialog}
          userID={user.id}
          userDeleteAt={user.deleteAt}
          endUserAccountIdentifier={user.endUserAccountID ?? undefined}
        />
        <AnonymizeUserDialog
          isHidden={anonymizeUserDialogIsHidden}
          onDismiss={onDismissAnonymizeUserDialog}
          userID={user.id}
          userAnonymizeAt={user.anonymizeAt}
          endUserAccountIdentifier={user.endUserAccountID ?? undefined}
        />
        <SetUserDisabledDialog
          isHidden={setUserDisabledDialogIsHidden}
          onDismiss={onDismissSetUserDisabledDialog}
          userID={user.id}
          userIsDisabled={userIsDisabled}
          userDeleteAt={user.deleteAt}
          userAnonymizeAt={user.anonymizeAt}
          endUserAccountIdentifier={user.endUserAccountID ?? undefined}
        />
      </FormContainer>
    );
  };

const UserDetailsScreen: React.VFC = function UserDetailsScreen() {
  const { appID, userID } = useParams() as { appID: string; userID: string };
  const { user, loading: loadingUser, error, refetch } = useUserQuery(userID);
  const {
    effectiveAppConfig,
    loading: loadingAppConfig,
    error: appConfigError,
    refetch: refetchAppConfig,
  } = useAppAndSecretConfigQuery(appID);
  const loading = loadingUser || loadingAppConfig;

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (appConfigError != null) {
    return <ShowError error={appConfigError} onRetry={refetchAppConfig} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (user == null || effectiveAppConfig == null) {
    return <ShowLoading />;
  }

  return (
    <UserDetailsScreenContent
      user={user}
      refreshUser={refetch}
      effectiveAppConfig={effectiveAppConfig}
    />
  );
};

export default UserDetailsScreen;
