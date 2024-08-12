import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupStyleProps,
  IChoiceGroupStyles,
  IStyleFunctionOrObject,
  Label,
  MessageBar,
  Text,
} from "@fluentui/react";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useCreateUserMutation } from "./mutations/createUserMutation";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ScreenContent from "../../ScreenContent";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import PasswordField from "../../PasswordField";
import { useCheckbox, useTextField } from "../../hook/useInput";
import {
  PrimaryAuthenticatorType,
  LoginIDKeyType,
  loginIDKeyTypes,
  PasswordPolicyConfig,
  PortalAPIAppConfig,
} from "../../types";
import {
  ErrorParseRule,
  makeInvariantViolatedErrorParseRule,
} from "../../error/parse";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormTextField from "../../FormTextField";
import FormContainer from "../../FormContainer";

import styles from "./AddUserScreen.module.css";
import { validatePassword } from "../../error/password";

enum PasswordCreationType {
  ManualEntry = "manual_entry",
  AutoGenerate = "auto_generate",
}

interface FormState {
  selectedLoginIDType: LoginIDKeyType | null;
  username: string;
  email: string;
  phone: string;
  password: string;
  passwordCreationType: PasswordCreationType;
  sendPassword: boolean;
  setPasswordExpired: boolean;
}

const loginIdTypeNameIds: Record<LoginIDKeyType, string> = {
  username: "login-id-key.username",
  email: "login-id-key.email",
  phone: "login-id-key.phone",
};

function makeDefaultFormState(loginIDTypes: LoginIDKeyType[]): FormState {
  if (loginIDTypes.length === 1) {
    return {
      selectedLoginIDType: loginIDTypes[0],
      username: "",
      email: "",
      phone: "",
      password: "",
      passwordCreationType: PasswordCreationType.ManualEntry,
      sendPassword: false,
      setPasswordExpired: true,
    };
  }

  return {
    selectedLoginIDType: null,
    username: "",
    email: "",
    phone: "",
    password: "",
    passwordCreationType: PasswordCreationType.ManualEntry,
    sendPassword: false,
    setPasswordExpired: true,
  };
}

function isPasswordNeeded(
  primaryAuthenticators: PrimaryAuthenticatorType[],
  loginIdKeySelected: LoginIDKeyType | null
) {
  // Unknown yet.
  if (loginIdKeySelected == null) {
    return false;
  }

  const filterAuthenticators = (allowedTypes: PrimaryAuthenticatorType[]) => {
    return primaryAuthenticators.filter((authenticator) =>
      allowedTypes.includes(authenticator)
    );
  };

  let relatedAuthenticators: PrimaryAuthenticatorType[];
  switch (loginIdKeySelected) {
    case "email":
      relatedAuthenticators = filterAuthenticators([
        "oob_otp_email",
        "password",
      ]);
      break;
    case "phone":
      relatedAuthenticators = filterAuthenticators(["oob_otp_sms", "password"]);
      break;
    case "username":
      relatedAuthenticators = filterAuthenticators(["password"]);
      break;
    default:
      relatedAuthenticators = filterAuthenticators([]);
  }

  return (
    relatedAuthenticators.length > 0 && relatedAuthenticators[0] === "password"
  );
}

