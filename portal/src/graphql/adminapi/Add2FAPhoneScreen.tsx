import React, {
  createContext,
  useMemo,
  useState,
  useContext,
  useCallback,
  useEffect,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { PortalAPIAppConfig } from "../../types";
import FormPhoneTextField from "../../FormPhoneTextField";

import styles from "./PhoneScreen.module.css";
import FormContainer from "../../FormContainer";
import ScreenContent from "../../ScreenContent";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { useCreateAuthenticatorMutation } from "./mutations/createAuthenticatorMutation";
import { AuthenticatorKind } from "./globalTypes.generated";

interface PhoneContextValue {
  effectiveAppConfig?: PortalAPIAppConfig;
  resetToken?: unknown;
}

interface FormState {
  phone: string;
}

const defaultState: FormState = {
  phone: "",
};

const PhoneContext = createContext<PhoneContextValue>({});

function PhoneField(props: { onChange: (value: string) => void }) {
  const { effectiveAppConfig, resetToken } = useContext(PhoneContext);
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

const PhoneScreen: React.VFC = function PhoneScreen() {
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
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      {
        to: `~/users/${user?.id}/details`,
        label: <FormattedMessage id="UserDetailsScreen.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="Add2FAPhoneScreen.title" />,
      },
    ];
  }, [user?.id]);

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
      await createAuthenticator({
        type: "oob_otp_sms",
        phone: state.phone,
        kind: AuthenticatorKind.Secondary,
      });
      await refetchUser();
    },
    [createAuthenticator, refetchUser]
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

  const onPhoneChange = useCallback(
    (value: string) => form.setState((state) => ({ ...state, phone: value })),
    [form]
  );

  const canSave = form.state.phone.length > 0;

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
    <PhoneContext.Provider value={contextValue}>
      <FormContainer form={form} canSave={canSave}>
        <ScreenContent>
          {title}
          <PhoneField onChange={onPhoneChange} />
        </ScreenContent>
      </FormContainer>
    </PhoneContext.Provider>
  );
};

export default PhoneScreen;
