import React, { useMemo, useState, useCallback, useContext } from "react";
import { DateTime } from "luxon";
import { useParams, useNavigate } from "react-router-dom";
import {
  Pivot,
  PivotItem,
  IButtonProps,
  ICommandBarItemProps,
  CommandButton,
  MessageBar,
  MessageBarType,
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

import { useSystemConfig } from "../../context/SystemConfigContext";
import { useUpdateUserMutation } from "./mutations/updateUserMutation";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { usePivotNavigation } from "../../hook/usePivot";
import { nonNullable } from "../../util/types";
import { getEndUserAccountIdentifier } from "../../util/user";
import {
  PortalAPIAppConfig,
  StandardAttributes,
  CustomAttributes,
  AccessControlLevelString,
  CustomAttributesAttributeConfig,
} from "../../types";
import { parseJSONPointer } from "../../util/jsonpointer";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./UserDetailsScreen.module.scss";

interface UserDetailsProps {
  form: SimpleFormModel<FormState>;
  data: UserQuery_node_User | null;
  appConfig: PortalAPIAppConfig | null;
}

const USER_PROFILE_KEY = "user-profile";
const ACCOUNT_SECURITY_PIVOT_KEY = "account-security";
const CONNECTED_IDENTITIES_PIVOT_KEY = "connected-identities";
const SESSION_PIVOT_KEY = "session";

interface FormState {
  userID: string;
  standardAttributes: StandardAttributesState;
  customAttributes: CustomAttributesState;
}

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

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { selectedKey, onLinkClick } = usePivotNavigation([
    USER_PROFILE_KEY,
    ACCOUNT_SECURITY_PIVOT_KEY,
    CONNECTED_IDENTITIES_PIVOT_KEY,
    SESSION_PIVOT_KEY,
  ]);
  const { form, data, appConfig } = props;
  const { state, setState } = form;
  const { renderToString } = React.useContext(Context);

  const availableLoginIdIdentities = useMemo(() => {
    const authenticationIdentities =
      appConfig?.authentication?.identities ?? [];
    const loginIdIdentityEnabled =
      authenticationIdentities.includes("login_id");
    if (!loginIdIdentityEnabled) {
      return [];
    }
    const rawLoginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
    return rawLoginIdKeys.map((loginIdKey) => loginIdKey.key);
  }, [appConfig]);

  const standardAttributeAccessControl = useMemo(() => {
    const record: Record<string, AccessControlLevelString> = {};
    for (const item of appConfig?.user_profile?.standard_attributes
      ?.access_control ?? []) {
      record[item.pointer] = item.access_control.portal_ui;
    }
    return record;
  }, [appConfig]);

  const customAttributesConfig: CustomAttributesAttributeConfig[] =
    useMemo(() => {
      return appConfig?.user_profile?.custom_attributes?.attributes ?? [];
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

  const verifiedClaims = data?.verifiedClaims ?? [];

  const identities =
    data?.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
    [];

  const authenticators =
    data?.authenticators?.edges
      ?.map((edge) => edge?.node)
      .filter(nonNullable) ?? [];

  const sessions =
    data?.sessions?.edges?.map((edge) => edge?.node).filter(nonNullable) ?? [];

  const endUserAccountIdentifier = getEndUserAccountIdentifier(
    data?.standardAttributes ?? {}
  );

  return (
    <div className={styles.widget}>
      <UserDetailSummary
        profileImageURL={data?.standardAttributes.picture}
        formattedName={data?.formattedName ?? undefined}
        endUserAccountIdentifier={endUserAccountIdentifier}
        createdAtISO={data?.createdAt ?? null}
        lastLoginAtISO={data?.lastLoginAt ?? null}
      />
      <Pivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
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
          <UserDetailsAccountSecurity authenticators={authenticators} />
        </PivotItem>
        <PivotItem
          itemKey={CONNECTED_IDENTITIES_PIVOT_KEY}
          headerText={renderToString("UserDetails.connected-identities.header")}
        >
          <UserDetailsConnectedIdentities
            identities={identities}
            verifiedClaims={verifiedClaims}
            availableLoginIdIdentities={availableLoginIdIdentities}
          />
        </PivotItem>
        <PivotItem
          itemKey={SESSION_PIVOT_KEY}
          headerText={renderToString("UserDetails.session.header")}
        >
          <UserDetailsSession sessions={sessions} />
        </PivotItem>
      </Pivot>
    </div>
  );
};

function useDeleteUserCommandBarItem(
  onClick: IButtonProps["onClick"]
): ICommandBarItemProps {
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();

  const itemProps: ICommandBarItemProps = useMemo(() => {
    return {
      key: "remove",
      text: renderToString("UserDetailsScreen.remove-user"),
      iconProps: { iconName: "Delete" },
      onRender: (props) => {
        return (
          <CommandButton
            {...props}
            theme={themes.destructive}
            onClick={onClick}
          />
        );
      },
    };
  }, [onClick, renderToString, themes.destructive]);

  return itemProps;
}

function useSetUserDisabledCommandBarItem(
  user: UserQuery_node_User,
  onClick: IButtonProps["onClick"]
): ICommandBarItemProps {
  const { renderToString } = useContext(Context);
  const itemProps: ICommandBarItemProps = useMemo(() => {
    const text =
      user.deleteAt != null
        ? renderToString("UserDetailsScreen.cancel-removal")
        : user.isDisabled
        ? renderToString("UserDetailsScreen.reenable-user")
        : renderToString("UserDetailsScreen.disable-user");
    const iconName =
      user.deleteAt != null ? "Undo" : user.isDisabled ? "Play" : "CircleStop";
    return {
      key: "setDisabledStatus",
      text,
      iconProps: {
        iconName,
      },
      onRender: (props) => {
        return <CommandButton {...props} onClick={onClick} />;
      },
    };
  }, [user.deleteAt, user.isDisabled, onClick, renderToString]);
  return itemProps;
}

interface WarnScheduledDeletionProps {
  user: UserQuery_node_User;
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
          date:
            formatDatetime(locale, user.deleteAt, DateTime.DATE_SHORT) ?? "",
        }}
      />
    </MessageBar>
  );
}