function getEnabledLoginIDTypes(
  appConfig: PortalAPIAppConfig | null
): LoginIDKeyType[] {
  const enabledIdentities = appConfig?.authentication?.identities ?? [];
  const isLoginIDIdentityEnabled = enabledIdentities.includes("login_id");
  // if login ID identity is disabled
  // we cannot add login ID identity to new user
  if (!isLoginIDIdentityEnabled) {
    return [];
  }

  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const loginIdKeyOptions = new Set<LoginIDKeyType>();
  for (const key of loginIdKeys) {
    switch (key.type) {
      case "username": {
        loginIdKeyOptions.add("username");
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

const errorRules: ErrorParseRule[] = [
  makeInvariantViolatedErrorParseRule(
    "DuplicatedIdentity",
    "AddUserScreen.error.duplicated-identity"
  ),
];

interface AddUserContentProps {
  primaryAuthenticators: PrimaryAuthenticatorType[];
  passwordPolicy: PasswordPolicyConfig;
  loginIDTypes: LoginIDKeyType[];
  form: SimpleFormModel<FormState>;
  isPasskeyOnly: boolean;
}

// eslint-disable-next-line complexity
const AddUserContent: React.VFC<AddUserContentProps> = function AddUserContent(
  props: AddUserContentProps
) {
  const {
    primaryAuthenticators,
    passwordPolicy,
    loginIDTypes,
    form: { state, setState },
    isPasskeyOnly,
  } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  const passwordCreateionTypeOptions = useMemo(() => {
    return [
      {
        key: PasswordCreationType.ManualEntry,
        text: renderToString("AddUserScreen.password-creation-type.manual"),
      },
      {
        key: PasswordCreationType.AutoGenerate,
        text: renderToString("AddUserScreen.password-creation-type.auto"),
      },
    ];
  }, [renderToString]);

  const onChangePasswordCreationType = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        setState((prev) => ({
          ...prev,
          password:
            option.key === PasswordCreationType.AutoGenerate
              ? ""
              : prev.password,
          passwordCreationType: option.key as PasswordCreationType,
          sendPassword:
            prev.sendPassword ||
            option.key === PasswordCreationType.AutoGenerate,
        }));
      }
    },
    [setState]
  );

  const { username, email, phone, password, selectedLoginIDType } = state;

  const passwordFieldNeeded = useMemo(() => {
    return isPasswordNeeded(primaryAuthenticators, selectedLoginIDType);
  }, [primaryAuthenticators, selectedLoginIDType]);

  const { onChange: onUsernameChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, username: value }));
  });
  const { onChange: onEmailChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, email: value }));
  });
  const { onChange: onPhoneChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, phone: value }));
  });
  const { onChange: onPasswordChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, password: value }));
  });
  const { onChange: onChangeSendPassword } = useCheckbox((value) => {
    setState((prev) => ({ ...prev, sendPassword: value }));
  });
  const { onChange: onChangeForceChangeOnLogin } = useCheckbox((value) => {
    setState((prev) => ({ ...prev, setPasswordExpired: value }));
  });

  const onSelectLoginIdType = useCallback(
    (_event, options?: IChoiceGroupOption) => {
      const loginIdType = (options?.key ?? null) as LoginIDKeyType | null;
      if (!loginIdType || !loginIDKeyTypes.includes(loginIdType)) {
        return;
      }
      setState((prev) => ({ ...prev, selectedLoginIDType: loginIdType }));
    },
    [setState]
  );
  const renderUsernameField = useCallback(() => {
    return (
      <FormTextField
        className={styles.textField}
        value={username}
        onChange={onUsernameChange}
        parentJSONPointer=""
        fieldName="username"
        errorRules={errorRules}
      />
    );
  }, [username, onUsernameChange]);

  const renderEmailField = useCallback(() => {
    return (
      <FormTextField
        className={styles.textField}
        value={email}
        onChange={onEmailChange}
        parentJSONPointer=""
        fieldName="email"
        errorRules={errorRules}
      />
    );
  }, [email, onEmailChange]);

  const renderPhoneField = useCallback(() => {
    return (
      <FormTextField
        className={styles.textField}
        value={phone}
        onChange={onPhoneChange}
        parentJSONPointer=""
        fieldName="phone"
        errorRules={errorRules}
      />
    );
  }, [phone, onPhoneChange]);

  const textFieldRenderer: Record<LoginIDKeyType, () => React.ReactNode> =
    useMemo(
      () => ({
        username: renderUsernameField,
        email: renderEmailField,
        phone: renderPhoneField,
      }),
      [renderUsernameField, renderEmailField, renderPhoneField]
    );

  const loginIdTypeOptions: IChoiceGroupOption[] = useMemo(() => {
    return loginIDTypes.map((loginIdType) => {
      const messageId = loginIdTypeNameIds[loginIdType];
      return {
        key: loginIdType,
        text: renderToString(messageId),
      };
    });
  }, [loginIDTypes, renderToString]);

  // NOTE: cannot add user identity if none of three field is available
  const canAddUser = loginIdTypeOptions.length > 0;

  const loginIDTypeOptionChoiceGroupStyle = useMemo<
    IStyleFunctionOrObject<IChoiceGroupStyleProps, IChoiceGroupStyles>
  >(() => {
    return {
      flexContainer: { display: "flex", gap: "16px" },
      label: { fontSize: "14px" },
    };
  }, []);

  // TODO: improve empty state
  if (!canAddUser) {
    return (
      <Text>
        <FormattedMessage id="AddUserScreen.cannot-add-user" />
      </Text>
    );
  }

  return (
    <ScreenContent>
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
      <div className={styles.verticalForm}>
        {isPasskeyOnly ? (
          <div className={styles.widget}>
            <MessageBar>
              <FormattedMessage id="AddUserScreen.passkey-only.message" />
            </MessageBar>
          </div>
        ) : (
          <>
            {loginIdTypeOptions.length > 1 ? (
              <ChoiceGroup
                className={styles.widget}
                styles={loginIDTypeOptionChoiceGroupStyle}
                selectedKey={selectedLoginIDType}
                options={loginIdTypeOptions}
                onChange={onSelectLoginIdType}
                label={renderToString("AddUserScreen.user-info.label")}
              />
            ) : null}

            {selectedLoginIDType ? (
              <div className={styles.identityOption}>
                <Label className={styles.identityOptionLabel}>
                  <FormattedMessage
                    id={loginIdTypeNameIds[selectedLoginIDType]}
                  />
                </Label>
                {textFieldRenderer[selectedLoginIDType]()}
                {passwordFieldNeeded && selectedLoginIDType === "email" ? (
                  <ChoiceGroup
                    className={styles.widget}
                    selectedKey={state.passwordCreationType}
                    options={passwordCreateionTypeOptions}
                    onChange={onChangePasswordCreationType}
                  />
                ) : null}
              </div>
            ) : null}
            <div className={styles.widget}>
              <PasswordField
                label={renderToString("AddUserScreen.password.label")}
                value={password}
                canRevealPassword={true}
                canGeneratePassword={true}
                onChange={onPasswordChange}
                passwordPolicy={passwordPolicy}
                parentJSONPointer=""
                fieldName="password"
                disabled={
                  !passwordFieldNeeded ||
                  state.passwordCreationType ===
                    PasswordCreationType.AutoGenerate
                }
              />
              {passwordFieldNeeded && selectedLoginIDType === "email" ? (
                <Checkbox
                  className={styles.checkbox}
                  label={renderToString("AddUserScreen.send-password")}
                  checked={state.sendPassword}
                  onChange={onChangeSendPassword}
                  disabled={
                    state.passwordCreationType ===
                    PasswordCreationType.AutoGenerate
                  }
                />
              ) : null}
              {passwordFieldNeeded ? (
                <Checkbox
                  className={styles.checkbox}
                  label={renderToString("AddUserScreen.force-change-on-login")}
                  checked={state.setPasswordExpired}
                  onChange={onChangeForceChangeOnLogin}
                />
              ) : null}
            </div>
          </>
        )}
      </div>
    </ScreenContent>
  );
};

