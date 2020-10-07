import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Dropdown, Label, TextField } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import UserDetailCommandBar from "./UserDetailCommandBar";
import { useDropdown, useTextField } from "../../hook/useInput";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { PortalAPIAppConfig } from "../../types";
import { parseError } from "../../util/error";
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";

import styles from "./AddPhoneScreen.module.scss";

interface AddPhoneFormProps {
  appConfig: PortalAPIAppConfig | null;
}

const AddPhoneForm: React.FC<AddPhoneFormProps> = function AddPhoneForm(
  props: AddPhoneFormProps
) {
  const { appConfig } = props;
  const { userID } = useParams();
  const navigate = useNavigate();

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);
  const { renderToString } = useContext(Context);

  const [submittedForm, setSubmittedForm] = useState<boolean>(false);

  const { value: phone, onChange: _onPhoneChange } = useTextField("");

  const onPhoneChange = useCallback(
    (_event, value?: string) => {
      if (value != null && /^[0-9]*$/.test(value)) {
        _onPhoneChange(_event, value);
      }
    },
    [_onPhoneChange]
  );

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
    }),
    [countryCode, phone]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(
      { countryCode: countryCodeConfig.default, phone: "" },
      screenState
    );
  }, [screenState, countryCodeConfig.default]);

  const onAddClicked = useCallback(() => {
    const combinedPhone = `+${countryCode}${phone}`;
    createIdentity({ key: "phone", value: combinedPhone })
      .then((identity) => {
        if (identity != null) {
          setSubmittedForm(true);
        }
      })
      .catch(() => {});
  }, [countryCode, phone, createIdentity]);

  useEffect(() => {
    if (submittedForm) {
      navigate("../#connected-identities");
    }
  }, [submittedForm, navigate]);

  const { errorMessage, unhandledViolations } = useMemo(() => {
    const violations = parseError(createIdentityError);
    const phoneNumberFieldErrorMessages: string[] = [];
    const unhandledViolations: Violation[] = [];
    for (const violation of violations) {
      if (violation.kind === "Invalid" || violation.kind === "format") {
        phoneNumberFieldErrorMessages.push(
          renderToString("AddPhoneScreen.error.invalid-phone-number")
        );
      } else if (violation.kind === "DuplicatedIdentity") {
        phoneNumberFieldErrorMessages.push(
          renderToString("AddPhoneScreen.error.duplicated-phone-number")
        );
      } else {
        unhandledViolations.push(violation);
      }
    }

    const errorMessage = {
      phoneNumber: defaultFormatErrorMessageList(phoneNumberFieldErrorMessages),
    };

    return { errorMessage, unhandledViolations };
  }, [createIdentityError, renderToString]);

  return (
    <div className={styles.form}>
      {unhandledViolations.length > 0 && (
        <ShowError error={createIdentityError} />
      )}
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
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
        <TextField
          className={styles.phone}
          value={phone}
          onChange={onPhoneChange}
          ariaLabel={renderToString("AddPhoneScreen.phone.label")}
          errorMessage={errorMessage.phoneNumber}
        />
      </section>
      <ButtonWithLoading
        onClick={onAddClicked}
        disabled={!isFormModified}
        labelId="add"
        loading={creatingIdentity}
      />
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
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
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
