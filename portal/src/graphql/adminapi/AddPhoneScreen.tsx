import React, {
  useCallback,
  useContext,
  useMemo,
  useEffect,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Dropdown, Label, Text } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import { useDropdown, useIntegerTextField } from "../../hook/useInput";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { PortalAPIAppConfig } from "../../types";
import { useValidationError } from "../../error/useValidationError";
import { useGenericError } from "../../error/useGenericError";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import { FormContext } from "../../error/FormContext";
import { getActiveCountryCallingCode } from "../../util/countryCallingCode";
import { canCreateLoginIDIdentity } from "../../util/loginID";

import styles from "./AddPhoneScreen.module.scss";

interface AddPhoneFormProps {
  appConfig: PortalAPIAppConfig | null;
  resetForm: () => void;
}

const AddPhoneForm: React.FC<AddPhoneFormProps> = function AddPhoneForm(
  props: AddPhoneFormProps
) {
  const { appConfig, resetForm } = props;
  const { userID } = useParams();
  const navigate = useNavigate();

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);
  const { renderToString } = useContext(Context);

  const countryCodeConfig = useMemo(() => {
    const countryCodeConfig = appConfig?.ui?.country_calling_code;
    const allowList = countryCodeConfig?.allowlist ?? [];
    const pinnedList = countryCodeConfig?.pinned_list ?? [];
    const values = getActiveCountryCallingCode(pinnedList, allowList);
    const defaultCallingCode = values[0];
    return {
      values,
      defaultCallingCode,
    };
  }, [appConfig]);

  const initialFormData = useMemo(() => {
    return {
      phone: "",
      countryCode: countryCodeConfig.defaultCallingCode,
    };
  }, [countryCodeConfig]);

  const [submittedForm, setSubmittedForm] = useState<boolean>(false);
  const [formData, setFormData] = useState(initialFormData);

  const { phone, countryCode } = formData;

  const { onChange: onPhoneChange } = useIntegerTextField((value) => {
    setFormData((prev) => ({
      ...prev,
      phone: value,
    }));
  });

  const displayCountryCode = useCallback((countryCode: string) => {
    return `+ ${countryCode}`;
  }, []);

  const {
    options: countryCodeOptions,
    onChange: onCountryCodeChange,
  } = useDropdown(
    countryCodeConfig.values,
    (option) => {
      setFormData((prev) => ({
        ...prev,
        countryCode: option,
      }));
    },
    countryCodeConfig.defaultCallingCode,
    displayCountryCode
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(initialFormData, formData);
  }, [formData, initialFormData]);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      const combinedPhone = `+${countryCode}${phone}`;
      createIdentity({ key: "phone", value: combinedPhone })
        .then((identity) => {
          if (identity != null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [countryCode, phone, createIdentity]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("..#connected-identities");
    }
  }, [submittedForm, navigate]);

  const {
    unhandledCauses: rawUnhandledCauses,
    otherError,
    value: formContextValue,
  } = useValidationError(createIdentityError);

  const {
    errorMessage: genericErrorMessage,
    unrecognizedError,
    unhandledCauses,
  } = useGenericError(otherError, rawUnhandledCauses, [
    {
      reason: "InvariantViolated",
      kind: "DuplicatedIdentity",
      errorMessageID: "AddPhoneScreen.error.duplicated-phone-number",
    },
  ]);

  if (!canCreateLoginIDIdentity(appConfig)) {
    return (
      <Text className={styles.helpText}>
        <FormattedMessage id="CreateIdentity.require-login-id" />
      </Text>
    );
  }

  return (
    <FormContext.Provider value={formContextValue}>
      <form className={styles.form} onSubmit={onFormSubmit}>
        {unrecognizedError && <ShowError error={unrecognizedError} />}
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        <NavigationBlockerDialog
          blockNavigation={!submittedForm && isFormModified}
        />
        <ModifiedIndicatorPortal
          resetForm={resetForm}
          isModified={isFormModified}
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
          <FormTextField
            jsonPointer=""
            parentJSONPointer=""
            fieldName="phone"
            fieldNameMessageID="AddPhoneScreen.phone.label"
            hideLabel={true}
            className={styles.phone}
            value={phone}
            onChange={onPhoneChange}
            ariaLabel={renderToString("AddPhoneScreen.phone.label")}
            errorMessage={genericErrorMessage}
          />
        </section>
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified || submittedForm}
          labelId="add"
          loading={creatingIdentity}
        />
      </form>
    </FormContext.Provider>
  );
};

const AddPhoneScreen: React.FC = function AddPhoneScreen() {
  const { appID } = useParams();
  const { effectiveAppConfig, loading, error, refetch } = useAppConfigQuery(
    appID
  );

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddPhoneScreen.title" /> },
    ];
  }, []);

  const [remountIdentifier, setRemountIdentifier] = useState(0);
  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <div className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.content}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <AddPhoneForm
          key={remountIdentifier}
          appConfig={effectiveAppConfig}
          resetForm={resetForm}
        />
      </ModifiedIndicatorWrapper>
    </div>
  );
};

export default AddPhoneScreen;
