import React, {useCallback, useContext, useEffect, useMemo, useState,} from "react";
import {useParams} from "react-router-dom";
import {Dropdown, Label} from "@fluentui/react";
import {Context, FormattedMessage} from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import AddIdentityForm from "./AddIdentityForm";
import {useDropdown, useIntegerTextField} from "../../hook/useInput";
import {useAppConfigQuery} from "../portal/query/appConfigQuery";
import {useUserQuery} from "./query/userQuery";
import {PortalAPIAppConfig} from "../../types";
import {passwordFieldErrorRules} from "../../PasswordField";
import {GenericErrorHandlingRule} from "../../error/useGenericError";
import {getActiveCountryCallingCode} from "../../util/countryCallingCode";

import styles from "./AddPhoneScreen.module.scss";

function makePhoneNumber(countryCode: string, phone: string) {
  if (phone.length === 0) {
    return "";
  }
  return `+${countryCode}${phone}`;
}

interface PhoneFieldProps {
  config: PortalAPIAppConfig | null;
  resetToken: unknown;

  value: string;
  onChange: (value: string) => void;
}

const PhoneField: React.FC<PhoneFieldProps> = function PhoneField(props) {
  const { config, resetToken, onChange } = props;
  const { renderToString } = useContext(Context);

  const countryCodes = useMemo(() => {
    const countryCodeConfig = config?.ui?.country_calling_code;
    const allowList = countryCodeConfig?.allowlist ?? [];
    const pinnedList = countryCodeConfig?.pinned_list ?? [];
    return getActiveCountryCallingCode(pinnedList, allowList);
  }, [config]);
  const defaultCountryCode = countryCodes[0];

  const [countryCode, setCountryCode] = useState(defaultCountryCode);
  const [phone, setPhone] = useState("");
  useEffect(() => {
    // Reset internal state when form is reset.
    setPhone("");
    setCountryCode(defaultCountryCode);
  }, [resetToken, defaultCountryCode]);

  const displayCountryCode = useCallback((countryCode: string) => {
    return `+ ${countryCode}`;
  }, []);

  const {
    options: countryCodeOptions,
    onChange: onCountryCodeChange,
  } = useDropdown(
    countryCodes,
    (option) => {
      setCountryCode(option);
      onChange(makePhoneNumber(option, phone));
    },
    defaultCountryCode,
    displayCountryCode
  );

  const { onChange: onPhoneChange } = useIntegerTextField((value) => {
    setPhone(value);
    onChange(makePhoneNumber(countryCode, value));
  });

  return (
    <section className={styles.phoneNumberFields}>
      <Label className={styles.phoneNumberLabel}>
        <FormattedMessage id="AddPhoneScreen.phone.label" />
      </Label>
      <Dropdown
        className={styles.countryCode}
        options={countryCodeOptions}
        selectedKey={countryCode}
        onChange={onCountryCodeChange}
        ariaLabel={renderToString("AddPhoneScreen.country-code.label")}
      />
      <FormTextField
        jsonPointer="phone"
        parentJSONPointer=""
        fieldName="phone"
        fieldNameMessageID="AddPhoneScreen.phone.label"
        hideLabel={true}
        className={styles.phone}
        value={phone}
        onChange={onPhoneChange}
        ariaLabel={renderToString("AddPhoneScreen.phone.label")}
      />
    </section>
  );
};

const AddPhoneScreen: React.FC = function AddPhoneScreen() {
  const { appID, userID } = useParams();
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
  } = useAppConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddPhoneScreen.title" /> },
    ];
  }, []);
  const title = <NavBreadcrumb items={navBreadcrumbItems} />;

  const rules: GenericErrorHandlingRule[] = useMemo(
    () => [
      {
        reason: "InvariantViolated",
        kind: "DuplicatedIdentity",
        errorMessageID: "AddPhoneScreen.error.duplicated-phone-number",
        field: "phone",
      },
      ...passwordFieldErrorRules,
    ],
    []
  );

  const [resetToken, setResetToken] = useState({});
  const renderPhoneField = useCallback(
    (props: Pick<PhoneFieldProps, "value"| "onChange">) => {
      return <PhoneField config={effectiveAppConfig} resetToken={resetToken} {...props} />;
    },
    [effectiveAppConfig, resetToken]
  );
  const onReset = useCallback(() => {
    setResetToken({});
  },[]);

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
    <AddIdentityForm
      appConfig={effectiveAppConfig}
      rawUser={user}
      loginIDType="phone"
      title={title}
      loginIDField={renderPhoneField}
      errorRules={rules}
      onReset={onReset}
    />
  );
};

export default AddPhoneScreen;
