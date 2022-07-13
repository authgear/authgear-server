import React, {
  createContext,
  useMemo,
  useState,
  useContext,
  useCallback,
  useEffect,
} from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import AddIdentityForm, { LoginIDFieldProps } from "./AddIdentityForm";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import {
  ErrorParseRule,
  makeInvariantViolatedErrorParseRule,
} from "../../error/parse";
import { PortalAPIAppConfig } from "../../types";
import FormPhoneTextField from "../../FormPhoneTextField";

import styles from "./AddPhoneScreen.module.css";

const errorRules: ErrorParseRule[] = [
  makeInvariantViolatedErrorParseRule(
    "DuplicatedIdentity",
    "AddPhoneScreen.error.duplicated-phone-number"
  ),
];

interface PhoneContextValue {
  effectiveAppConfig?: PortalAPIAppConfig;
  resetToken?: unknown;
}

const PhoneContext = createContext<PhoneContextValue>({});

function LoginIDField(props: LoginIDFieldProps) {
  const { effectiveAppConfig, resetToken } = useContext(PhoneContext);
  const [inputValue, setInputValue] = useState("");
  const { onChange } = props;
  const onChangeValues = useCallback(
    (valid: string, input: string) => {
      onChange(valid);
      setInputValue(input);
    },
    [onChange]
  );
  useEffect(() => {
    setInputValue("");
  }, [resetToken]);
  return (
    <FormPhoneTextField
      parentJSONPointer=""
      fieldName="login_id"
      errorRules={errorRules}
      className={styles.widget}
      allowlist={effectiveAppConfig?.ui?.phone_input?.allowlist}
      pinnedList={effectiveAppConfig?.ui?.phone_input?.pinned_list}
      inputValue={inputValue}
      onChange={onChangeValues}
    />
  );
}

const AddPhoneScreen: React.FC = function AddPhoneScreen() {
  const { appID, userID } = useParams() as { appID: string; userID: string };
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

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "./../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "./..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddPhoneScreen.title" /> },
    ];
  }, []);

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
      <AddIdentityForm
        appConfig={effectiveAppConfig}
        rawUser={user}
        loginIDType="phone"
        title={title}
        loginIDField={LoginIDField}
        onReset={onReset}
      />
    </PhoneContext.Provider>
  );
};

export default AddPhoneScreen;
