import React, { useCallback, useMemo, useState } from "react";
import { Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import { FormattedMessage } from "@oursky/react-messageformat";

import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import AddIdentityForm from "./AddIdentityForm";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { useTextField } from "../../hook/useInput";
import { FormContext } from "../../error/FormContext";
import { useValidationError } from "../../error/useValidationError";
import { useGenericError } from "../../error/useGenericError";
import { passwordFieldErrorRules } from "../../PasswordField";
import { PortalAPIAppConfig } from "../../types";
import { canCreateLoginIDIdentity } from "../../util/loginID";

import styles from "./AddEmailScreen.module.scss";

interface AddEmailFormProps {
  appConfig: PortalAPIAppConfig | null;
  user: UserQuery_node_User | null;
  resetForm: () => void;
}

interface AddEmailFormData {
  email: string;
  password: string;
}

const AddEmailForm: React.FC<AddEmailFormProps> = function AddEmailForm(
  props: AddEmailFormProps
) {
  const { resetForm, appConfig, user } = props;
  const { userID } = useParams();

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);

  const initialFormData = useMemo(() => {
    return {
      email: "",
      password: "",
    };
  }, []);
  const [formData, setFormData] = useState<AddEmailFormData>(initialFormData);
  const { email, password } = formData;

  const { onChange: onEmailChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, email: value }));
  });
  const { onChange: onPasswordChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, password: value }));
  });

  const isFormModified = useMemo(() => {
    return !deepEqual(initialFormData, formData);
  }, [initialFormData, formData]);

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
      errorMessageID: "AddEmailScreen.error.duplicated-email",
      field: "email",
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
      {unrecognizedError && (
        <div className={styles.error}>
          <ShowError error={unrecognizedError} />
        </div>
      )}
      <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      <AddIdentityForm
        className={styles.content}
        appConfig={appConfig}
        user={user}
        loginIDKey="email"
        loginID={email}
        loginIDField={
          <FormTextField
            jsonPointer=""
            parentJSONPointer=""
            fieldName="email"
            fieldNameMessageID="AddEmailScreen.email.label"
            className={styles.emailField}
            value={email}
            onChange={onEmailChange}
            errorMessage={errorMessageMap.email}
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
  } = useAppConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddEmailScreen.title" /> },
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
        <AddEmailForm
          key={remountIdentifier}
          appConfig={effectiveAppConfig}
          user={user}
          resetForm={resetForm}
        />
      </ModifiedIndicatorWrapper>
    </div>
  );
};

export default AddEmailScreen;
