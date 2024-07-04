import React, { useCallback, useMemo, useContext } from "react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import FormTextField from "../../FormTextField";
import IdentityForm from "./IdentityForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";

import styles from "./UsernameScreen.module.css";

const errorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "DuplicatedIdentity",
    "UsernameScreen.error.duplicated-username"
  ),
];

interface UsernameFieldProps {
  value: string;
  onChange: (value: string) => void;
}

const UsernameField: React.VFC<UsernameFieldProps> = function UsernameField(
  props
) {
  const { value, onChange } = props;
  const { renderToString } = useContext(Context);
  const onUsernameChange = useCallback(
    (_, value?: string) => onChange(value ?? ""),
    [onChange]
  );
  return (
    <FormTextField
      parentJSONPointer=""
      fieldName="login_id"
      label={renderToString("UsernameScreen.username.label")}
      className={styles.widget}
      value={value}
      onChange={onUsernameChange}
      errorRules={errorRules}
    />
  );
};

const UsernameScreen: React.VFC = function UsernameScreen() {
  const { appID, userID, identityID } = useParams() as {
    appID: string;
    userID: string;
    identityID?: string;
  };
  const {
    user,
    loading: loadingUser,
    error: userError,
    refetch: refetchUser,
  } = useUserQuery(userID);
  const {
    effectiveAppConfig,
    loading: loadingAppConfig,
    error: appConfigError,
    refetch: refetchAppConfig,
  } = useAppAndSecretConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      {
        to: `~/users/${user?.id}/details`,
        label: <FormattedMessage id="UserDetailsScreen.title" />,
      },
      {
        to: ".",
        label: identityID ? (
          <FormattedMessage id="UsernameScreen.edit.title" />
        ) : (
          <FormattedMessage id="UsernameScreen.add.title" />
        ),
      },
    ];
  }, [identityID, user?.id]);

  const originalIdentity = useMemo(() => {
    if (!identityID) {
      return null;
    }
    const identity = user?.identities?.edges?.find((edge) => {
      const node = edge?.node;
      return (
        node != null && node.id === identityID && node.claims.preferred_username
      );
    });
    if (identity == null) {
      return null;
    }
    return {
      id: identity.node!.id,
      value: identity.node!.claims.preferred_username!,
    };
  }, [identityID, user?.identities?.edges]);

  const currentValueMessage = useMemo(() => {
    if (originalIdentity == null) {
      return null;
    }
    return (
      <FormattedMessage
        id="UsernameScreen.edit.current-value"
        values={{ value: originalIdentity.value }}
      />
    );
  }, [originalIdentity]);

  const title = (
    <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
  );

  if (loadingUser || loadingAppConfig) {
    return <ShowLoading />;
  }

  if (userError != null) {
    return <ShowError error={userError} onRetry={refetchUser} />;
  }

  if (appConfigError != null) {
    return <ShowError error={appConfigError} onRetry={refetchAppConfig} />;
  }

  return (
    <IdentityForm
      originalIdentityID={originalIdentity?.id ?? null}
      currentValueMessage={currentValueMessage}
      appConfig={effectiveAppConfig}
      rawUser={user}
      loginIDType="username"
      title={title}
      loginIDField={UsernameField}
    />
  );
};

export default UsernameScreen;
