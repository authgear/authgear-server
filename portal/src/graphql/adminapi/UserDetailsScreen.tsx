import React, { useMemo, useCallback, useRef } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "../../intl";
import { produce } from "immer";
import cn from "classnames";

import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { Callout } from "../../components/v2/Callout/Callout";
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
import { SimpleFormModel } from "../../hook/useSimpleForm";
import { useFormWithExternalInitialState } from "../../hook/useFormWithExternalInitialState";
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
  Authorization,
  Session,
} from "../../types";
import { jsonPointerToString, parseJSONPointer } from "../../util/jsonpointer";
import { extractRawID } from "../../util/graphql";

import styles from "./UserDetailsScreen.module.css";
import { makeInvariantViolatedErrorParseRule } from "../../error/parse";
import UserDetailsScreenGroupListContainer from "../../components/roles-and-groups/list/UserDetailsScreenGroupListContainer";
import UserDetailsScreenRoleListContainer from "../../components/roles-and-groups/list/UserDetailsScreenRoleListContainer";
import UserDetailsAccountStatus, {
  AccountStatusMessageBar,
} from "./UserDetailsAccountStatus";
import UserDetailsLogs from "./UserDetailsLogs";
import { useSystemConfig } from "../../context/SystemConfigContext";

// Temporary UI preview data for Sam Lee — remove after Sessions & Apps UI work.
const UI_PREVIEW_RAW_USER_ID = "8d442fbc-ea7c-4130-9114-432631fdb5d3";

function buildUIPreviewSessions(oauthClients: OAuthClientConfig[]): Session[] {
  const clientID = oauthClients[0]?.client_id ?? "portal";
  const secondClientID = oauthClients[1]?.client_id ?? clientID;
  return [
    {
      id: "fake-session-1",
      type: "IDP",
      displayName: "Chrome on macOS",
      userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/131.0",
      clientID,
      lastAccessedByIP: "203.80.12.45",
      lastAccessedAt: "2025-02-07T07:55:00.000Z",
    },
    {
      id: "fake-session-2",
      type: "OFFLINE_GRANT",
      displayName: "Safari on iPhone",
      userAgent:
        "Mozilla/5.0 (iPhone; CPU iPhone OS 18_2 like Mac OS X) Safari/605.1.15",
      clientID: secondClientID,
      lastAccessedByIP: "49.130.8.221",
      lastAccessedAt: "2025-02-06T14:12:00.000Z",
    },
    {
      id: "fake-session-3",
      type: "OFFLINE_GRANT",
      displayName: "",
      userAgent: null,
      clientID: null,
      lastAccessedByIP: "10.0.0.18",
      lastAccessedAt: "2025-02-01T02:30:00.000Z",
    },
  ];
}

function buildUIPreviewAuthorizations(
  oauthClients: OAuthClientConfig[]
): Authorization[] {
  const clientID = oauthClients[0]?.client_id ?? "portal";
  const secondClientID = oauthClients[1]?.client_id ?? clientID;
  return [
    {
      id: "fake-authorization-1",
      clientID,
      createdAt: "2024-11-12T03:20:00.000Z",
      scopes: [
        "openid",
        "offline_access",
        "https://authgear.com/scopes/full-userinfo",
      ],
    },
    {
      id: "fake-authorization-2",
      clientID: secondClientID,
      createdAt: "2025-01-18T09:05:00.000Z",
      scopes: ["openid"],
    },
  ];
}
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "../../ErrorMessageBar";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { ProfilePictureDialog } from "./ProfilePictureDialog";

interface UserDetailsProps {
  form: SimpleFormModel<FormState>;
  data: UserQueryNodeFragment;
  appConfig: PortalAPIAppConfig;
  refreshUser?: () => unknown;
  profileContentRef?: React.RefObject<HTMLDivElement>;
}

const USER_PROFILE_KEY = "user-profile";
const ACCOUNT_SECURITY_PIVOT_KEY = "account-security";
const CONNECTED_IDENTITIES_PIVOT_KEY = "connected-identities";
const SESSION_PIVOT_KEY = "session";
const ROLES_KEY = "roles";
const GROUPS_KEY = "groups";
const ACCOUNT_STATUS_KEY = "account-status";
const LOGS_KEY = "logs";

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
      // eslint-disable-next-line @typescript-eslint/no-base-to-string
      state[c.pointer] = String(unknownValue);
    }

    if (c.type === "phone_number") {
      if (unknownValue == null) {
        state["phone_number" + c.pointer] = "";
      } else {
        // eslint-disable-next-line @typescript-eslint/no-base-to-string
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
      const value = (state as any)[key];
      if (value === "") {
        delete (state as any)[key];
      }
    }

    for (const key of Object.keys(state.address)) {
      const value = (state.address as any)[key];
      if (value === "") {
        delete (state.address as any)[key];
      }
    }
    if (Object.keys(state.address).length === 0) {
      delete (state as any).address;
    }
  });
}

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

