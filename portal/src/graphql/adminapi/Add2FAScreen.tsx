import React, {
  createContext,
  useMemo,
  useState,
  useContext,
  useCallback,
  useEffect,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { PortalAPIAppConfig } from "../../types";
import FormPhoneTextField from "../../FormPhoneTextField";

import styles from "./Add2FAScreen.module.css";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { useCreateAuthenticatorMutation } from "./mutations/createAuthenticatorMutation";
import { AuthenticatorKind } from "./globalTypes.generated";
import FormTextField from "../../FormTextField";
import PasswordField from "../../PasswordField";

interface FieldContextValue {
  effectiveAppConfig?: PortalAPIAppConfig;
  resetToken?: unknown;
}

interface FormState {
  value: string;
}

const defaultState: FormState = {
  value: "",
};

const FieldContext = createContext<FieldContextValue>({});

function PhoneField(props: { onChange: (value: string) => void }) {
  const { effectiveAppConfig, resetToken } = useContext(FieldContext);
  const [inputValue, setInputValue] = useState("");
  const { onChange } = props;
  const onChangeValues = useCallback(
    (values: { e164?: string; rawInputValue: string }) => {
      const { e164, rawInputValue } = values;
      onChange(e164 ?? "");
      setInputValue(rawInputValue);
    },
    [onChange]
  );
  useEffect(() => {
    setInputValue("");
  }, [resetToken]);
  return (
    <FormPhoneTextField
      parentJSONPointer=""
      fieldName="phone"
      className={styles.widget}
      allowlist={effectiveAppConfig?.ui?.phone_input?.allowlist}
      pinnedList={effectiveAppConfig?.ui?.phone_input?.pinned_list}
      inputValue={inputValue}
      onChange={onChangeValues}
    />
  );
}

function EmailField(props: {
  value: string;
  onChange: (value: string) => void;
}) {
  const { value, onChange } = props;
  const { renderToString } = useContext(MessageContext);
  const onEmailChange = useCallback(
    (_, value?: string) => onChange(value ?? ""),
    [onChange]
  );

  return (
    <FormTextField
      className={styles.widget}
      parentJSONPointer=""
      fieldName="email"
      label={renderToString("EmailScreen.email.label")}
      value={value}
      onChange={onEmailChange}
    />
  );
}

function PaswordField(props: {
  value: string;
  onChange: (value: string) => void;
}) {
  const { value, onChange } = props;
  const { renderToString } = useContext(MessageContext);
  const { effectiveAppConfig } = useContext(FieldContext);

  const onFieldChange = useCallback(
    (e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      onChange(e.currentTarget.value);
    },
    [onChange]
  );

  const passwordPolicy = useMemo(() => {
    return effectiveAppConfig?.authenticator?.password?.policy ?? {};
  }, [effectiveAppConfig]);

  return (
    <PasswordField
      className={styles.widget}
      passwordPolicy={passwordPolicy}
      label={renderToString("UsernameScreen.password.label")}
      value={value}
      onChange={onFieldChange}
      parentJSONPointer=""
      fieldName="password"
    />
  );
}

interface Add2FAScreenProps {
  authenticatorType: "oob_otp_email" | "oob_otp_sms" | "password";
}

const Add2FAScreen: React.VFC<Add2FAScreenProps> = function Add2FAScreen({
  authenticatorType,
}) {
  const { appID, userID } = useParams() as {
    appID: string;
    userID: string;
  };
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
  } = useAppAndSecretConfigQuery(appID);

  const navigate = useNavigate();

  const { createAuthenticator } = useCreateAuthenticatorMutation(
    user?.id ?? ""
  );

  const navBreadcrumbItems = useMemo(() => {
    const titleId = {
      oob_otp_email: "Add2FAScreen.title.email",
      oob_otp_sms: "Add2FAScreen.title.phone",
      password: "Add2FAScreen.title.password",
    }[authenticatorType];
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      {
        to: `~/users/${user?.id}/details`,
        label: <FormattedMessage id="UserDetailsScreen.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id={titleId} />,
      },
    ];
  }, [authenticatorType, user?.id]);

  const [resetToken, setResetToken] = useState({});

  const onReset = useCallback(() => {
    setResetToken({});
  }, []);

  const contextValue = useMemo(() => {
    return {
      effectiveAppConfig: effectiveAppConfig ?? undefined,
      resetToken,
    };
  }, [resetToken, effectiveAppConfig]);

  const title = (
    <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
  );

  const validate = useCallback((_state: FormState) => {
    return null;
  }, []);

  const submit = useCallback(
    async (state: FormState) => {
      switch (authenticatorType) {
        case "oob_otp_sms":
          await createAuthenticator({
            type: "oob_otp_sms",
            phone: state.value,
            kind: AuthenticatorKind.Secondary,
          });
          break;
        case "oob_otp_email":
          await createAuthenticator({
            type: "oob_otp_email",
            email: state.value,
            kind: AuthenticatorKind.Secondary,
          });
          break;
        case "password":
          await createAuthenticator({
            type: "password",
            password: state.value,
            kind: AuthenticatorKind.Secondary,
          });
          break;
        default:
          throw new Error("unknown authenticator type");
      }
      await refetchUser();
    },
    [authenticatorType, createAuthenticator, refetchUser]
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
        onReset();
      },
    }),
    [rawForm, onReset]
  );

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("./..#account-security");
    }
  }, [form.isSubmitted, navigate]);

  const onValueChange = useCallback(
    (value: string) => form.setState((state) => ({ ...state, value: value })),
    [form]
  );

  const canSave = form.state.value.length > 0;

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
    <FieldContext.Provider value={contextValue}>
      <FormContainer form={form} canSave={canSave}>
        <ScreenContent>
          {title}
          {authenticatorType === "oob_otp_sms" ? (
            <PhoneField onChange={onValueChange} />
          ) : authenticatorType === "oob_otp_email" ? (
            <EmailField value={form.state.value} onChange={onValueChange} />
          ) : (
            <PaswordField value={form.state.value} onChange={onValueChange} />
          )}
        </ScreenContent>
      </FormContainer>
    </FieldContext.Provider>
  );
};

export default Add2FAScreen;
