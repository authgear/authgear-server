import React, { useCallback, useContext, useMemo } from "react";
import { PrimaryButton, TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import deepEqual from "deep-equal";

import UserDetailCommandBar from "./UserDetailCommandBar";
import NavBreadcrumb from "../../NavBreadcrumb";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { useTextField } from "../../hook/useInput";

import styles from "./AddUsernameScreen.module.scss";

const AddUsernameScreen: React.FC = function AddUsernameScreen() {
  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUsernameScreen.title" /> },
    ];
  }, []);

  const { value: username, onChange: onUsernameChange } = useTextField("");

  const { renderToString } = useContext(Context);

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
    // TODO: mutation to be integrated
  }, []);

  return (
    <div className={styles.root}>
      <UserDetailCommandBar />
      <NavBreadcrumb className={styles.breadcrumb} items={navBreadcrumbItems} />
      <section className={styles.content}>
        <NavigationBlockerDialog blockNavigation={isFormModified} />
        <TextField
          className={styles.usernameField}
          label={renderToString("AddUsernameScreen.username.label")}
          value={username}
          onChange={onUsernameChange}
        />
        <PrimaryButton onClick={onAddClicked} disabled={!isFormModified}>
          <FormattedMessage id="add" />
        </PrimaryButton>
      </section>
    </div>
  );
};

export default AddUsernameScreen;