const AddUserScreen: React.VFC = function AddUserScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const { effectiveAppConfig, loading, error, refetch } =
    useAppAndSecretConfigQuery(appID);

  const primaryAuthenticators = useMemo(
    () => effectiveAppConfig?.authentication?.primary_authenticators ?? [],
    [effectiveAppConfig]
  );

  const loginIDTypes = useMemo(
    () => getEnabledLoginIDTypes(effectiveAppConfig),
    [effectiveAppConfig]
  );

  const passwordPolicy = useMemo(
    () => effectiveAppConfig?.authenticator?.password?.policy ?? {},
    [effectiveAppConfig]
  );

  const isPasskeyOnly = useMemo(() => {
    const primaryAuthenticators =
      effectiveAppConfig?.authentication?.primary_authenticators ?? [];
    return (
      primaryAuthenticators.length === 1 &&
      primaryAuthenticators[0] === "passkey"
    );
  }, [effectiveAppConfig]);

  const defaultState = useMemo(() => {
    return makeDefaultFormState(loginIDTypes);
  }, [loginIDTypes]);

  const { createUser } = useCreateUserMutation();

  const validate = useCallback(
    (state: FormState) => {
      if (!isPasswordNeeded(primaryAuthenticators, state.selectedLoginIDType)) {
        return null;
      }
      return validatePassword(state.password, passwordPolicy);
    },
    [primaryAuthenticators, passwordPolicy]
  );

  const submit = useCallback(
    async (state: FormState) => {
      const loginIDType = state.selectedLoginIDType;
      if (!loginIDType) {
        return;
      }

      const needPassword = isPasswordNeeded(
        primaryAuthenticators,
        state.selectedLoginIDType
      );
      const identityValue = state[loginIDType];
      const password =
        needPassword &&
        state.passwordCreationType === PasswordCreationType.ManualEntry
          ? state.password
          : undefined;
      const { sendPassword, setPasswordExpired } = state;

      await createUser({
        identity: { key: loginIDType, value: identityValue },
        password,
        sendPassword,
        setPasswordExpired,
      });
    },
    [createUser, primaryAuthenticators]
  );

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
  });

  const canSave =
    form.state.selectedLoginIDType != null &&
    form.state[form.state.selectedLoginIDType].length > 0;

  const saveButtonProps = useMemo(
    () => ({
      labelId: "AddUserScreen.add-user.label",
      iconName: "Add",
    }),
    []
  );

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("./..");
    }
  }, [form.isSubmitted, navigate]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContainer
      form={form}
      canSave={canSave}
      saveButtonProps={saveButtonProps}
    >
      <AddUserContent
        form={form}
        primaryAuthenticators={primaryAuthenticators}
        loginIDTypes={loginIDTypes}
        passwordPolicy={passwordPolicy}
        isPasskeyOnly={isPasskeyOnly}
      />
    </FormContainer>
  );
};

export default AddUserScreen;
