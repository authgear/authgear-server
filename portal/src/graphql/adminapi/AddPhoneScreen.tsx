import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useParams } from "react-router-dom";
import { Dropdown, Label } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import AddIdentityForm from "./AddIdentityForm";
import { useDropdown, useIntegerTextField } from "../../hook/useInput";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useUserQuery } from "./query/userQuery";
import { PortalAPIAppConfig } from "../../types";
import { ErrorParseRule } from "../../error/parse";
import ALL_COUNTRIES from "../../data/country.json";

import styles from "./AddPhoneScreen.module.scss";

const errorRules: ErrorParseRule[] = [
  {
    reason: "InvariantViolated",
    kind: "DuplicatedIdentity",
    errorMessageID: "AddPhoneScreen.error.duplicated-phone-number",
  },
  {
    reason: "ValidationFailed",
    location: "",
    kind: "format",
    errorMessageID: "errors.validation.format",
  },
];

type Country = typeof ALL_COUNTRIES[number];
type CountryMap = Record<string, Country>;

const COUNTRY_MAP = ALL_COUNTRIES.reduce<CountryMap>(
  (acc: CountryMap, country: Country) => {
    acc[country.Alpha2] = country;
    return acc;
  },
  {}
);

function makePhoneNumber(alpha2: string, phone: string) {
  if (phone.length === 0) {
    return "";
  }
  const countryCallingCode = COUNTRY_MAP[alpha2].CountryCallingCode;
  return `+${countryCallingCode}${phone}`;
}

function getAlpha2(pinned: string[], allowed: string[]): string[] {
  const list = [...pinned];
  const pinnedSet = new Set(pinned);
  for (const alpha2 of allowed) {
    if (!pinnedSet.has(alpha2)) {
      list.push(alpha2);
    }
  }
  return list;
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

  const alpha2List = useMemo(() => {
    const phoneInputConfig = config?.ui?.phone_input;
    const allowList = phoneInputConfig?.allowlist ?? [];
    const pinnedList = phoneInputConfig?.pinned_list ?? [];
    return getAlpha2(pinnedList, allowList);
  }, [config]);
  const defaultAlpha2 = alpha2List[0];

  const [alpha2, setAlpha2] = useState(defaultAlpha2);
  const [phone, setPhone] = useState("");
  useEffect(() => {
    // Reset internal state when form is reset.
    setPhone("");
    setAlpha2(defaultAlpha2);
  }, [resetToken, defaultAlpha2]);

  const getLabel = useCallback((alpha2: string) => {
    const country = COUNTRY_MAP[alpha2];
    return `${country.Alpha2} +${country.CountryCallingCode}`;
  }, []);

  const { options: countryCodeOptions, onChange: onCountryCodeChange } =
    useDropdown(
      alpha2List,
      (option) => {
        setAlpha2(option);
        onChange(makePhoneNumber(option, phone));
      },
      alpha2,
      getLabel
    );

  const { onChange: onPhoneChange } = useIntegerTextField((value) => {
    setPhone(value);
    onChange(makePhoneNumber(alpha2, value));
  });

  return (
    <section className={styles.phoneNumberFields}>
      <Label className={styles.phoneNumberLabel}>
        <FormattedMessage id="AddPhoneScreen.phone.label" />
      </Label>
      <Dropdown
        className={styles.countryCode}
        options={countryCodeOptions}
        selectedKey={alpha2}
        onChange={onCountryCodeChange}
        ariaLabel={renderToString("AddPhoneScreen.country-code.label")}
      />
      <FormTextField
        parentJSONPointer="/"
        fieldName="phone"
        fieldNameMessageID="AddPhoneScreen.phone.label"
        errorRules={errorRules}
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
  } = useAppAndSecretConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddPhoneScreen.title" /> },
    ];
  }, []);
  const title = <NavBreadcrumb items={navBreadcrumbItems} />;

  const [resetToken, setResetToken] = useState({});
  const renderPhoneField = useCallback(
    (props: Pick<PhoneFieldProps, "value" | "onChange">) => {
      return (
        <PhoneField
          config={effectiveAppConfig}
          resetToken={resetToken}
          {...props}
        />
      );
    },
    [effectiveAppConfig, resetToken]
  );
  const onReset = useCallback(() => {
    setResetToken({});
  }, []);

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
      onReset={onReset}
    />
  );
};

export default AddPhoneScreen;
