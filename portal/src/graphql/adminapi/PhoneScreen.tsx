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
import IdentityForm, { LoginIDFieldProps } from "./IdentityForm";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import { PortalAPIAppConfig } from "../../types";
import FormPhoneTextField from "../../FormPhoneTextField";

import styles from "./PhoneScreen.module.css";

const errorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "DuplicatedIdentity",
    "PhoneScreen.error.duplicated-phone-number"
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

const PhoneScreen: React.VFC = function PhoneScreen() {
  const { appID, userID, identityID } = useParams() as {
    appID: string;
    userID: string;
    identityID?: string;
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

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      {
        to: `~/users/${user?.id}/details`,
        label: <FormattedMessage id="UserDetailsScreen.title" />,
      },
      {
        to: ".",
        label: identityID ? (
          <FormattedMessage id="PhoneScreen.edit.title" />
        ) : (
          <FormattedMessage id="PhoneScreen.add.title" />
        ),
      },
    ];
  }, [identityID, user?.id]);

  const [resetToken, setResetToken] = useState({});

  const onReset = useCallback(() => {
    setResetToken({});
  }, []);

  const originalIdentity = useMemo(() => {
    if (!identityID) {
      return null;
    }
    const identity = user?.identities?.edges?.find((edge) => {
      const node = edge?.node;
      return node != null && node.id === identityID && node.claims.phone_number;
    });
    if (identity == null) {
      return null;
    }
    return {
      id: identity.node!.id,
      value: identity.node!.claims.phone_number!,
    };
  }, [identityID, user?.identities?.edges]);

  const currentValueMessage = useMemo(() => {
    if (originalIdentity == null) {
      return null;
    }
    return (
      <FormattedMessage
        id="PhoneScreen.edit.current-value"
        values={{ value: originalIdentity.value }}
      />
    );
  }, [originalIdentity]);

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
      <IdentityForm
        originalIdentityID={originalIdentity?.id ?? null}
        currentValueMessage={currentValueMessage}
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

export default PhoneScreen;
