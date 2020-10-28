import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Text,
  TextField,
  ChoiceGroup,
  IChoiceGroupOption,
  Label,
} from "@fluentui/react";

import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useCreateUserMutation } from "./mutations/createUserMutation";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import PasswordField, {
  localValidatePassword,
  passwordFieldErrorRules,
} from "../../PasswordField";
import { useTextField } from "../../hook/useInput";
import { PortalAPIAppConfig } from "../../types";
import { useGenericError } from "../../error/useGenericError";

import styles from "./AddUserScreen.module.scss";

type LoginIDKey = "username" | "email" | "phone";
function isLoginIDKey(value?: string): value is LoginIDKey {
  return ["username", "email", "phone"].includes(value ?? "");
}

interface AddUserContentProps {
  appConfig: PortalAPIAppConfig | null;
  resetForm: () => void;
}

interface AddUserFormState {
  selectedLoginIdKey?: LoginIDKey;
  username: string;
  email: string;
  phone: string;
  password: string;
}

interface LoginIdIdentityOptionProps {
  option?: IChoiceGroupOption;
  renderTextField?: () => React.ReactNode;
}

const loginIdLocaleKey: Record<LoginIDKey, string> = {
  username: "login-id-key.username",
  email: "login-id-key.email",
  phone: "login-id-key.phone",
};

function determineIsPasswordNeeded(
  appConfig: PortalAPIAppConfig | null,
  loginIdKeySelected: LoginIDKey | undefined
) {
  if (loginIdKeySelected == null) {
    return false;
  }
  const primaryAuthenticators =
    appConfig?.authentication?.primary_authenticators ?? [];
  // password is first one, all login ID types need password
  if (primaryAuthenticators[0] === "password") {
    return true;
  }
  // password is second one, require password if user choose username
  // only id is username -> need password
  if (deepEqual(["oob_otp", "password"], primaryAuthenticators)) {
    return loginIdKeySelected === "username";
  }
  return false;
}

function getLoginIdKeyOptions(appConfig: PortalAPIAppConfig | null) {
  const primaryAuthenticators =
    appConfig?.authentication?.primary_authenticators ?? [];

  // need password authenticator in order to use username to login
  const usernameAllowed = primaryAuthenticators.includes("password");

  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const enabledIdentities = appConfig?.authentication?.identities ?? [];
  const isLoginIDIdentityEnabled = enabledIdentities.includes("login_id");

  // if login ID identity is disabled
  // we cannot add login ID identity to new user
  if (!isLoginIDIdentityEnabled) {
    return [];
  }

  const loginIdKeyOptions = new Set<LoginIDKey>();
  for (const key of loginIdKeys) {
    switch (key.type) {
      case "username": {
        if (usernameAllowed) {
          loginIdKeyOptions.add("username");
        }
        break;
      }
      case "email":
        loginIdKeyOptions.add("email");
        break;
      case "phone":
        loginIdKeyOptions.add("phone");
        break;
      default:
        break;
    }
  }
  return Array.from(loginIdKeyOptions);
}

const LoginIdIdentityOption: React.FC<LoginIdIdentityOptionProps> = function (
  props: LoginIdIdentityOptionProps
) {
  const { option, renderTextField } = props;
  if (option == null) {
    return null;
  }
  return (
    <div className={styles.identityOption}>
      <Label className={styles.identityOptionLabel}>{option.text}</Label>
      {renderTextField?.()}
    </div>
  );
};

