import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Dropdown, PrimaryButton, TextField, Toggle } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import UserDetailCommandBar from "./UserDetailCommandBar";
import { useDropdown, useTextField } from "../../hook/useInput";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { PortalAPIAppConfig } from "../../types";

import styles from "./AddPhoneScreen.module.scss";

interface AddPhoneFormProps {
  appConfig: PortalAPIAppConfig | null;
}

const AddPhoneForm: React.FC<AddPhoneFormProps> = function AddPhoneForm(
  props: AddPhoneFormProps
) {
  const { appConfig } = props;
  const { renderToString } = useContext(Context);

  const { value: phone, onChange: _onPhoneChange } = useTextField("");

  const onPhoneChange = useCallback(
    (_event, value?: string) => {
      if (value != null && /^[0-9]*$/.test(value)) {
        _onPhoneChange(_event, value);
      }
    },
    [_onPhoneChange]
  );
  const [verified, setVerified] = useState(false);

  const onVerifiedToggled = useCallback((_event: any, checked?: boolean) => {
    if (checked == null) {
      return;
    }
    setVerified(checked);
  }, []);

  const countryCodeConfig = useMemo(() => {
    const countryCodeConfig = appConfig?.ui?.country_calling_code;
    const values = countryCodeConfig?.values ?? [];
    return {
      values,
      default: countryCodeConfig?.default ?? values[0],
    };
  }, [appConfig]);

  const displayCountryCode = useCallback((countryCode: string) => {
    return `+ ${countryCode}`;
  }, []);

  const {
    options: countryCodeOptions,
    onChange: onCountryCodeChange,
    selectedKey: countryCode,
  } = useDropdown(
    countryCodeConfig.values,
    countryCodeConfig.default,
    displayCountryCode
  );

  const screenState = useMemo(
    () => ({
      countryCode,
      phone,
      verified,
    }),
    [countryCode, phone, verified]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(
      { countryCode: countryCodeConfig.default, phone: "", verified: false },
      screenState
    );
  }, [screenState, countryCodeConfig.default]);

  const onAddClicked = useCallback(() => {
    // TODO: mutation to be integrated
  }, []);

  return (
    <div className={styles.form}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <section className={styles.phoneNumberFields}>
        <Dropdown
          className={styles.countryCode}
          label={renderToString("AddPhoneScreen.phone.label")}
          options={countryCodeOptions}
          selectedKey={countryCode}
          onChange={onCountryCodeChange}
          ariaLabel={renderToString("AddPhoneScreen.country-code.label")}
        />
        <TextField
          className={styles.phone}
          value={phone}
          onChange={onPhoneChange}
          ariaLabel={renderToString("AddPhoneScreen.phone.label")}
        />
      </section>
      <Toggle
        className={styles.verified}
        label={<FormattedMessage id="verified" />}
        inlineLabel={true}
        checked={verified}
        onChange={onVerifiedToggled}
      />
      <PrimaryButton onClick={onAddClicked} disabled={!isFormModified}>
        <FormattedMessage id="add" />
      </PrimaryButton>
    </div>
  );
};

const AddPhoneScreen: React.FC = function AddPhoneScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddPhoneScreen.title" /> },
    ];
  }, []);

  const appConfig =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;

  if (loading) {
    <ShowLoading />;
  }

  if (error != null) {
    <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <div className={styles.root}>
      <UserDetailCommandBar />
      <section className={styles.content}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <AddPhoneForm appConfig={appConfig} />
      </section>
    </div>
  );
};

export default AddPhoneScreen;
