import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { TextField } from "@fluentui/react";

import NavBreadcrumb, { BreadcrumbItem } from "./NavBreadcrumb";
import ButtonWithLoading from "./ButtonWithLoading";
import NavigationBlockerDialog from "./NavigationBlockerDialog";

import styles from "./InviteAdminScreen.module.scss";

const InviteAdminContent: React.FC = function InviteAdminContent() {
  const { renderToString } = useContext(Context);

  const [email, setEmail] = useState("");

  const isFormModified = useMemo(() => {
    return email !== "";
  }, [email]);

  const creatingCollaboratorInvitation = false;

  const onEmailChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setEmail(value);
  }, []);

  const onFormSubmit = useCallback((ev: React.SyntheticEvent<HTMLElement>) => {
    ev.preventDefault();
    ev.stopPropagation();

    // TODO: handle create collaborator invitation mutation
    alert("Not yet implemented");
  }, []);

  return (
    <form className={styles.content} onSubmit={onFormSubmit}>
      <TextField
        className={styles.emailField}
        type="text"
        label={renderToString("InviteAdminScreen.email.label")}
        value={email}
        onChange={onEmailChange}
      />
      <ButtonWithLoading
        type="submit"
        disabled={!isFormModified}
        labelId="InviteAdminScreen.add-user.label"
        loading={creatingCollaboratorInvitation}
      />
      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </form>
  );
};

const InviteAdminScreen: React.FC = function InviteAdminScreen() {
  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "../", label: <FormattedMessage id="SettingsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="InviteAdminScreen.title" /> },
    ];
  }, []);

  return (
    <main className={styles.root}>
      <NavBreadcrumb className={styles.breadcrumb} items={navBreadcrumbItems} />
      <InviteAdminContent />
    </main>
  );
};

export default InviteAdminScreen;