const AddUserContent: React.FC<AddUserContentProps> = function AddUserContent(
  props: AddUserContentProps
) {
  const { appConfig, resetForm } = props;
  const { renderToString } = useContext(Context);

  const navigate = useNavigate();
  const {
    createUser,
    loading: creatingUser,
    error: createUserError,
  } = useCreateUserMutation();

  const [
    localValidationErrorMessage,
    setLocalValidationErrorMessage,
  ] = useState<string | undefined>(undefined);

  const selectedLoginIdInLastSubmission = useRef<LoginIDKey | null>(null);
  const [submittedForm, setSubmittedForm] = useState(false);

  const initialFormState = useMemo<AddUserFormState>(() => {
    return {
      username: "",
      email: "",
      phone: "",
      password: "",
    };
  }, []);

  const [formState, setFormState] = useState(initialFormState);
  const { username, email, phone, password, selectedLoginIdKey } = formState;

  const isFormModified = useMemo(() => {
    return !deepEqual(initialFormState, formState);
  }, [initialFormState, formState]);

  const { onChange: onUsernameChange } = useTextField((value) => {
    setFormState((prev) => ({ ...prev, username: value }));
  });
  const { onChange: onEmailChange } = useTextField((value) => {
    setFormState((prev) => ({ ...prev, email: value }));
  });
  const { onChange: onPhoneChange } = useTextField((value) => {
    setFormState((prev) => ({ ...prev, phone: value }));
  });
  const { onChange: onPasswordChange } = useTextField((value) => {
    setFormState((prev) => ({ ...prev, password: value }));
  });

  const onSelectLoginIdKey = useCallback(
    (_event, options?: IChoiceGroupOption) => {
      const loginIdKey = options?.key;
      if (!isLoginIDKey(loginIdKey)) {
        return;
      }
      setFormState((prev) => ({ ...prev, selectedLoginIdKey: loginIdKey }));
    },
    []
  );

  const { errorMessageMap, unrecognizedError } = useGenericError(
    createUserError,
    [],
    [
      ...passwordFieldErrorRules,
      {
        reason: "InvariantViolated",
        kind: "DuplicatedIdentity",
        errorMessageID: "AddUserScreen.error.duplicated-identity",
        field: selectedLoginIdInLastSubmission.current ?? "",
      },
      // NOTE: workaround, validation error has no location
      // cannot distinguish which field fails the validation
      {
        reason: "ValidationFailed",
        jsonPointer: "",
        kind: "format",
        errorMessageID: "AddUserScreen.error.invalid-identity",
        field: selectedLoginIdInLastSubmission.current ?? "",
      },
    ]
  );
  const renderUsernameField = useCallback(() => {
    return (
      <TextField
        className={styles.textField}
        value={username}
        onChange={onUsernameChange}
        errorMessage={errorMessageMap.username}
      />
    );
  }, [username, onUsernameChange, errorMessageMap]);

  const renderEmailField = useCallback(() => {
    return (
      <TextField
        className={styles.textField}
        value={email}
        onChange={onEmailChange}
        errorMessage={errorMessageMap.email}
      />
    );
  }, [email, onEmailChange, errorMessageMap]);

  const renderPhoneField = useCallback(() => {
    return (
      <TextField
        className={styles.textField}
        value={phone}
        onChange={onPhoneChange}
        errorMessage={errorMessageMap.phone}
      />
    );
  }, [phone, onPhoneChange, errorMessageMap]);

  const textFieldRenderer: Record<LoginIDKey, () => React.ReactNode> = useMemo(
    () => ({
      username: renderUsernameField,
      email: renderEmailField,
      phone: renderPhoneField,
    }),
    [renderUsernameField, renderEmailField, renderPhoneField]
  );

  const passwordRequired = useMemo(() => {
    return determineIsPasswordNeeded(appConfig, selectedLoginIdKey);
  }, [appConfig, selectedLoginIdKey]);

  const loginIdKeyOptions: IChoiceGroupOption[] = useMemo(() => {
    const list = getLoginIdKeyOptions(appConfig);
    return list.map((loginIdKey) => {
      const messageId = loginIdLocaleKey[loginIdKey];
      const renderTextField =
        selectedLoginIdKey === loginIdKey
          ? textFieldRenderer[loginIdKey]
          : undefined;
      return {
        key: loginIdKey,
        text: renderToString(messageId),
        onRenderLabel: (option) => (
          <LoginIdIdentityOption
            option={option}
            renderTextField={renderTextField}
          />
        ),
      };
    });
  }, [appConfig, renderToString, textFieldRenderer, selectedLoginIdKey]);

  // NOTE: cannot add user identity if none of three field is available
  const canAddUser = loginIdKeyOptions.length > 0;

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  const screenState = useMemo(
    () => ({
      selectedLoginIdKey,
      username,
      email,
      phone,
      password,
    }),
    [selectedLoginIdKey, username, email, phone, password]
  );

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      const selectedKey = screenState.selectedLoginIdKey;
      const passwordValidationResult = passwordRequired
        ? localValidatePassword(
            renderToString,
            passwordPolicy,
            screenState.password
          )
        : null;
      if (passwordValidationResult != null || selectedKey == null) {
        setLocalValidationErrorMessage(passwordValidationResult?.password);
        return;
      }
      selectedLoginIdInLastSubmission.current = selectedKey;
      const identityValue = screenState[selectedKey];
      const password = passwordRequired ? screenState.password : undefined;
      createUser({ key: selectedKey, value: identityValue }, password)
        .then((userID) => {
          if (userID != null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [renderToString, screenState, passwordPolicy, passwordRequired, createUser]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("../");
    }
  }, [submittedForm, navigate]);

  // TODO: improve empty state
  if (!canAddUser) {
    return (
      <Text>
        <FormattedMessage id="AddUserScreen.cannot-add-user" />
      </Text>
    );
  }

  return (
    <form className={styles.content} onSubmit={onFormSubmit}>
      {unrecognizedError && <ShowError error={unrecognizedError} />}
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
      <ModifiedIndicatorPortal
        isModified={isFormModified}
        resetForm={resetForm}
      />
      <ChoiceGroup
        className={styles.userInfo}
        styles={{ label: { marginBottom: "15px", fontSize: "14px" } }}
        selectedKey={selectedLoginIdKey}
        options={loginIdKeyOptions}
        onChange={onSelectLoginIdKey}
        label={renderToString("AddUserScreen.user-info.label")}
      />
      <PasswordField
        textFieldClassName={styles.textField}
        disabled={!passwordRequired}
        label={renderToString("AddUserScreen.password.label")}
        value={password}
        onChange={onPasswordChange}
        passwordPolicy={passwordPolicy}
        errorMessage={localValidationErrorMessage ?? errorMessageMap.password}
      />
      <ButtonWithLoading
        type="submit"
        className={styles.addUserButton}
        loading={creatingUser}
        labelId="AddUserScreen.add-user.label"
        disabled={
          !isFormModified || selectedLoginIdKey == null || submittedForm
        }
      />
    </form>
  );
};

const AddUserScreen: React.FC = function AddUserScreen() {
  const { appID } = useParams();

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  const { effectiveAppConfig, loading, error, refetch } = useAppConfigQuery(
    appID
  );

  const [remountIdentifier, setRemountIdentifier] = useState(0);
  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.wrapper}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <AddUserContent
          key={remountIdentifier}
          appConfig={effectiveAppConfig}
          resetForm={resetForm}
        />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default AddUserScreen;
