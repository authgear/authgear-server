import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import { FormattedMessage } from "@oursky/react-messageformat";

import UserDetailCommandBar from "./UserDetailCommandBar";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import NavBreadcrumb from "../../NavBreadcrumb";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ShowError from "../../ShowError";
import FormTextField from "../../FormTextField";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { useTextField } from "../../hook/useInput";
import { FormContext } from "../../error/FormContext";
import { useValidationError } from "../../error/useValidationError";
import { useGenericError } from "../../error/useGenericError";

import styles from "./AddEmailScreen.module.scss";

interface AddEmailFormData {
  email: string;
}

const AddEmailScreen: React.FC = function AddEmailScreen() {
  const { userID } = useParams();
  const navigate = useNavigate();

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);

  const [submittedForm, setSubmittedForm] = useState<boolean>(false);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddEmailScreen.title" /> },
    ];
  }, []);

  const initialFormData = useMemo(() => {
    return {
      email: "",
    };
  }, []);
  const [formData, setFormData] = useState<AddEmailFormData>(initialFormData);
  const { email } = formData;

  const { onChange: onEmailChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, email: value }));
  });

  const isFormModified = useMemo(() => {
    return !deepEqual(initialFormData, formData);
  }, [initialFormData, formData]);

  const resetForm = useCallback(() => {
    setFormData(initialFormData);
  }, [initialFormData]);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      createIdentity({ key: "email", value: email })
        .then((identity) => {
          if (identity != null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [email, createIdentity]
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
      errorMessageID: "AddEmailScreen.error.duplicated-email",
    },
  ]);

  return (
    <div className={styles.root}>
      <UserDetailCommandBar />
      <ModifiedIndicatorWrapper>
        <NavBreadcrumb
          className={styles.breadcrumb}
          items={navBreadcrumbItems}
        />
        <ModifiedIndicatorPortal
          resetForm={resetForm}
          isModified={isFormModified}
        />
        <FormContext.Provider value={formContextValue}>
          <form className={styles.content} onSubmit={onFormSubmit}>
            {unrecognizedError && (
              <div className={styles.error}>
                <ShowError error={unrecognizedError} />
              </div>
            )}
            <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
            <NavigationBlockerDialog
              blockNavigation={!submittedForm && isFormModified}
            />
            <FormTextField
              jsonPointer=""
              parentJSONPointer=""
              fieldName="email"
              fieldNameMessageID="AddEmailScreen.email.label"
              className={styles.emailField}
              value={email}
              onChange={onEmailChange}
              errorMessage={genericErrorMessage}
            />
            <ButtonWithLoading
              type="submit"
              disabled={!isFormModified || submittedForm}
              labelId="add"
              loading={creatingIdentity}
            />
          </form>
        </FormContext.Provider>
      </ModifiedIndicatorWrapper>
    </div>
  );
};

export default AddEmailScreen;
