import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import deepEqual from "deep-equal";

import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import UserDetailCommandBar from "./UserDetailCommandBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import PasswordField, {
  handleLocalPasswordViolations,
  handlePasswordPolicyViolatedViolation,
  localValidatePassword,
} from "../../PasswordField";
import ButtonWithLoading from "../../ButtonWithLoading";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { useTextField } from "../../hook/useInput";
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";
import { parseError } from "../../util/error";
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

  const [localViolations, setLocalViolations] = useState<Violation[]>([]);
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
  }, [initialFormData]);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      const newLocalViolations: Violation[] = [];
      if (isPasswordRequired) {
        localValidatePassword(newLocalViolations, passwordPolicy, password);
      }
      setLocalViolations(newLocalViolations);
      if (newLocalViolations.length > 0) {
        return;
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
    [username, createIdentity, isPasswordRequired, password, passwordPolicy]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("..#connected-identities");
    }
  }, [submittedForm, navigate]);

  const { errorMessages, unhandledViolations } = useMemo(() => {
    const violations =
      localViolations.length > 0
        ? localViolations
        : parseError(createIdentityError);

    const usernameFieldErrorMessages: string[] = [];
    const passwordFieldErrorMessages: string[] = [];
    const unhandledViolations: Violation[] = [];
    for (const violation of violations) {
      if (violation.kind === "Invalid" || violation.kind === "format") {
        usernameFieldErrorMessages.push(
          renderToString("AddUsernameScreen.error.invalid-username")
        );
      } else if (violation.kind === "DuplicatedIdentity") {
        usernameFieldErrorMessages.push(
          renderToString("AddUsernameScreen.error.duplicated-username")
        );
      } else if (violation.kind === "custom") {
        handleLocalPasswordViolations(
          renderToString,
          violation,
          passwordFieldErrorMessages,
          null,
          unhandledViolations
        );
      } else if (violation.kind === "PasswordPolicyViolated") {
        handlePasswordPolicyViolatedViolation(
          renderToString,
          violation,
          passwordFieldErrorMessages,
          unhandledViolations
        );
      } else {
        unhandledViolations.push(violation);
      }
    }

    const errorMessages = {
      username: defaultFormatErrorMessageList(usernameFieldErrorMessages),
      password: defaultFormatErrorMessageList(passwordFieldErrorMessages),
    };

    return { errorMessages, unhandledViolations };
  }, [createIdentityError, localViolations, renderToString]);

  return (
    <form className={styles.content} onSubmit={onFormSubmit}>
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      {unhandledViolations.length > 0 && (
        <ShowError error={createIdentityError} />
      )}
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
      <TextField
        className={styles.usernameField}
        label={renderToString("AddUsernameScreen.username.label")}
        value={username}
        onChange={onUsernameChange}
        errorMessage={errorMessages.username}
      />
      {isPasswordRequired && (
        <PasswordField
          className={styles.password}
          textFieldClassName={styles.passwordField}
          passwordPolicy={passwordPolicy}
          label={renderToString("AddUsernameScreen.password.label")}
          value={password}
          onChange={onPasswordChange}
          errorMessage={errorMessages.password}
        />
      )}
      <ButtonWithLoading
        type="submit"
        disabled={!isFormModified || submittedForm}
        labelId="add"
        loading={creatingIdentity}
      />
    </form>
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
