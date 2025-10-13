import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
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
import FormContainer, { FormSaveButton } from "../../FormContainer";

import styles from "./AddUserScreen.module.css";
import { validatePassword } from "../../error/password";

enum PasswordCreationType {
  ManualEntry = "manual_entry",
  AutoGenerate = "auto_generate",
  NoPassword = "no_password",
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

function isPasswordFieldDisplayed(
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
    relatedAuthenticators.length > 0 &&
    relatedAuthenticators.includes("password")
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

const ManualEntryPasswordField: React.VFC<{
  disabled: boolean;
  passwordPolicy: PasswordPolicyConfig;
  password: string;
  onPasswordChange: (
    event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
    newValue?: string
  ) => void;
  selectedLoginIDType: LoginIDKeyType | null;
  sendPassword: boolean;
  onChangeSendPassword: (
    event?: React.FormEvent<HTMLElement | HTMLInputElement>,
    checked?: boolean
  ) => void;
}> = function ManualEntryPasswordField(props) {
  const {
    disabled,
    passwordPolicy,
    password,
    onPasswordChange,
    selectedLoginIDType,
    sendPassword,
    onChangeSendPassword,
  } = props;
  const { renderToString } = useContext(Context);

  return (
    <div>
      <PasswordField
        label={renderToString("AddUserScreen.password.label")}
        disabled={disabled}
        value={password}
        canRevealPassword={true}
        canGeneratePassword={true}
        onChange={onPasswordChange}
        passwordPolicy={passwordPolicy}
        parentJSONPointer=""
        fieldName="password"
      />
      {selectedLoginIDType === "email" ? (
        <Checkbox
          disabled={disabled}
          className={styles.checkbox}
          label={renderToString("AddUserScreen.send-password")}
          checked={sendPassword}
          onChange={onChangeSendPassword}
        />
      ) : null}
    </div>
  );
};

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

  const { username, email, phone, password, selectedLoginIDType } = state;

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

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  const renderManualEntryField = useCallback(
    (
      props?: IChoiceGroupOption & IChoiceGroupOptionProps,
      defaultRender?: (
        props?: IChoiceGroupOption & IChoiceGroupOptionProps
      ) => JSX.Element | null
    ) => {
      return (
        <>
          {defaultRender?.(props)}
          <div className={styles.choiceGroupOptionContent}>
            <Text block={true}>
              <FormattedMessage id="AddUserScreen.password-creation-type.manual.description" />
            </Text>
            <ManualEntryPasswordField
              disabled={!props?.checked}
              passwordPolicy={passwordPolicy}
              password={password}
              onPasswordChange={onPasswordChange}
              selectedLoginIDType={selectedLoginIDType}
              sendPassword={state.sendPassword}
              onChangeSendPassword={onChangeSendPassword}
            />
          </div>
        </>
      );
    },
    [
      passwordPolicy,
      password,
      onPasswordChange,
      selectedLoginIDType,
      state.sendPassword,
      onChangeSendPassword,
    ]
  );

  const renderAutoGenerateField = useCallback(
    (
      props?: IChoiceGroupOption & IChoiceGroupOptionProps,
      defaultRender?: (
        props?: IChoiceGroupOption & IChoiceGroupOptionProps
      ) => JSX.Element | null
    ) => {
      return (
        <>
          {defaultRender?.(props)}
          <div className={styles.choiceGroupOptionContent}>
            <Text block={true}>
              <FormattedMessage id="AddUserScreen.password-creation-type.auto-generate.description" />
            </Text>
          </div>
        </>
      );
    },
    []
  );

  const renderNoPasswordField = useCallback(
    (
      props?: IChoiceGroupOption & IChoiceGroupOptionProps,
      defaultRender?: (
        props?: IChoiceGroupOption & IChoiceGroupOptionProps
      ) => JSX.Element | null
    ) => {
      return (
        <>
          {defaultRender?.(props)}
          <div className={styles.choiceGroupOptionContent}>
            <Text block={true}>
              <FormattedMessage id="AddUserScreen.password-creation-type.no-password.description" />
            </Text>
          </div>
        </>
      );
    },
    []
  );

  const passwordCreateionTypeOptions = useMemo((): IChoiceGroupOption[] => {
    return [
      {
        key: PasswordCreationType.ManualEntry,
        text: renderToString("AddUserScreen.password-creation-type.manual"),
        onRenderField: renderManualEntryField,
        styles: {
          choiceFieldWrapper: { flex: "1 1 0px" },
          field: { fontWeight: "600" },
        },
      },
      {
        key: PasswordCreationType.AutoGenerate,
        text: renderToString("AddUserScreen.password-creation-type.auto"),
        onRenderField: renderAutoGenerateField,
        styles: {
          choiceFieldWrapper: { flex: "1 1 0px" },
          field: { fontWeight: "600" },
        },
      },
      {
        key: PasswordCreationType.NoPassword,
        text: renderToString(
          "AddUserScreen.password-creation-type.no-password"
        ),
        onRenderField: renderNoPasswordField,
        styles: {
          choiceFieldWrapper: { flex: "1 1 0px" },
          field: { fontWeight: "600" },
        },
      },
    ].filter((options) => {
      switch (selectedLoginIDType) {
        case "email":
          return true;
        case "phone":
        case "username":
          return [
            PasswordCreationType.ManualEntry,
            PasswordCreationType.NoPassword,
          ].includes(options.key);
        default:
          return false;
      }
    });
  }, [
    renderToString,
    renderManualEntryField,
    renderAutoGenerateField,
    renderNoPasswordField,
    selectedLoginIDType,
  ]);

  const onChangePasswordCreationType = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        setState((prev) => {
          const newPasswordCreationType = option.key as PasswordCreationType;
          if (prev.passwordCreationType === newPasswordCreationType) {
            return prev;
          }

          let newSendPassword = false;
          let newSetPasswordExpired = false;

          switch (newPasswordCreationType) {
            case PasswordCreationType.AutoGenerate:
              newSendPassword = true;
              break;
            case PasswordCreationType.NoPassword:
              newSendPassword = false;
              newSetPasswordExpired = false;
              break;
            case PasswordCreationType.ManualEntry:
              newSendPassword = prev.sendPassword;
              newSetPasswordExpired = prev.setPasswordExpired;
              break;
            default:
              break;
          }

          return {
            ...prev,
            password:
              newPasswordCreationType === PasswordCreationType.AutoGenerate ||
              newPasswordCreationType === PasswordCreationType.NoPassword
                ? ""
                : prev.password,
            passwordCreationType: newPasswordCreationType,
            sendPassword: newSendPassword,
            setPasswordExpired: newSetPasswordExpired,
          };
        });
      }
    },
    [setState]
  );

  const passwordFieldNeeded = useMemo(() => {
    return isPasswordFieldDisplayed(primaryAuthenticators, selectedLoginIDType);
  }, [primaryAuthenticators, selectedLoginIDType]);

  const onSelectLoginIdType = useCallback(
    (_event, options?: IChoiceGroupOption) => {
      const loginIdType = (options?.key ?? null) as LoginIDKeyType | null;
      if (!loginIdType || !loginIDKeyTypes.includes(loginIdType)) {
        return;
      }
      setState(() => ({
        ...makeDefaultFormState([...loginIDKeyTypes]),
        selectedLoginIDType: loginIdType,
      }));
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
            <ChoiceGroup
              className={styles.widget}
              styles={loginIDTypeOptionChoiceGroupStyle}
              selectedKey={selectedLoginIDType}
              options={loginIdTypeOptions}
              onChange={onSelectLoginIdType}
              label={renderToString(
                "AddUserScreen.select-sign-in-method.label"
              )}
            />

            {selectedLoginIDType ? (
              <div className={styles.identityOption}>
                <Label className={styles.identityOptionLabel}>
                  <FormattedMessage
                    id={loginIdTypeNameIds[selectedLoginIDType]}
                  />
                </Label>
                {textFieldRenderer[selectedLoginIDType]()}
              </div>
            ) : null}

            {passwordFieldNeeded ? (
              <>
                <ChoiceGroup
                  className={styles.widget}
                  selectedKey={state.passwordCreationType}
                  options={passwordCreateionTypeOptions}
                  onChange={onChangePasswordCreationType}
                  label={renderToString("AddUserScreen.password-setup.label")}
                />

                <div className={styles.widget}>
                  <Label className={styles.additionalOption}>
                    <FormattedMessage id="AddUserScreen.additional-option.label" />
                  </Label>
                  <Checkbox
                    className={styles.checkbox}
                    label={renderToString(
                      "AddUserScreen.force-change-on-login"
                    )}
                    checked={state.setPasswordExpired}
                    onChange={onChangeForceChangeOnLogin}
                    disabled={
                      state.passwordCreationType ===
                      PasswordCreationType.NoPassword
                    }
                  />
                </div>
              </>
            ) : null}
          </>
        )}
      </div>
      <div className={styles.widget}>
        <FormSaveButton
          saveButtonProps={{
            labelId: "AddUserScreen.add-user.label",
            iconProps: {
              iconName: "Add",
            },
          }}
        />
      </div>
    </ScreenContent>
  );
};

