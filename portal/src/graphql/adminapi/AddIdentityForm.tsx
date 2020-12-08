import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import PasswordField from "../../PasswordField";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { LoginIDKeyType, PortalAPIAppConfig } from "../../types";
import {
  AuthenticatorKind,
  AuthenticatorType,
} from "./__generated__/globalTypes";

import styles from "./AddIdentityForm.module.scss";
import { useSimpleForm } from "../../hook/useSimpleForm";
import FormContainer from "../../FormContainer";
import { ErrorParseRule } from "../../error/parse";
import { canCreateLoginIDIdentity } from "../../util/loginID";
import { Text } from "@fluentui/react";
import { UserQuery_node_User } from "./query/__generated__/UserQuery";
import { validatePassword } from "../../error/password";

interface FormState {
  loginID: string;
  password: string;
}

const defaultFormState: FormState = {
  loginID: "",
  password: "",
};

interface User {
  id: string;
  primaryAuthenticators: AuthenticatorType[];
}

function isPasswordRequired(
  config: PortalAPIAppConfig | null,
  user: User | null,
  loginIDType: LoginIDKeyType
) {
  let needPrimaryPassword: boolean;
  switch (loginIDType) {
    case "username":
      needPrimaryPassword =
        config?.authentication?.primary_authenticators?.includes("password") ??
        true;
      break;
    case "email":
    case "phone":
      needPrimaryPassword = false;
      break;
  }
  const hasPrimaryPassword =
    user?.primaryAuthenticators.includes(AuthenticatorType.PASSWORD) ?? false;
  return needPrimaryPassword && !hasPrimaryPassword;
}

interface LoginIDFieldProps {
  value: string;
  onChange: (value: string) => void;
}

interface AddIdentityFormProps {
  appConfig: PortalAPIAppConfig | null;
  rawUser: UserQuery_node_User | null;
  loginIDType: LoginIDKeyType;
  title: React.ReactNode;
  loginIDField: React.ComponentType<LoginIDFieldProps>;
  errorRules?: ErrorParseRule[];
  onReset?: () => void;
}

const AddIdentityForm: React.FC<AddIdentityFormProps> = function AddIdentityForm(
  props: AddIdentityFormProps
) {
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
        .filter((a) => a?.kind === AuthenticatorKind.PRIMARY)
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

  const rawForm = useSimpleForm(defaultFormState, submit, validate);
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
      navigate("..#connected-identities");
    }
  }, [form.isSubmitted, navigate]);

  const onLoginIDChange = useCallback(
    (value: string) => form.setState((state) => ({ ...state, loginID: value })),
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
      <div className={styles.root}>
        {title}
        <LoginIDField value={form.state.loginID} onChange={onLoginIDChange} />
        {requirePassword && (
          <PasswordField
            className={styles.password}
            textFieldClassName={styles.passwordField}
            passwordPolicy={passwordPolicy}
            label={renderToString("AddUsernameScreen.password.label")}
            value={form.state.password}
            onChange={onPasswordChange}
            parentJSONPointer="/"
            fieldName="password"
          />
        )}
      </div>
    </FormContainer>
  );
};

export default AddIdentityForm;
