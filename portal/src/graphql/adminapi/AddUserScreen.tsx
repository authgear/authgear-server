import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Label, PrimaryButton, Text, TextField } from "@fluentui/react";

import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import PasswordField from "../../PasswordField";
import { useTextField } from "../../hook/useInput";
import { PortalAPIAppConfig } from "../../types";

import styles from "./AddUserScreen.module.scss";

interface AddUserContentProps {
  appConfig: PortalAPIAppConfig | null;
}

const AddUserContent: React.FC<AddUserContentProps> = function AddUserContent(
  props: AddUserContentProps
) {
  const { appConfig } = props;
  const { renderToString } = useContext(Context);

  const isFieldVisible = useMemo(() => {
    const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
    // We consider them as enabled if they are listed as allowed login ID keys.
    const username = loginIdKeys.find((key) => key.type === "username") != null;
    const email = loginIdKeys.find((key) => key.type === "email") != null;
    const phone = loginIdKeys.find((key) => key.type === "phone") != null;

    const passwordAuthenticatorKey = appConfig?.authentication?.primary_authenticators?.find(
      (authticator) => authticator === "password"
    );
    const password = !!passwordAuthenticatorKey;
    return {
      username,
      email,
      phone,
      password,
    };
  }, [appConfig]);

  // NOTE: cannot add user identity if none of three field is available
  const canAddUser =
    isFieldVisible.username || isFieldVisible.email || isFieldVisible.phone;

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  const { value: name, onChange: onNameChange } = useTextField("");
  const { value: username, onChange: onUsernameChange } = useTextField("");
  const { value: email, onChange: onEmailChange } = useTextField("");
  const { value: phone, onChange: onPhoneChange } = useTextField("");
  const { value: password, onChange: onPasswordChange } = useTextField("");

  const onClickAddUser = useCallback(() => {
    // TODO: to be implemented
  }, []);

  // TODO: improve empty state
  if (!canAddUser) {
    return (
      <Text>
        <FormattedMessage id="AddUserScreen.cannot-add-user" />
      </Text>
    );
  }

  return (
    <section className={styles.content}>
      <TextField
        className={styles.textField}
        label={renderToString("AddUserScreen.name")}
        value={name}
        onChange={onNameChange}
      />
      <Label className={styles.userInfoLabel}>
        <FormattedMessage id="AddUserScreen.user-info.label" />
      </Label>
      <section className={styles.userInfo}>
        {isFieldVisible.username && (
          <TextField
            className={styles.textField}
            label={renderToString("AddUserScreen.user-info.username")}
            value={username}
            onChange={onUsernameChange}
          />
        )}
        {isFieldVisible.email && (
          <TextField
            className={styles.textField}
            label={renderToString("AddUserScreen.user-info.email")}
            value={email}
            onChange={onEmailChange}
          />
        )}
        {isFieldVisible.phone && (
          <TextField
            className={styles.textField}
            label={renderToString("AddUserScreen.user-info.phone")}
            value={phone}
            onChange={onPhoneChange}
          />
        )}
      </section>
      {isFieldVisible.password && (
        <PasswordField
          textFieldClassName={styles.textField}
          label={renderToString("AddUserScreen.password.label")}
          value={password}
          onChange={onPasswordChange}
          passwordPolicy={passwordPolicy}
        />
      )}
      <PrimaryButton className={styles.addUserButton} onClick={onClickAddUser}>
        <FormattedMessage id="AddUserScreen.add-user.label" />
      </PrimaryButton>
    </section>
  );
};

const AddUserScreen: React.FC = function AddUserScreen() {
  const { appID } = useParams();

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: "../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: ".", label: <FormattedMessage id="AddUserScreen.title" /> },
    ];
  }, []);

  const { data, loading, error, refetch } = useAppConfigQuery(appID);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  const appConfig =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <AddUserContent appConfig={appConfig} />
    </main>
  );
};

export default AddUserScreen;