const UserDetails: React.VFC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { auditLogEnabled } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const { form, data, appConfig, refreshUser, profileContentRef } = props;
  const pivotItemKeys = useMemo(
    () =>
      data.isAnonymized
        ? [ACCOUNT_STATUS_KEY, ...(auditLogEnabled ? [LOGS_KEY] : [])]
        : [
            USER_PROFILE_KEY,
            ACCOUNT_SECURITY_PIVOT_KEY,
            CONNECTED_IDENTITIES_PIVOT_KEY,
            SESSION_PIVOT_KEY,
            ROLES_KEY,
            GROUPS_KEY,
            ACCOUNT_STATUS_KEY,
            ...(auditLogEnabled ? [LOGS_KEY] : []),
          ],
    [data.isAnonymized, auditLogEnabled]
  );
  const { selectedKey, onChangeKey } = usePivotNavigation(pivotItemKeys);
  const { state, setState } = form;
  const { renderToString } = React.useContext(Context);
  const [selectedProfileImage, setSelectedProfileImage] =
    React.useState<File | null>(null);

  const onDismissProfilePictureDialog = useCallback(() => {
    setSelectedProfileImage(null);
  }, []);

  const profilePictureDialog = (
    <ProfilePictureDialog
      appID={appID}
      user={data}
      file={selectedProfileImage}
      onDismiss={onDismissProfilePictureDialog}
      onSaved={refreshUser}
    />
  );

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

  const sessions = useMemo(() => {
    const realSessions =
      data.sessions?.edges?.map((edge) => edge?.node).filter(nonNullable) ?? [];
    if (
      realSessions.length === 0 &&
      extractRawID(data.id) === UI_PREVIEW_RAW_USER_ID
    ) {
      return buildUIPreviewSessions(oauthClientConfig);
    }
    return realSessions;
  }, [data.id, data.sessions, oauthClientConfig]);

  const authorizations = useMemo(() => {
    const realAuthorizations =
      data.authorizations?.edges
        ?.map((edge) => edge?.node)
        .filter(nonNullable) ?? [];
    if (
      realAuthorizations.length === 0 &&
      extractRawID(data.id) === UI_PREVIEW_RAW_USER_ID
    ) {
      return buildUIPreviewAuthorizations(oauthClientConfig);
    }
    return realAuthorizations;
  }, [data.authorizations, data.id, oauthClientConfig]);

  const profileImageEditable = useMemo(() => {
    const ptr = jsonPointerToString(["picture"]);
    const level = standardAttributeAccessControl[ptr];
    return level === "readwrite";
  }, [standardAttributeAccessControl]);

  const tabs = useMemo(
    () => [
      {
        value: USER_PROFILE_KEY,
        label: renderToString("UserDetails.user-profile.header"),
      },
      {
        value: CONNECTED_IDENTITIES_PIVOT_KEY,
        label: renderToString("UserDetails.connected-identities.header"),
      },
      {
        value: ACCOUNT_SECURITY_PIVOT_KEY,
        label: renderToString("UserDetails.account-security.header"),
      },
      {
        value: SESSION_PIVOT_KEY,
        label: renderToString("UserDetails.session.header"),
      },
      {
        value: ROLES_KEY,
        label: renderToString("UserDetails.roles.header"),
      },
      {
        value: GROUPS_KEY,
        label: renderToString("UserDetails.groups.header"),
      },
      {
        value: ACCOUNT_STATUS_KEY,
        label: renderToString("UserDetails.account-status.header"),
      },
      ...(auditLogEnabled
        ? [
            {
              value: LOGS_KEY,
              label: renderToString("UserDetails.logs.header"),
            },
          ]
        : []),
    ],
    [auditLogEnabled, renderToString]
  );

  const tabsForAnonymized = useMemo(
    () => [
      {
        value: ACCOUNT_STATUS_KEY,
        label: renderToString("UserDetails.account-status.header"),
      },
      ...(auditLogEnabled
        ? [
            {
              value: LOGS_KEY,
              label: renderToString("UserDetails.logs.header"),
            },
          ]
        : []),
    ],
    [auditLogEnabled, renderToString]
  );

  if (data.isAnonymized) {
    return (
      <>
        <div className={styles.widget}>
          <UserDetailSummary
            isAnonymous={data.isAnonymous}
            isAnonymized={data.isAnonymized}
            profileImageURL={data.standardAttributes.picture}
            profileImageEditable={profileImageEditable}
            onSelectProfileImage={setSelectedProfileImage}
            rawUserID={extractRawID(data.id)}
            formattedName={data.formattedName ?? undefined}
            endUserAccountIdentifier={data.endUserAccountID ?? undefined}
            createdAtISO={data.createdAt ?? null}
            lastLoginAtISO={data.lastLoginAt ?? null}
            accountStatus={data}
          />
          <Callout
            type="warning"
            color="yellow"
            showCloseButton={false}
            text={
              <FormattedMessage id="UserDetailsScreen.user-anonymized.message" />
            }
          />
          <OverflowTabs
            className={styles.tabs}
            value={selectedKey}
            onValueChange={(v) => onChangeKey(v)}
            tabs={tabsForAnonymized}
          />
          {selectedKey === ACCOUNT_STATUS_KEY ? (
            <div className={styles.tabContent}>
              <UserDetailsAccountStatus data={data} />
            </div>
          ) : null}
          {selectedKey === LOGS_KEY ? (
            <div className={styles.tabContent}>
              <UserDetailsLogs userID={data.id} />
            </div>
          ) : null}
        </div>
        {profilePictureDialog}
      </>
    );
  }

  return (
    <>
      <div className={styles.widget}>
        <UserDetailSummary
          isAnonymous={data.isAnonymous}
          isAnonymized={data.isAnonymized}
          profileImageURL={data.standardAttributes.picture}
          profileImageEditable={profileImageEditable}
          onSelectProfileImage={setSelectedProfileImage}
          rawUserID={extractRawID(data.id)}
          formattedName={data.formattedName ?? undefined}
          endUserAccountIdentifier={data.endUserAccountID ?? undefined}
          createdAtISO={data.createdAt ?? null}
          lastLoginAtISO={data.lastLoginAt ?? null}
          accountStatus={data}
        />
        <AccountStatusMessageBar accountStatus={data} />
        <OverflowTabs
          className={styles.tabs}
          value={selectedKey}
          onValueChange={(v) => onChangeKey(v)}
          tabs={tabs}
        />

        {selectedKey === USER_PROFILE_KEY ? (
          <div className={styles.profileTabContent}>
            <div ref={profileContentRef} className={styles.profileTabMain}>
              <UserProfileForm
                identities={identities}
                standardAttributes={state.standardAttributes}
                onChangeStandardAttributes={onChangeStandardAttributes}
                standardAttributeAccessControl={standardAttributeAccessControl}
                customAttributesConfig={customAttributesConfig}
                customAttributes={state.customAttributes}
                onChangeCustomAttributes={onChangeCustomAttributes}
                profileImageEditable={profileImageEditable}
                onSelectProfileImage={setSelectedProfileImage}
              />
            </div>
          </div>
        ) : null}

        {selectedKey === ACCOUNT_SECURITY_PIVOT_KEY ? (
          <div className={styles.tabContent}>
            <UserDetailsAccountSecurity
              userID={data.id}
              authenticationConfig={appConfig.authentication}
              authenticatorConfig={appConfig.authenticator}
              identities={identities}
              authenticators={authenticators}
              phoneInputAllowlist={appConfig.ui?.phone_input?.allowlist}
              phoneInputPinnedList={appConfig.ui?.phone_input?.pinned_list}
              onAuthenticatorCreated={refreshUser}
            />
          </div>
        ) : null}

        {selectedKey === CONNECTED_IDENTITIES_PIVOT_KEY ? (
          <div className={styles.tabContent}>
            <UserDetailsConnectedIdentities
              identities={identities}
              verifiedClaims={verifiedClaims}
              availableLoginIdIdentities={availableLoginIdIdentities}
              phoneInputAllowlist={appConfig.ui?.phone_input?.allowlist}
              phoneInputPinnedList={appConfig.ui?.phone_input?.pinned_list}
              onIdentityCreated={refreshUser}
            />
          </div>
        ) : null}

        {selectedKey === SESSION_PIVOT_KEY ? (
          <div className={cn(styles.tabContent, styles.fullWidthTabContent)}>
            <UserDetailsSession
              sessions={sessions}
              oauthClients={oauthClientConfig}
            />
            <UserDetailsAuthorization
              authorizations={authorizations}
              oauthClientConfig={oauthClientConfig}
            />
          </div>
        ) : null}

        {selectedKey === ROLES_KEY ? (
          <div className={styles.tabContent}>
            <UserDetailsScreenRoleListContainer user={data} />
          </div>
        ) : null}

        {selectedKey === GROUPS_KEY ? (
          <div className={styles.tabContent}>
            <UserDetailsScreenGroupListContainer user={data} />
          </div>
        ) : null}

        {selectedKey === ACCOUNT_STATUS_KEY ? (
          <div className={styles.tabContent}>
            <UserDetailsAccountStatus data={data} />
          </div>
        ) : null}

        {selectedKey === LOGS_KEY ? (
          <div className={styles.tabContent}>
            <UserDetailsLogs userID={data.id} />
          </div>
        ) : null}
      </div>
      {profilePictureDialog}
    </>
  );
};

