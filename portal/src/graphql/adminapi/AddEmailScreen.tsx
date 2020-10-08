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

  const { value: email, onChange: onEmailChange } = useTextField("");

  const screenState = useMemo(
    () => ({
      email,
    }),
    [email]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual({ email: "" }, screenState);
  }, [screenState]);

  const onAddClicked = useCallback(() => {
    createIdentity({ key: "email", value: email })
      .then((identity) => {
        if (identity != null) {
          setSubmittedForm(true);
        }
      })
      .catch(() => {});
  }, [email, createIdentity]);

  useEffect(() => {
    if (submittedForm) {
      navigate("../#connected-identities");
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
      <NavBreadcrumb className={styles.breadcrumb} items={navBreadcrumbItems} />
      <section className={styles.content}>
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
          onClick={onAddClicked}
          disabled={!isFormModified}
          labelId="add"
          loading={creatingIdentity}
        />
      </section>
    </div>
  );
};

export default AddEmailScreen;
