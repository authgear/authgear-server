import React, { useCallback, useMemo, useContext } from "react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import FormTextField from "../../FormTextField";
import AddIdentityForm from "./AddIdentityForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { ErrorParseRule } from "../../error/parse";

import styles from "./AddUsernameScreen.module.scss";

const errorRules: ErrorParseRule[] = [
  {
    reason: "InvariantViolated",
    kind: "DuplicatedIdentity",
    errorMessageID: "AddUsernameScreen.error.duplicated-username",
  },
];

interface UsernameFieldProps {
  value: string;
  onChange: (value: string) => void;
}

const UsernameField: React.FC<UsernameFieldProps> = function UsernameField(
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
      label={renderToString("AddUsernameScreen.username.label")}
      className={styles.usernameField}
      value={value}
      onChange={onUsernameChange}
      errorRules={errorRules}
    />
  );
};

const AddUsernameScreen: React.FC = function AddUsernameScreen() {
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
      { to: ".", label: <FormattedMessage id="AddUsernameScreen.title" /> },
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
      loginIDType="username"
      title={title}
      loginIDField={UsernameField}
    />
  );
};

export default AddUsernameScreen;
