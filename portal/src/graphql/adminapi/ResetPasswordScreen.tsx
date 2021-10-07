import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useResetPasswordMutation } from "./mutations/resetPasswordMutation";
import NavBreadcrumb from "../../NavBreadcrumb";
import PasswordField from "../../PasswordField";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { useTextField } from "../../hook/useInput";
import { PortalAPIAppConfig } from "../../types";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import FormContainer from "../../FormContainer";
import FormTextField from "../../FormTextField";

import styles from "./ResetPasswordScreen.module.scss";
import { validatePassword } from "../../error/password";

interface FormState {
  newPassword: string;
  confirmPassword: string;
}

const defaultFormState: FormState = {
  newPassword: "",
  confirmPassword: "",
};

interface ResetPasswordContentProps {
  appConfig: PortalAPIAppConfig | null;
  form: SimpleFormModel<FormState>;
}

const ResetPasswordContent: React.FC<ResetPasswordContentProps> = function (
  props
) {
  const {
    appConfig,
    form: { state, setState },
  } = props;
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "..", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="ResetPasswordScreen.title" /> },
    ];
  }, []);

  const { onChange: onNewPasswordChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, newPassword: value }));
  });
  const { onChange: onConfirmPasswordChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, confirmPassword: value }));
  });

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <PasswordField
        className={styles.newPasswordField}
        textFieldClassName={styles.passwordField}
        label={renderToString("ResetPasswordScreen.new-password")}
        value={state.newPassword}
        onChange={onNewPasswordChange}
        passwordPolicy={appConfig?.authenticator?.password?.policy ?? {}}
        parentJSONPointer=""
        fieldName="password"
      />
      <FormTextField
        className={cn(styles.passwordField, styles.confirmPasswordField)}
        label={renderToString("ResetPasswordScreen.confirm-password")}
        type="password"
        value={state.confirmPassword}
        onChange={onConfirmPasswordChange}
        parentJSONPointer=""
        fieldName="confirm_password"
      />
    </div>
  );
};

const ResetPasswordScreen: React.FC = function ResetPasswordScreen() {
  const { appID } = useParams();
  const navigate = useNavigate();

  const { effectiveAppConfig, loading, error, refetch } =
    useAppAndSecretConfigQuery(appID);
  const passwordPolicy = useMemo(
    () => effectiveAppConfig?.authenticator?.password?.policy ?? {},
    [effectiveAppConfig]
  );

  const { userID } = useParams();
  const { resetPassword } = useResetPasswordMutation(userID);

  const validate = useCallback(
    (state: FormState) => {
      return validatePassword(
        state.newPassword,
        passwordPolicy,
        state.confirmPassword
      );
    },
    [passwordPolicy]
  );

  const submit = useCallback(
    async (state: FormState) => {
      await resetPassword(state.newPassword);
    },
    [resetPassword]
  );

  const form = useSimpleForm(defaultFormState, submit, validate);

  const canSave =
    form.state.newPassword.length > 0 && form.state.confirmPassword.length > 0;

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("..#account-security");
    }
  }, [form.isSubmitted, navigate]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContainer form={form} canSave={canSave}>
      <ResetPasswordContent form={form} appConfig={effectiveAppConfig} />
    </FormContainer>
  );
};

export default ResetPasswordScreen;
