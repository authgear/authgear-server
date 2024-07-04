import React, { useCallback, useMemo, useContext } from "react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import IdentityForm from "./IdentityForm";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";

import styles from "./EmailScreen.module.css";

const errorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "DuplicatedIdentity",
    "EmailScreen.error.duplicated-email"
  ),
];

interface EmailFieldProps {
  value: string;
  onChange: (value: string) => void;
}

const EmailField: React.VFC<EmailFieldProps> = function EmailField(props) {
  const { value, onChange } = props;
  const { renderToString } = useContext(Context);
  const onEmailChange = useCallback(
    (_, value?: string) => onChange(value ?? ""),
    [onChange]
  );
  return (
    <FormTextField
      className={styles.widget}
      parentJSONPointer=""
      fieldName="login_id"
      label={renderToString("EmailScreen.email.label")}
      value={value}
      onChange={onEmailChange}
      errorRules={errorRules}
    />
  );
};

const EmailScreen: React.VFC = function EmailScreen() {
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
          <FormattedMessage id="EmailScreen.edit.title" />
        ) : (
          <FormattedMessage id="EmailScreen.add.title" />
        ),
      },
    ];
  }, [identityID, user?.id]);
  const title = (
    <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
  );

  const originalIdentity = useMemo(() => {
    if (!identityID) {
      return null;
    }
    const identity = user?.identities?.edges?.find((edge) => {
      const node = edge?.node;
      return node != null && node.id === identityID && node.claims.email;
    });
    if (identity == null) {
      return null;
    }
    return {
      id: identity.node!.id,
      value: identity.node!.claims.email!,
    };
  }, [identityID, user?.identities?.edges]);

  const currentValueMessage = useMemo(() => {
    if (originalIdentity == null) {
      return null;
    }
    return (
      <FormattedMessage
        id="EmailScreen.edit.current-value"
        values={{ value: originalIdentity.value }}
      />
    );
  }, [originalIdentity]);

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
      loginIDType="email"
      title={title}
      loginIDField={EmailField}
    />
  );
};

export default EmailScreen;
