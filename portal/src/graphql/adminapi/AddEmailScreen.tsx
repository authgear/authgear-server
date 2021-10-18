import React, { useCallback, useMemo, useContext } from "react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import AddIdentityForm from "./AddIdentityForm";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { ErrorParseRule } from "../../error/parse";

import styles from "./AddEmailScreen.module.scss";

const errorRules: ErrorParseRule[] = [
  {
    reason: "InvariantViolated",
    kind: "DuplicatedIdentity",
    errorMessageID: "AddEmailScreen.error.duplicated-email",
  },
];

interface EmailFieldProps {
  value: string;
  onChange: (value: string) => void;
}

const EmailField: React.FC<EmailFieldProps> = function EmailField(props) {
  const { value, onChange } = props;
  const { renderToString } = useContext(Context);
  const onEmailChange = useCallback(
    (_, value?: string) => onChange(value ?? ""),
    [onChange]
  );
  return (
    <FormTextField
      parentJSONPointer=""
      fieldName="login_id"
      label={renderToString("AddEmailScreen.email.label")}
      className={styles.emailField}
      value={value}
      onChange={onEmailChange}
      errorRules={errorRules}
    />
  );
};

const AddEmailScreen: React.FC = function AddEmailScreen() {
  const { appID, userID } = useParams();
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
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddEmailScreen.title" /> },
    ];
  }, []);
  const title = <NavBreadcrumb items={navBreadcrumbItems} />;

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
    <AddIdentityForm
      appConfig={effectiveAppConfig}
      rawUser={user}
      loginIDType="email"
      title={title}
      loginIDField={EmailField}
    />
  );
};

export default AddEmailScreen;