const AddUserScreen: React.VFC = function AddUserScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const {
    effectiveAppConfig,
    isLoading: loading,
    loadError: error,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

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
      if (
        !isPasswordFieldDisplayed(
          primaryAuthenticators,
          state.selectedLoginIDType
        ) ||
        state.passwordCreationType === PasswordCreationType.NoPassword
      ) {
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

      const hasPasswordField = isPasswordFieldDisplayed(
        primaryAuthenticators,
        state.selectedLoginIDType
      );
      const identityValue = state[loginIDType];
      let password: string | undefined;
      let sendPassword: boolean | undefined;
      let setPasswordExpired: boolean | undefined;
      if (hasPasswordField) {
        switch (state.passwordCreationType) {
          case PasswordCreationType.AutoGenerate:
            password = "";
            sendPassword = loginIDType === "email" ? true : undefined;
            setPasswordExpired = state.setPasswordExpired;
            break;
          case PasswordCreationType.ManualEntry:
            password = state.password;
            sendPassword =
              loginIDType === "email" ? state.sendPassword : undefined;
            setPasswordExpired = state.setPasswordExpired;
            break;
          case PasswordCreationType.NoPassword:
            password = undefined;
            sendPassword = false;
            setPasswordExpired = false;
            break;
          default:
            break;
        }
      } else {
        sendPassword = false;
        setPasswordExpired = false;
      }

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

  const canSave = useMemo(() => {
    if (form.state.selectedLoginIDType == null) {
      return false;
    }
    if (!form.state[form.state.selectedLoginIDType]) {
      return false;
    }

    if (
      isPasswordFieldDisplayed(
        primaryAuthenticators,
        form.state.selectedLoginIDType
      )
    ) {
      switch (form.state.passwordCreationType) {
        case PasswordCreationType.ManualEntry:
          return form.state.password.length > 0;
        case PasswordCreationType.AutoGenerate:
          return true;
        case PasswordCreationType.NoPassword:
          return true;
        default:
          throw new Error("unknown passwordCreationType");
      }
    }

    return true;
  }, [form.state, primaryAuthenticators]);

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
    <FormContainer form={form} canSave={canSave} hideFooterComponent={true}>
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
