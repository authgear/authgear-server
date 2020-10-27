import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";
import deepEqual from "deep-equal";
import { Text, TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useResetPasswordMutation } from "./mutations/resetPasswordMutation";
import NavBreadcrumb from "../../NavBreadcrumb";
import PasswordField, {
  localValidatePassword,
  passwordFieldErrorRules,
  PasswordFieldLocalErrorMessageMap,
} from "../../PasswordField";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import { useAppConfigQuery } from "../portal/query/appConfigQuery";
import { useTextField } from "../../hook/useInput";
import { useGenericError } from "../../error/useGenericError";
import { PortalAPIAppConfig } from "../../types";

import styles from "./ResetPasswordScreen.module.scss";

interface ResetPasswordFormProps {
  appConfig: PortalAPIAppConfig | null;
}

interface ResetPasswordFormData {
  newPassword: string;
  confirmPassword: string;
}

const ResetPasswordForm: React.FC<ResetPasswordFormProps> = function (
  props: ResetPasswordFormProps
) {
  const { appConfig } = props;

  const { userID } = useParams();
  const navigate = useNavigate();
  const {
    resetPassword,
    loading: resettingPassword,
    error: resetPasswordError,
  } = useResetPasswordMutation(userID);
  const { renderToString } = useContext(Context);

  const [
    localValidationErrorMessageMap,
    setLocalValidationErrorMessageMap,
  ] = useState<PasswordFieldLocalErrorMessageMap>(null);
  const [submittedForm, setSubmittedForm] = useState(false);

  const passwordPolicy = useMemo(() => {
    return appConfig?.authenticator?.password?.policy ?? {};
  }, [appConfig]);

  const initialFormData = useMemo<ResetPasswordFormData>(() => {
    return {
      newPassword: "",
      confirmPassword: "",
    };
  }, []);
  const [formData, setFormData] = useState(initialFormData);
  const { newPassword, confirmPassword } = formData;

  const isFormModified = useMemo(() => {
    return !deepEqual(initialFormData, formData);
  }, [formData, initialFormData]);

  const resetForm = useCallback(() => {
    setFormData(initialFormData);
    setLocalValidationErrorMessageMap(null);
  }, [initialFormData]);

  const { onChange: onNewPasswordChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, newPassword: value }));
  });
  const { onChange: onConfirmPasswordChange } = useTextField((value) => {
    setFormData((prev) => ({ ...prev, confirmPassword: value }));
  });

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      const localErrorMessageMap = localValidatePassword(
        renderToString,
        passwordPolicy,
        formData.newPassword,
        formData.confirmPassword
      );
      setLocalValidationErrorMessageMap(localErrorMessageMap);

      if (localErrorMessageMap != null) {
        return;
      }

      resetPassword(formData.newPassword)
        .then((userID) => {
          if (userID != null) {
            setSubmittedForm(true);
          }
        })
        .catch(() => {});
    },
    [renderToString, formData, passwordPolicy, resetPassword]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("..#account-security");
    }
  }, [submittedForm, navigate]);

  const { errorMessage, unrecognizedError } = useGenericError(
    resetPasswordError,
    [],
    passwordFieldErrorRules
  );

  if (appConfig == null) {
    return (
      <Text>
        <FormattedMessage id="ResetPasswordScreen.error.fetch-password-policy" />
      </Text>
    );
  }

  return (
    <form className={styles.form} onSubmit={onFormSubmit}>
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      {unrecognizedError && <ShowError error={unrecognizedError} />}
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
      <PasswordField
        className={styles.newPasswordField}
        textFieldClassName={styles.passwordField}
        errorMessage={localValidationErrorMessageMap?.password ?? errorMessage}
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
        errorMessage={localValidationErrorMessageMap?.confirmPassword}
      />
      <ButtonWithLoading
        type="submit"
        className={styles.confirm}
        disabled={!isFormModified || submittedForm}
        loading={resettingPassword}
        labelId="confirm"
      />
    </form>
  );
};

const ResetPasswordScreen: React.FC = function ResetPasswordScreen() {
  const { appID } = useParams();
  const { effectiveAppConfig, loading, error, refetch } = useAppConfigQuery(
    appID
  );

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "../../..", label: <FormattedMessage id="UsersScreen.title" /> },
      { to: "../", label: <FormattedMessage id="UserDetailsScreen.title" /> },
      { to: ".", label: <FormattedMessage id="ResetPasswordScreen.title" /> },
    ];
  }, []);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.content}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <ResetPasswordForm appConfig={effectiveAppConfig} />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default ResetPasswordScreen;
