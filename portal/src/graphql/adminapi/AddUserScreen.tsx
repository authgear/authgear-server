import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { ChoiceGroup, IChoiceGroupOption, Label, Text } from "@fluentui/react";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useCreateUserMutation } from "./mutations/createUserMutation";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import PasswordField from "../../PasswordField";
import { useTextField } from "../../hook/useInput";
import {
  PrimaryAuthenticatorType,
  LoginIDKeyType,
  loginIDKeyTypes,
  PasswordPolicyConfig,
  PortalAPIAppConfig,
} from "../../types";
import { ErrorParseRule } from "../../error/parse";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormTextField from "../../FormTextField";
import FormContainer from "../../FormContainer";

import styles from "./AddUserScreen.module.scss";
import { validatePassword } from "../../error/password";

interface FormState {
  selectedLoginIDType: LoginIDKeyType | null;
  username: string;
  email: string;
  phone: string;
  password: string;
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
    };
  }

  return {
    selectedLoginIDType: null,
    username: "",
    email: "",
    phone: "",
    password: "",
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

  switch (loginIdKeySelected) {
    case "email":
      return !primaryAuthenticators.includes("oob_otp_email");
    case "phone":
      return !primaryAuthenticators.includes("oob_otp_sms");
    case "username":
      return true;
    default:
      return false;
  }
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
  {
    reason: "ValidationFailed",
    location: "",
    kind: "format",
    errorMessageID: "AddUserScreen.error.invalid-identity",
  },
  {
    reason: "InvariantViolated",
    kind: "DuplicatedIdentity",
    errorMessageID: "AddUserScreen.error.duplicated-identity",
  },
];

interface AddUserContentProps {
  primaryAuthenticators: PrimaryAuthenticatorType[];
  passwordPolicy: PasswordPolicyConfig;
  loginIDTypes: LoginIDKeyType[];
  form: SimpleFormModel<FormState>;
}

const AddUserContent: React.FC<AddUserContentProps> = function AddUserContent(
  props: AddUserContentProps
) {
  const {
    primaryAuthenticators,
    passwordPolicy,
    loginIDTypes,
    form: { state, setState },
  } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  const { username, email, phone, password, selectedLoginIDType } = state;

  const passwordFieldDisabled = useMemo(() => {
    return !isPasswordNeeded(primaryAuthenticators, selectedLoginIDType);
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
      const renderTextField =
        selectedLoginIDType === loginIdType
          ? textFieldRenderer[loginIdType]
          : undefined;
      return {
        key: loginIdType,
        text: renderToString(messageId),
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderLabel: (option) => {
          return option ? (
            <div className={styles.identityOption}>
              <Label className={styles.identityOptionLabel}>
                {option.text}
              </Label>
              {renderTextField?.()}
            </div>
          ) : null;
        },
      };
    });
  }, [loginIDTypes, renderToString, textFieldRenderer, selectedLoginIDType]);

  // NOTE: cannot add user identity if none of three field is available
  const canAddUser = loginIdTypeOptions.length > 0;

  // TODO: improve empty state
  if (!canAddUser) {
    return (
      <Text>
        <FormattedMessage id="AddUserScreen.cannot-add-user" />
      </Text>
    );
  }

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <ChoiceGroup
        className={styles.userInfo}
        styles={{ label: { marginBottom: "15px", fontSize: "14px" } }}
        selectedKey={selectedLoginIDType ?? undefined}
        options={loginIdTypeOptions}
        onChange={onSelectLoginIdType}
        label={renderToString("AddUserScreen.user-info.label")}
      />
      <PasswordField
        textFieldClassName={styles.textField}
        disabled={passwordFieldDisabled}
        label={renderToString("AddUserScreen.password.label")}
        value={password}
        onChange={onPasswordChange}
        passwordPolicy={passwordPolicy}
        parentJSONPointer=""
        fieldName="password"
      />
    </div>
  );
};

const AddUserScreen: React.FC = function AddUserScreen() {
  const { appID } = useParams();
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

  const defaultFormState = useMemo(
    () => makeDefaultFormState(loginIDTypes),
    [loginIDTypes]
  );

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
      const password = needPassword ? state.password : undefined;

      await createUser({ key: loginIDType, value: identityValue }, password);
    },
    [createUser, primaryAuthenticators]
  );

  const form = useSimpleForm(defaultFormState, submit, validate);

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
      navigate("..");
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
      />
    </FormContainer>
  );
};

export default AddUserScreen;
