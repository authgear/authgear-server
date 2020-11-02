import React, { useCallback, useMemo, useState } from "react";
import { Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import deepEqual from "deep-equal";

import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import NavBreadcrumb from "../../NavBreadcrumb";
import { passwordFieldErrorRules } from "../../PasswordField";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import FormTextField from "../../FormTextField";
import AddIdentityForm from "./AddIdentityForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { useTextField } from "../../hook/useInput";
import { useValidationError } from "../../error/useValidationError";
import { FormContext } from "../../error/FormContext";
import { useGenericError } from "../../error/useGenericError";
import { PortalAPIAppConfig } from "../../types";
import { canCreateLoginIDIdentity } from "../../util/loginID";

import styles from "./AddUsernameScreen.module.scss";

interface AddUsernameFormProps {
  appConfig: PortalAPIAppConfig | null;
  user: UserQuery_node_User | null;
  resetForm: () => void;
}

const AddUsernameForm: React.FC<AddUsernameFormProps> = function AddUsernameForm(
  props: AddUsernameFormProps
) {
  const { appConfig, user, resetForm } = props;
  const { userID } = useParams();

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);

  const initialFormData = useMemo(() => {
    return {
      password: "",
      username: "",
    };
  }, []);

  const [formData, setFormData] = useState(initialFormData);
  const { username, password } = formData;

  const { onChange: onUsernameChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, username: value }));
  });
  const { onChange: onPasswordChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, password: value }));
  });

  const isFormModified = useMemo(() => {
    return !deepEqual(formData, initialFormData);
  }, [formData, initialFormData]);

  const {
    unhandledCauses: rawUnhandledCauses,
    otherError,
    value: formContextValue,
  } = useValidationError(createIdentityError);

  const {
    errorMessageMap,
    unrecognizedError,
    unhandledCauses,
  } = useGenericError(otherError, rawUnhandledCauses, [
    {
      reason: "InvariantViolated",
      kind: "DuplicatedIdentity",
      errorMessageID: "AddUsernameScreen.error.duplicated-username",
      field: "username",
    },
    ...passwordFieldErrorRules,
  ]);

  if (!canCreateLoginIDIdentity(appConfig)) {
    return (
      <Text className={styles.helpText}>
        <FormattedMessage id="CreateIdentity.require-login-id" />
      </Text>
    );
  }

  return (
    <FormContext.Provider value={formContextValue}>
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      {unrecognizedError && <ShowError error={unrecognizedError} />}
      <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
      <AddIdentityForm
        className={styles.content}
        appConfig={appConfig}
        user={user}
        loginIDKey="username"
        loginID={username}
        loginIDField={
          <FormTextField
            jsonPointer=""
            parentJSONPointer=""
            fieldName="username"
            fieldNameMessageID="AddUsernameScreen.username.label"
            className={styles.usernameField}
            value={username}
            onChange={onUsernameChange}
            errorMessage={errorMessageMap.username}
          />
        }
        password={password}
        onPasswordChange={onPasswordChange}
        passwordFieldErrorMessage={errorMessageMap.password}
        isFormModified={isFormModified}
        createIdentity={createIdentity}
        creatingIdentity={creatingIdentity}
      />
    </FormContext.Provider>
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
  } = useAppConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUsernameScreen.title" /> },
    ];
  }, []);

  const [remountIdentifier, setRemountIdentifier] = useState(0);
  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);

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
    <div className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.wrapper}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <AddUsernameForm
          key={remountIdentifier}
          appConfig={effectiveAppConfig}
          user={user}
          resetForm={resetForm}
        />
      </ModifiedIndicatorWrapper>
    </div>
  );
};

export default AddUsernameScreen;
