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
import {
  ErrorParseRule,
  makeInvariantViolatedErrorParseRule,
} from "../../error/parse";

import styles from "./AddEmailScreen.module.css";

const errorRules: ErrorParseRule[] = [
  makeInvariantViolatedErrorParseRule(
    "DuplicatedIdentity",
    "AddEmailScreen.error.duplicated-email"
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
      label={renderToString("AddEmailScreen.email.label")}
      value={value}
      onChange={onEmailChange}
      errorRules={errorRules}
    />
  );
};

const AddEmailScreen: React.VFC = function AddEmailScreen() {
  const { appID, userID } = useParams() as { appID: string; userID: string };
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
      { to: ".", label: <FormattedMessage id="AddEmailScreen.title" /> },
    ];
  }, [user?.id]);
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
      originalIdentityID={null}
      appConfig={effectiveAppConfig}
      rawUser={user}
      loginIDType="email"
      title={title}
      loginIDField={EmailField}
    />
  );
};

export default AddEmailScreen;
