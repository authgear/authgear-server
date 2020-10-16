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
  handleLocalPasswordViolations,
  handlePasswordPolicyViolatedViolation,
  localValidatePassword,
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
import {
  defaultFormatErrorMessageList,
  Violation,
} from "../../util/validation";
import { parseError } from "../../util/error";
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

  const [localViolations, setLocalViolations] = useState<Violation[]>([]);
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
    return !deepEqual({ newPassword: "", confirmPassword: "" }, formData);
  }, [formData]);

  const resetForm = useCallback(() => {
    setFormData(initialFormData);
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

      const newLocalViolations: Violation[] = [];
      localValidatePassword(
        newLocalViolations,
        passwordPolicy,
        formData.newPassword,
        formData.confirmPassword
      );
      setLocalViolations(newLocalViolations);
      if (newLocalViolations.length > 0) {
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
    [formData, passwordPolicy, resetPassword]
  );

  useEffect(() => {
    if (submittedForm) {
      navigate("..#account-security");
    }
  }, [submittedForm, navigate]);

  const { errorMessages, unhandledViolations } = useMemo(() => {
    const violations =
      localViolations.length > 0
        ? localViolations
        : parseError(resetPasswordError);
    const newPasswordErrorMessages: string[] = [];
    const confirmPasswordErrorMessages: string[] = [];
    const unhandledViolations: Violation[] = [];
    for (const violation of violations) {
      if (violation.kind === "custom") {
        handleLocalPasswordViolations(
          renderToString,
          violation,
          newPasswordErrorMessages,
          confirmPasswordErrorMessages,
          unhandledViolations
        );
      } else if (violation.kind === "PasswordPolicyViolated") {
        handlePasswordPolicyViolatedViolation(
          renderToString,
          violation,
          newPasswordErrorMessages,
          unhandledViolations
        );
      } else {
        unhandledViolations.push(violation);
      }
    }

    const errorMessages = {
      newPassword: defaultFormatErrorMessageList(newPasswordErrorMessages),
      confirmPassword: defaultFormatErrorMessageList(
        confirmPasswordErrorMessages
      ),
    };

    return { errorMessages, unhandledViolations };
  }, [localViolations, resetPasswordError, renderToString]);

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
      {unhandledViolations.length > 0 && (
        <ShowError error={resetPasswordError} />
      )}
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
      <PasswordField
        className={styles.newPasswordField}
        textFieldClassName={styles.passwordField}
        errorMessage={errorMessages.newPassword}
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
        errorMessage={errorMessages.confirmPassword}
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
