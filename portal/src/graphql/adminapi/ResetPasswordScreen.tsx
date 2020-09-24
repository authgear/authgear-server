import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import zxcvbn from "zxcvbn";
import deepEqual from "deep-equal";
import { PrimaryButton, Text, TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb from "../../NavBreadcrumb";
import PasswordField, {
  extractGuessableLevel,
  isPasswordValid,
} from "../../PasswordField";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useTextField } from "../../hook/useInput";
import { PasswordPolicyConfig, PortalAPIAppConfig } from "../../types";

import styles from "./ResetPasswordScreen.module.scss";

interface ResetPasswordFormProps {
  appConfig: PortalAPIAppConfig | null;
}

interface ResetPasswordScreenState {
  newPassword: string;
  confirmPassword: string;
}

type ViolationType = "confirm-password-not-match" | "invalid-password";
interface ResetPasswordScreenViolation {
  violationType: ViolationType;
}

function validate(
  screenState: ResetPasswordScreenState,
  passwordPolicy: PasswordPolicyConfig
) {
  const violations: ResetPasswordScreenViolation[] = [];
  if (screenState.newPassword !== screenState.confirmPassword) {
    violations.push({ violationType: "confirm-password-not-match" });
  }

  const guessableLevel = extractGuessableLevel(zxcvbn(screenState.newPassword));
  const passwordValid = isPasswordValid(
    passwordPolicy,
    screenState.newPassword,
    guessableLevel
  );
  if (!passwordValid) {
    violations.push({ violationType: "invalid-password" });
  }

  return violations;
}

function hasViolation(
  violationType: ViolationType,
  violations: ResetPasswordScreenViolation[]
): boolean {
  return (
    violations.filter((violation) => violation.violationType === violationType)
      .length > 0
  );
}

const ResetPasswordForm: React.FC<ResetPasswordFormProps> = function (
  props: ResetPasswordFormProps
) {
  const { appConfig } = props;

  const { renderToString } = useContext(Context);

  const [violations, setViolations] = useState<ResetPasswordScreenViolation[]>(
    []
  );

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  const { value: newPassword, onChange: onNewPasswordChange } = useTextField(
    ""
  );
  const {
    value: confirmPassword,
    onChange: onConfirmPasswordChange,
  } = useTextField("");

  const screenState = useMemo(
    () => ({
      newPassword,
      confirmPassword,
    }),
    [newPassword, confirmPassword]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual({ newPassword: "", confirmPassword: "" }, screenState);
  }, [screenState]);

  const onConfirmClicked = useCallback(() => {
    const newViolations = validate(screenState, passwordPolicy);
    setViolations(newViolations);
    if (newViolations.length > 0) {
      // return
    }

    // TODO: integrate mutation
  }, [screenState, passwordPolicy]);

  const newPasswordErrorMessage = useMemo(() => {
    if (hasViolation("invalid-password", violations)) {
      return renderToString("ResetPasswordScreen.error.invalid-password");
    }
    return undefined;
  }, [violations, renderToString]);

  const confirmPasswordErrorMessage = useMemo(() => {
    if (hasViolation("confirm-password-not-match", violations)) {
      return renderToString(
        "ResetPasswordScreen.error.confirm-password-not-match"
      );
    }
    return undefined;
  }, [violations, renderToString]);

  if (appConfig == null) {
    return (
      <Text>
        <FormattedMessage id="ResetPasswordScreen.error.fetch-password-policy" />
      </Text>
    );
  }

  return (
    <div className={styles.form}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <PasswordField
        className={styles.newPasswordField}
        textFieldClassName={styles.passwordField}
        errorMessage={newPasswordErrorMessage}
        label={renderToString("ResetPasswordScreen.new-password")}
        value={newPassword}
        onChange={onNewPasswordChange}
        passwordPolicy={passwordPolicy}
      />
      <TextField
        className={cn(styles.passwordField, styles.confirmPasswordField)}
        label={renderToString("ResetPasswordScreen.confirm-password")}
        type="password"
        value={confirmPassword}
        onChange={onConfirmPasswordChange}
        errorMessage={confirmPasswordErrorMessage}
      />
      <PrimaryButton className={styles.confirm} onClick={onConfirmClicked}>
        <FormattedMessage id="confirm" />
      </PrimaryButton>
    </div>
  );
};

const ResetPasswordScreen: React.FC = function ResetPasswordScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppConfigQuery(appID);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="ResetPasswordScreen.title" /> },
    ];
  }, []);

  const appConfig =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      <section className={styles.content}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <ResetPasswordForm appConfig={appConfig} />
      </section>
    </main>
  );
};

export default ResetPasswordScreen;