interface UserDetailsScreenContentProps {
  user: UserQueryNodeFragment;
  refreshUser?: () => unknown;
  effectiveAppConfig: PortalAPIAppConfig;
}

interface UserDetailsScreenFormProps {
  form: SimpleFormModel<FormState>;
  user: UserQueryNodeFragment;
  refreshUser?: () => unknown;
  effectiveAppConfig: PortalAPIAppConfig;
}

const UserDetailsScreenForm: React.VFC<UserDetailsScreenFormProps> =
  function UserDetailsScreenForm(props: UserDetailsScreenFormProps) {
    const { form, user, refreshUser, effectiveAppConfig } = props;
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    return (
      <>
        <div
          className={cn(
            styles.screenContent,
            isDirty ? styles.contentWithSaveBar : null
          )}
        >
          <UserDetails
            form={form}
            data={user}
            appConfig={effectiveAppConfig}
            refreshUser={refreshUser}
            profileContentRef={contentWidthAnchorRef}
          />
        </div>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </>
    );
  };

const UserDetailsScreenContent: React.VFC<UserDetailsScreenContentProps> =
  function UserDetailsScreenContent(props: UserDetailsScreenContentProps) {
    const { user, refreshUser, effectiveAppConfig } = props;
    const customAttributesConfig = useMemo(() => {
      return (
        effectiveAppConfig.user_profile?.custom_attributes?.attributes ?? []
      );
    }, [effectiveAppConfig.user_profile?.custom_attributes?.attributes]);

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
        return { result: undefined };
      },
      [updateUser, customAttributesConfig, refreshUser]
    );

    const form = useFormWithExternalInitialState({
      defaultState,
      submit,
    });

    return (
      <ErrorMessageBarContextProvider>
        <div className={styles.screenRoot}>
          <div className={styles.topBar}>
            <ErrorMessageBar />
          </div>
          <FormContainer
            className={styles.formContainer}
            errorRules={ERROR_RULES}
            form={form}
            hideFooterComponent={true}
          >
            <UserDetailsScreenForm
              form={form}
              user={user}
              refreshUser={refreshUser}
              effectiveAppConfig={effectiveAppConfig}
            />
          </FormContainer>
        </div>
      </ErrorMessageBarContextProvider>
    );
  };

const UserDetailsScreen: React.VFC = function UserDetailsScreen() {
  const { appID, userID } = useParams() as { appID: string; userID: string };
  const {
    user,
    loading: loadingUser,
    error,
    refetch,
  } = useUserQuery(userID, {
    fetchPolicy: "cache-and-network",
  });
  const {
    effectiveAppConfig,
    isLoading: loadingAppConfig,
    loadError: appConfigError,
    refetch: refetchAppConfig,
  } = useAppAndSecretConfigQuery(appID);
  const loading = loadingUser || loadingAppConfig;

  if (error != null) {
    return (
      <ShowError
        error={error}
        onRetry={() => {
          refetch().finally(() => {});
        }}
      />
    );
  }

  if (appConfigError != null) {
    return (
      <ShowError
        error={appConfigError}
        onRetry={() => {
          refetchAppConfig().finally(() => {});
        }}
      />
    );
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
