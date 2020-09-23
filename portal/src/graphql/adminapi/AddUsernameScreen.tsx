import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import deepEqual from "deep-equal";

import UserDetailCommandBar from "./UserDetailCommandBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useCreateLoginIDIdentityMutation } from "./mutations/createIdentityMutation";
import { useTextField } from "../../hook/useInput";
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";
import { parseError } from "../../util/error";

import styles from "./AddUsernameScreen.module.scss";
import ShowError from "../../ShowError";

const AddUsernameScreen: React.FC = function AddUsernameScreen() {
  const { userID } = useParams();
  const navigate = useNavigate();

  const {
    createIdentity,
    loading: creatingIdentity,
    error: createIdentityError,
  } = useCreateLoginIDIdentityMutation(userID);
  const { renderToString } = useContext(Context);

  const [violations, setViolations] = useState<Violation[]>([]);
  const [unhandledViolations, setUnhandledViolations] = useState<Violation[]>(
    []
  );
  const [disableBlockNavigation, setDisableBlockNavigation] = useState<boolean>(
    false
  );

  const { value: username, onChange: onUsernameChange } = useTextField("");

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUsernameScreen.title" /> },
    ];
  }, []);

  const screenState = useMemo(
    () => ({
      username,
    }),
    [username]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual({ username: "" }, screenState);
  }, [screenState]);

  const onAddClicked = useCallback(() => {
    setDisableBlockNavigation(true);
    createIdentity({ key: "username", value: username })
      .then((identity) => {
        if (identity != null) {
          navigate("../");
        } else {
          throw new Error();
        }
      })
      .catch((err) => {
        setDisableBlockNavigation(false);
        const violations = parseError(err);
        setViolations(violations);
      });
  }, [username, navigate, createIdentity]);

  const errorMessage = useMemo(() => {
    const usernameFieldErrorMessages: string[] = [];
    const unknownViolations: Violation[] = [];
    for (const violation of violations) {
      if (violation.kind === "Invalid" || violation.kind === "format") {
        usernameFieldErrorMessages.push(
          renderToString("AddUsernameScreen.error.invalid-username")
        );
      } else if (violation.kind === "DuplicatedIdentity") {
        usernameFieldErrorMessages.push(
          renderToString("AddUsernameScreen.error.duplicated-username")
        );
      } else {
        unknownViolations.push(violation);
      }
    }

    setUnhandledViolations(unknownViolations);

    return {
      username: defaultFormatErrorMessageList(usernameFieldErrorMessages),
    };
  }, [violations, renderToString]);

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
          className={styles.usernameField}
          label={renderToString("AddUsernameScreen.username.label")}
          value={username}
          onChange={onUsernameChange}
          errorMessage={errorMessage.username}
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

export default AddUsernameScreen;