interface UserDetailsScreenContentProps {
  user: UserQuery_node_User;
  refreshUser?: () => void;
  effectiveAppConfig: PortalAPIAppConfig;
}

const UserDetailsScreenContent: React.FC<UserDetailsScreenContentProps> =
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
        { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
        { to: ".", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      ];
    }, []);

    const [deleteUserDialogIsHidden, setDeleteUserDialogIsHidden] =
      useState(true);
    const onDismissDeleteUserDialog = useCallback(
      (deletedUser: boolean) => {
        setDeleteUserDialogIsHidden(true);
        if (deletedUser) {
          setTimeout(() => navigate("../.."), 0);
        }
      },
      [navigate]
    );
    const onClickDeleteUser = useCallback(() => {
      setDeleteUserDialogIsHidden(false);
    }, []);

    const [setUserDisabledDialogIsHidden, setSetUserDisabledDialogIsHidden] =
      useState(true);
    const onDismissSetUserDisabledDialog = useCallback(() => {
      setSetUserDisabledDialogIsHidden(true);
    }, []);
    const deleteUserCommandBarItem =
      useDeleteUserCommandBarItem(onClickDeleteUser);
    const onClickSetUserDisabled = useCallback(() => {
      setSetUserDisabledDialogIsHidden(false);
    }, []);

    const setUserDisabledCommandBarItem = useSetUserDisabledCommandBarItem(
      user,
      onClickSetUserDisabled
    );

    const primaryItems: ICommandBarItemProps[] = useMemo(() => {
      return [deleteUserCommandBarItem, setUserDisabledCommandBarItem];
    }, [deleteUserCommandBarItem, setUserDisabledCommandBarItem]);

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

    const endUserAccountIdentifier = useMemo(
      () => getEndUserAccountIdentifier(user.standardAttributes),
      [user.standardAttributes]
    );

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
        form={form}
        primaryItems={primaryItems}
        messageBar={<WarnScheduledDeletion user={user} />}
      >
        <ScreenContent>
          <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
          <UserDetails form={form} data={user} appConfig={effectiveAppConfig} />
        </ScreenContent>
        <DeleteUserDialog
          isHidden={deleteUserDialogIsHidden}
          onDismiss={onDismissDeleteUserDialog}
          userID={user.id}
          userDeleteAt={user.deleteAt}
          endUserAccountIdentifier={endUserAccountIdentifier}
        />
        <SetUserDisabledDialog
          isHidden={setUserDisabledDialogIsHidden}
          onDismiss={onDismissSetUserDisabledDialog}
          userID={user.id}
          userIsDisabled={user.isDisabled}
          userDeleteAt={user.deleteAt}
          endUserAccountIdentifier={endUserAccountIdentifier}
        />
      </FormContainer>
    );
  };

const UserDetailsScreen: React.FC = function UserDetailsScreen() {
  const { appID, userID } = useParams();
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
