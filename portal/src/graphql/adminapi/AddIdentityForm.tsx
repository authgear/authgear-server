import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import PasswordField from "../../PasswordField";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { LoginIDKeyType, PortalAPIAppConfig } from "../../types";
import { AuthenticatorKind, AuthenticatorType } from "./globalTypes.generated";

import styles from "./AddIdentityForm.module.css";
import { useSimpleForm } from "../../hook/useSimpleForm";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import { ErrorParseRule } from "../../error/parse";
import { canCreateLoginIDIdentity } from "../../util/loginID";
import { Text } from "@fluentui/react";
import { UserQueryNodeFragment } from "./query/userQuery.generated";
import { validatePassword } from "../../error/password";

interface FormState {
  loginID: string;
  password: string;
}

const defaultState: FormState = {
  loginID: "",
  password: "",
};

interface User {
  id: string;
  primaryAuthenticators: AuthenticatorType[];
}

// eslint-disable-next-line complexity
function isPasswordRequired(
  config: PortalAPIAppConfig | null,
  user: User | null,
  loginIDType: LoginIDKeyType
) {
  let needPrimaryPassword: boolean;
  const isPasswordEnabled =
    config?.authentication?.primary_authenticators?.includes("password") ??
    true;
  let isOOBOTPEmailFirst = false;
  let isOOBOTPSMSFirst = false;
  const primaryAuthenticators =
    config?.authentication?.primary_authenticators ?? [];
  // reverse order is important
  for (let i = primaryAuthenticators.length - 1; i >= 0; i--) {
    switch (primaryAuthenticators[i]) {
      case "oob_otp_email":
        isOOBOTPEmailFirst = true;
        break;
      case "oob_otp_sms":
        isOOBOTPSMSFirst = true;
        break;
      case "password":
        isOOBOTPEmailFirst = false;
        isOOBOTPSMSFirst = false;
        break;
      default:
        break;
    }
  }

  switch (loginIDType) {
    case "username":
      needPrimaryPassword = isPasswordEnabled;
      break;
    case "email":
      needPrimaryPassword = isPasswordEnabled && !isOOBOTPEmailFirst;
      break;
    case "phone":
      needPrimaryPassword = isPasswordEnabled && !isOOBOTPSMSFirst;
      break;
  }
  const hasPrimaryPassword =
    user?.primaryAuthenticators.includes(AuthenticatorType.Password) ?? false;
  return needPrimaryPassword && !hasPrimaryPassword;
}

export interface LoginIDFieldProps {
  value: string;
  onChange: (value: string) => void;
}

interface AddIdentityFormProps {
  appConfig: PortalAPIAppConfig | null;
  rawUser: UserQueryNodeFragment | null;
  loginIDType: LoginIDKeyType;
  title: React.ReactNode;
  loginIDField: React.ComponentType<LoginIDFieldProps>;
  errorRules?: ErrorParseRule[];
  onReset?: () => void;
}

const AddIdentityForm: React.VFC<AddIdentityFormProps> =
  function AddIdentityForm(props: AddIdentityFormProps) {
    const {
      appConfig,
      rawUser,
      loginIDType,
      title,
      loginIDField: LoginIDField,
      onReset,
    } = props;

    const navigate = useNavigate();
    const { renderToString } = useContext(Context);

    const user: User = useMemo(() => {
      if (!rawUser) {
        return { id: "", primaryAuthenticators: [] };
      }
      const authenticators =
        rawUser.authenticators?.edges?.map((e) => e?.node) ?? [];
      return {
        id: rawUser.id,
        primaryAuthenticators: authenticators
          .filter((a) => a?.kind === AuthenticatorKind.Primary)
          .map((a) => a!.type),
      };
    }, [rawUser]);

    const { createIdentity } = useCreateLoginIDIdentityMutation(user.id);

    const requirePassword = useMemo(() => {
      return isPasswordRequired(appConfig, user, loginIDType);
    }, [appConfig, user, loginIDType]);

    const passwordPolicy = useMemo(() => {
      return appConfig?.authenticator?.password?.policy ?? {};
    }, [appConfig]);

    const validate = useCallback(
      (state: FormState) => {
        if (!requirePassword) {
          return null;
        }
        return validatePassword(state.password, passwordPolicy);
      },
      [requirePassword, passwordPolicy]
    );

    const submit = useCallback(
      async (state: FormState) => {
        const password = requirePassword ? state.password : undefined;
        await createIdentity(
          { key: loginIDType, value: state.loginID },
          password
        );
      },
      [loginIDType, requirePassword, createIdentity]
    );

    const rawForm = useSimpleForm({
      stateMode:
        "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
      defaultState,
      submit,
      validate,
    });
    const form = useMemo(
      () => ({
        ...rawForm,
        reset: () => {
          rawForm.reset();
          onReset?.();
        },
      }),
      [rawForm, onReset]
    );

    useEffect(() => {
      if (form.isSubmitted) {
        navigate("./..#connected-identities");
      }
    }, [form.isSubmitted, navigate]);

    const onLoginIDChange = useCallback(
      (value: string) =>
        form.setState((state) => ({ ...state, loginID: value })),
      [form]
    );
    const onPasswordChange = useCallback(
      (_, value?: string) =>
        form.setState((state) => ({ ...state, password: value ?? "" })),
      [form]
    );

    const canSave =
      form.state.loginID.length > 0 &&
      (!requirePassword || form.state.password.length > 0);

    if (!canCreateLoginIDIdentity(appConfig)) {
      return (
        <Text className={styles.helpText}>
          <FormattedMessage id="CreateIdentity.require-login-id" />
        </Text>
      );
    }

    return (
      <FormContainer form={form} canSave={canSave}>
        <ScreenContent>
          {title}
          <LoginIDField value={form.state.loginID} onChange={onLoginIDChange} />
          {requirePassword ? (
            <PasswordField
              className={styles.widget}
              passwordPolicy={passwordPolicy}
              label={renderToString("AddUsernameScreen.password.label")}
              value={form.state.password}
              onChange={onPasswordChange}
              parentJSONPointer=""
              fieldName="password"
            />
          ) : null}
        </ScreenContent>
      </FormContainer>
    );
  };

export default AddIdentityForm;
