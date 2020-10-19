import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import deepEqual from "deep-equal";

import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import UserDetailCommandBar from "./UserDetailCommandBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import PasswordField, {
  localValidatePassword,
  passwordFieldErrorRules,
} from "../../PasswordField";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import FormTextField from "../../FormTextField";
import ButtonWithLoading from "../../ButtonWithLoading";
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
import { nonNullable } from "../../util/types";
import { AuthenticatorType } from "./__generated__/globalTypes";
import { PortalAPIAppConfig } from "../../types";

import styles from "./AddUsernameScreen.module.scss";

interface AddUsernameFormProps {
  appConfig: PortalAPIAppConfig | null;
  user: UserQuery_node_User | null;
}

function determineIsPasswordRequired(user: UserQuery_node_User | null) {
  const authenticators =
    user?.authenticators?.edges
      ?.map((edge) => edge?.node?.type)
      .filter(nonNullable) ?? [];
  const hasPasswordAuthenticator = authenticators.includes(
    AuthenticatorType.PASSWORD
  );
  return !hasPasswordAuthenticator;
}

const AddUsernameForm: React.FC<AddUsernameFormProps> = function AddUsernameForm(
  props: AddUsernameFormProps
) {
  const { appConfig, user } = props;

  const { userID } = useParams();
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);

  const isPasswordRequired = useMemo(() => {
    return determineIsPasswordRequired(user);
  }, [user]);

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);

  const [localValidationErrorMessage, setLocalViolationErrorMessage] = useState<
    string | undefined
  >(undefined);
  const [submittedForm, setSubmittedForm] = useState(false);

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

  const resetForm = useCallback(() => {
    setFormData(initialFormData);
    setLocalViolationErrorMessage(undefined);
  }, [initialFormData]);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      if (isPasswordRequired) {
        const localErrorMessageMap = localValidatePassword(
          renderToString,
          passwordPolicy,
          password
        );
        setLocalViolationErrorMessage(localErrorMessageMap?.password);

        if (localErrorMessageMap != null) {
          return;
        }
      }

      const requestPassword = isPasswordRequired ? password : undefined;
      createIdentity({ key: "username", value: username }, requestPassword)
        .then((identity) => {
          if (identity != null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [
      renderToString,
      username,
      createIdentity,
      isPasswordRequired,
      password,
      passwordPolicy,
    ]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("..#connected-identities");
    }
  }, [submittedForm, navigate]);

  const {
    unhandledCauses,
    otherError,
    value: formContextValue,
  } = useValidationError(createIdentityError);

  const { errorMessageMap, unrecognizedError } = useGenericError(otherError, [
    {
      reason: "InvariantViolated",
      kind: "DuplicatedIdentity",
      errorMessageID: "AddUsernameScreen.error.duplicated-username",
      field: "username",
    },
    ...passwordFieldErrorRules,
  ]);

  return (
    <FormContext.Provider value={formContextValue}>
      <form className={styles.content} onSubmit={onFormSubmit}>
        <ModifiedIndicatorPortal
          resetForm={resetForm}
          isModified={isFormModified}
        />
        {unrecognizedError && <ShowError error={unrecognizedError} />}
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        <NavigationBlockerDialog
          blockNavigation={!submittedForm && isFormModified}
        />
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
        {isPasswordRequired && (
          <PasswordField
            className={styles.password}
            textFieldClassName={styles.passwordField}
            passwordPolicy={passwordPolicy}
            label={renderToString("AddUsernameScreen.password.label")}
            value={password}
            onChange={onPasswordChange}
            errorMessage={
              localValidationErrorMessage ?? errorMessageMap.password
            }
          />
        )}
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified || submittedForm}
          labelId="add"
          loading={creatingIdentity}
        />
      </form>
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
      <UserDetailCommandBar />
      <ModifiedIndicatorWrapper className={styles.wrapper}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <AddUsernameForm appConfig={effectiveAppConfig} user={user} />
      </ModifiedIndicatorWrapper>
    </div>
  );
};

export default AddUsernameScreen;
