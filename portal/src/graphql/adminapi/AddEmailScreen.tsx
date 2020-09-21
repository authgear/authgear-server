import React, { useCallback, useContext, useMemo, useState } from "react";
import { PrimaryButton, TextField, Toggle } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { useTextField } from "../../hook/useInput";

import styles from "./AddEmailScreen.module.scss";

const AddEmailScreen: React.FC = function AddEmailScreen() {
  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddEmailScreen.title" /> },
    ];
  }, []);

  const { value: email, onChange: onEmailChange } = useTextField("");
  const [verified, setVerified] = useState(false);

  const { renderToString } = useContext(Context);

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
    // TODO: mutation to be integrated
  }, []);

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <section className={styles.content}>
        <NavigationBlockerDialog blockNavigation={isFormModified} />
        <TextField
          className={styles.emailField}
          label={renderToString("AddEmailScreen.email.label")}
          value={email}
          onChange={onEmailChange}
        />
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
      </section>
    </div>
  );
};

export default AddEmailScreen;
