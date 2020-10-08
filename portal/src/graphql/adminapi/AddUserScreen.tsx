import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import zxcvbn from "zxcvbn";
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
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import PasswordField, {
  extractGuessableLevel,
  isPasswordValid,
} from "../../PasswordField";
import { useTextField } from "../../hook/useInput";
import { PasswordPolicyConfig, PortalAPIAppConfig } from "../../types";
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";
import { parseError } from "../../util/error";

import styles from "./AddUserScreen.module.scss";

type LoginIDKey = "username" | "email" | "phone";
function isLoginIDKey(value?: string): value is LoginIDKey {
  return ["username", "email", "phone"].includes(value ?? "");
}

interface AddUserContentProps {
  appConfig: PortalAPIAppConfig | null;
}

interface AddUserScreenState {
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

function validate(
  screenState: AddUserScreenState,
  passwordPolicy: PasswordPolicyConfig,
  passwordRequired: boolean
) {
  const errors: Violation[] = [];
  if (passwordRequired) {
    const passwordCheckResult = zxcvbn(screenState.password);
    const guessableLevel = extractGuessableLevel(passwordCheckResult);
    if (
      !isPasswordValid(passwordPolicy, screenState.password, guessableLevel)
    ) {
      errors.push({ kind: "Invalid" });
    }
  }
  return errors;
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
  const { appConfig } = props;
  const { renderToString } = useContext(Context);

  const navigate = useNavigate();
  const {
    createUser,
    loading: creatingUser,
    error: createUserError,
  } = useCreateUserMutation();

  const selectedLoginIdInLastSubmission = useRef<LoginIDKey | null>(null);
  const [submittedForm, setSubmittedForm] = useState(false);
  const [selectedLoginIdKey, setSelectedLoginIdKey] = useState<
    LoginIDKey | undefined
  >(undefined);
  const { value: username, onChange: onUsernameChange } = useTextField("");
  const { value: email, onChange: onEmailChange } = useTextField("");
  const { value: phone, onChange: onPhoneChange } = useTextField("");
  const { value: password, onChange: onPasswordChange } = useTextField("");

  const onSelectLoginIdKey = useCallback(
    (_event, options?: IChoiceGroupOption) => {
      const loginIdKey = options?.key;
      if (!isLoginIDKey(loginIdKey)) {
        return;
      }
      setSelectedLoginIdKey(loginIdKey);
    },
    []
  );

  const [localViolations, setLocalViolations] = useState<Violation[]>([]);

  const { errorMessages, unhandledViolations } = useMemo(() => {
    const violations =
      localViolations.length === 0
        ? parseError(createUserError)
        : localViolations;

    const lastSumissionLoginId = selectedLoginIdInLastSubmission.current;
    const lastSumissionLoginIdText =
      lastSumissionLoginId != null
        ? renderToString(loginIdLocaleKey[lastSumissionLoginId])
        : "";
    const selectedIdentityFieldErrorMessage: string[] = [];
    const passwordFieldErrorMessages: string[] = [];
    const unhandledViolations: Violation[] = [];
    for (const violation of violations) {
      if (violation.kind === "format") {
        const violationDetail = violation.detail;
        if (isLoginIDKey(violationDetail)) {
          selectedIdentityFieldErrorMessage.push(
            renderToString("AddUserScreen.error.invalid-identity", {
              loginIdType: lastSumissionLoginIdText,
            })
          );
        }
      } else if (violation.kind === "DuplicatedIdentity") {
        selectedIdentityFieldErrorMessage.push(
          renderToString("AddUserScreen.error.duplicated-identity", {
            loginIdType: lastSumissionLoginIdText,
          })
        );
      } else if (violation.kind === "Invalid") {
        passwordFieldErrorMessages.push(
          renderToString("AddUserScreen.error.invalid-password")
        );
      } else {
        unhandledViolations.push(violation);
      }
    }

    const errorMessages: Partial<Record<LoginIDKey | "password", string>> = {
      password: defaultFormatErrorMessageList(passwordFieldErrorMessages),
    };

    if (lastSumissionLoginId != null) {
      errorMessages[lastSumissionLoginId] = defaultFormatErrorMessageList(
        selectedIdentityFieldErrorMessage
      );
    }

    return {
      errorMessages,
      unhandledViolations,
    };
  }, [renderToString, localViolations, createUserError]);

  const renderUsernameField = useCallback(() => {
    return (
      <TextField
        className={styles.textField}
        value={username}
        onChange={onUsernameChange}
        errorMessage={errorMessages.username}
      />
    );
  }, [username, onUsernameChange, errorMessages.username]);

  const renderEmailField = useCallback(() => {
    return (
      <TextField
        className={styles.textField}
        value={email}
        onChange={onEmailChange}
        errorMessage={errorMessages.email}
      />
    );
  }, [email, onEmailChange, errorMessages.email]);

  const renderPhoneField = useCallback(() => {
    return (
      <TextField
        className={styles.textField}
        value={phone}
        onChange={onPhoneChange}
        errorMessage={errorMessages.phone}
      />
    );
  }, [phone, onPhoneChange, errorMessages.phone]);

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

  const isFormModified = useMemo(() => {
    const initialState: AddUserScreenState = {
      username: "",
      email: "",
      phone: "",
      password: "",
    };
    return !deepEqual(initialState, screenState);
  }, [screenState]);

  const onClickAddUser = useCallback(() => {
    const selectedKey = screenState.selectedLoginIdKey;
    const localValidationErrors = validate(
      screenState,
      passwordPolicy,
      passwordRequired
    );
    if (localValidationErrors.length > 0 || selectedKey == null) {
      setLocalViolations(localValidationErrors);
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
  }, [screenState, passwordPolicy, passwordRequired, createUser]);

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
    <section className={styles.content}>
      {unhandledViolations.length > 0 && <ShowError error={createUserError} />}
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
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
        errorMessage={errorMessages.password}
      />
      <ButtonWithLoading
        className={styles.addUserButton}
        loading={creatingUser}
        labelId="AddUserScreen.add-user.label"
        disabled={!isFormModified || selectedLoginIdKey == null}
        onClick={onClickAddUser}
      />
    </section>
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

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <AddUserContent appConfig={effectiveAppConfig} />
    </main>
  );
};

export default AddUserScreen;
