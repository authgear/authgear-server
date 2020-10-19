import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import { TextField } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import UserDetailCommandBar from "./UserDetailCommandBar";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import NavBreadcrumb from "../../NavBreadcrumb";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ShowError from "../../ShowError";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { useTextField } from "../../hook/useInput";
import { parseError } from "../../util/error";
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";

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
  const { renderToString } = useContext(Context);

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

  const { errorMessage, unhandledViolations } = useMemo(() => {
    const violations = parseError(createIdentityError);
    const emailFieldErrorMessages: string[] = [];
    const unhandledViolations: Violation[] = [];
    for (const violation of violations) {
      if (violation.kind === "Invalid" || violation.kind === "format") {
        emailFieldErrorMessages.push(
          renderToString("AddEmailScreen.error.invalid-email")
        );
      } else if (violation.kind === "DuplicatedIdentity") {
        emailFieldErrorMessages.push(
          renderToString("AddEmailScreen.error.duplicated-email")
        );
      } else {
        unhandledViolations.push(violation);
      }
    }

    const errorMessage = {
      email: defaultFormatErrorMessageList(emailFieldErrorMessages),
    };
    return { errorMessage, unhandledViolations };
  }, [createIdentityError, renderToString]);

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
        <form className={styles.content} onSubmit={onFormSubmit}>
          {unhandledViolations.length > 0 && (
            <ShowError error={createIdentityError} />
          )}
          <NavigationBlockerDialog
            blockNavigation={!submittedForm && isFormModified}
          />
          <TextField
            className={styles.emailField}
            label={renderToString("AddEmailScreen.email.label")}
            value={email}
            onChange={onEmailChange}
            errorMessage={errorMessage.email}
          />
          <ButtonWithLoading
            type="submit"
            disabled={!isFormModified || submittedForm}
            labelId="add"
            loading={creatingIdentity}
          />
        </form>
      </ModifiedIndicatorWrapper>
    </div>
  );
};

export default AddEmailScreen;
