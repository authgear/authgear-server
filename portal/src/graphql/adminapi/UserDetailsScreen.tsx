import React, { useMemo, useState, useCallback, useContext } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Pivot,
  PivotItem,
  IButtonProps,
  ICommandBarItemProps,
  CommandButton,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import DeleteUserDialog from "./DeleteUserDialog";
import SetUserDisabledDialog from "./SetUserDisabledDialog";
import UserDetailSummary from "./UserDetailSummary";
import UserDetailsStandardAttributes, {
  StandardAttributesState,
} from "./UserDetailsStandardAttributes";
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
import { extractUserInfoFromIdentities } from "../../util/user";
import { PortalAPIAppConfig, StandardAttributes } from "../../types";

import styles from "./UserDetailsScreen.module.scss";

interface UserDetailsProps {
  form: SimpleFormModel<FormState>;
  data: UserQuery_node_User | null;
  appConfig: PortalAPIAppConfig | null;
}

const STANDARD_ATTRIBUTES_KEY = "standard-attributes";
const ACCOUNT_SECURITY_PIVOT_KEY = "account-security";
const CONNECTED_IDENTITIES_PIVOT_KEY = "connected-identities";
const SESSION_PIVOT_KEY = "session";

interface FormState {
  userID: string;
  standardAttributes: StandardAttributesState;
}

// eslint-disable-next-line complexity
function makeState(attrs: StandardAttributes): StandardAttributesState {
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
    locale: attrs.zoneinfo ?? "",
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

const UserDetails: React.FC<UserDetailsProps> = function UserDetails(
  props: UserDetailsProps
) {
  const { selectedKey, onLinkClick } = usePivotNavigation([
    STANDARD_ATTRIBUTES_KEY,
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

  const verifiedClaims = data?.verifiedClaims ?? [];

  const identities =
    data?.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
    [];
  const userInfo = extractUserInfoFromIdentities(identities);

  const authenticators =
    data?.authenticators?.edges
      ?.map((edge) => edge?.node)
      .filter(nonNullable) ?? [];

  const sessions =
    data?.sessions?.edges?.map((edge) => edge?.node).filter(nonNullable) ?? [];

  return (
    <div className={styles.userDetails}>
      <UserDetailSummary
        userInfo={userInfo}
        createdAtISO={data?.createdAt ?? null}
        lastLoginAtISO={data?.lastLoginAt ?? null}
      />
      <div className={styles.userDetailsTab}>
        <Pivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
          <PivotItem
            itemKey={STANDARD_ATTRIBUTES_KEY}
            headerText={renderToString(
              "UserDetails.standard-attributes.header"
            )}
          >
            <UserDetailsStandardAttributes
              identities={identities}
              standardAttributes={state.standardAttributes}
              onChangeStandardAttributes={onChangeStandardAttributes}
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
            headerText={renderToString(
              "UserDetails.connected-identities.header"
            )}
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
      text: renderToString("remove"),
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
  userIsDisabled: boolean,
  onClick: IButtonProps["onClick"]
): ICommandBarItemProps {
  const { renderToString } = useContext(Context);
  const itemProps: ICommandBarItemProps = useMemo(() => {
    return {
      key: "setDisabledStatus",
      text: userIsDisabled
        ? renderToString("enable")
        : renderToString("disable"),
      iconProps: {
        iconName: userIsDisabled ? "Play" : "CircleStop",
      },
      onRender: (props) => {
        return <CommandButton {...props} onClick={onClick} />;
      },
    };
  }, [userIsDisabled, onClick, renderToString]);
  return itemProps;
}

interface UserDetailsScreenContentProps {
  user: UserQuery_node_User;
  effectiveAppConfig: PortalAPIAppConfig;
}

const UserDetailsScreenContent: React.FC<UserDetailsScreenContentProps> =
  // eslint-disable-next-line complexity
  function UserDetailsScreenContent(props: UserDetailsScreenContentProps) {
    const { user, effectiveAppConfig } = props;
    const navigate = useNavigate();

    const identities =
      user.identities?.edges?.map((edge) => edge?.node).filter(nonNullable) ??
      [];
    const { username, email, phone } =
      extractUserInfoFromIdentities(identities);

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
      user.isDisabled,
      onClickSetUserDisabled
    );

    const farItems: ICommandBarItemProps[] = useMemo(() => {
      return [deleteUserCommandBarItem, setUserDisabledCommandBarItem];
    }, [deleteUserCommandBarItem, setUserDisabledCommandBarItem]);

    const defaultFormState = useMemo(() => {
      return {
        userID: user.id,
        standardAttributes: makeState(user.standardAttributes),
      };
    }, [user.id, user.standardAttributes]);

    const { updateUser } = useUpdateUserMutation();

    const submit = useCallback(
      async (state: FormState) => {
        await updateUser(state.userID, state.standardAttributes);
      },
      [updateUser]
    );

    const form = useSimpleForm<FormState>(defaultFormState, submit);

    return (
      <FormContainer form={form} farItems={farItems}>
        <main className={styles.root}>
          <NavBreadcrumb items={navBreadcrumbItems} />
          <UserDetails form={form} data={user} appConfig={effectiveAppConfig} />
        </main>
        <DeleteUserDialog
          isHidden={deleteUserDialogIsHidden}
          onDismiss={onDismissDeleteUserDialog}
          userID={user.id}
          username={username ?? email ?? phone}
        />
        <SetUserDisabledDialog
          isHidden={setUserDisabledDialogIsHidden}
          onDismiss={onDismissSetUserDisabledDialog}
          userID={user.id}
          username={username ?? email ?? phone}
          isDisablingUser={!user.isDisabled}
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
      effectiveAppConfig={effectiveAppConfig}
    />
  );
};

export default UserDetailsScreen;
