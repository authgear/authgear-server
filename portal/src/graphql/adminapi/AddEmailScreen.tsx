import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { TextField, Toggle } from "@fluentui/react";
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

  const [unhandledViolations, setUnhandledViolations] = useState<Violation[]>(
    []
  );
  const [disableBlockNavigation, setDisableBlockNavigation] = useState<boolean>(
    false
  );

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddEmailScreen.title" /> },
    ];
  }, []);

  const { value: email, onChange: onEmailChange } = useTextField("");
  const [verified, setVerified] = useState(false);

  const onVerifiedToggled = useCallback((_event: any, checked?: boolean) => {
    if (checked == null) {
      return;
    }
    setVerified(checked);
  }, []);

  const screenState = useMemo(
    () => ({
      email,
      verified,
    }),
    [email, verified]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual({ email: "", verified: false }, screenState);
  }, [screenState]);

  const onAddClicked = useCallback(() => {
    setDisableBlockNavigation(true);
    createIdentity({ key: "email", value: email })
      .then((identity) => {
        if (identity != null) {
          navigate("../#connected-identities");
        }
      })
      .catch(() => {
        setDisableBlockNavigation(false);
      });
  }, [email, navigate, createIdentity]);

  const errorMessage = useMemo(() => {
    const violations = parseError(createIdentityError);
    const emailFieldErrorMessages: string[] = [];
    const unknownViolations: Violation[] = [];
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
        unknownViolations.push(violation);
      }
    }

    setUnhandledViolations(unknownViolations);

    return {
      email: defaultFormatErrorMessageList(emailFieldErrorMessages),
    };
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
          blockNavigation={!disableBlockNavigation && isFormModified}
        />
        <TextField
          className={styles.emailField}
          label={renderToString("AddEmailScreen.email.label")}
          value={email}
          onChange={onEmailChange}
          errorMessage={errorMessage.email}
        />
        <Toggle
          className={styles.verified}
          label={<FormattedMessage id="verified" />}
          inlineLabel={true}
          checked={verified}
          onChange={onVerifiedToggled}
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
