import React, { useCallback, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { Heading } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";

import Link from "../../Link";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppAndSecretConfigQuery } from "../portal/query/appAndSecretConfigQuery";
import { PortalAPIAppConfig } from "../../types";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import ScreenContent from "../../ScreenContent";
import styles from "./ChangePasswordScreen.module.css";
import { validatePassword } from "../../error/password";
import { useUserQuery } from "./query/userQuery";
import { FormContainerBase } from "../../FormContainerBase";
import ErrorDialog from "../../error/ErrorDialog";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import { ResetPasswordForm } from "../../components/users/ResetPasswordForm";
import { useResetPasswordMutation } from "./mutations/resetPasswordMutation";

enum PasswordCreationType {
  ManualEntry = "manual_entry",
  AutoGenerate = "auto_generate",
}

interface FormState {
  newPassword: string;
  passwordCreationType: PasswordCreationType;
  sendPassword: boolean;
  setPasswordExpired: boolean;
}

const defaultState: FormState = {
  newPassword: "",
  passwordCreationType: PasswordCreationType.ManualEntry,
  sendPassword: false,
  setPasswordExpired: true,
};

interface ResetPasswordContentProps {
  appConfig: PortalAPIAppConfig | null;
  form: SimpleFormModel<FormState>;
  firstEmail: string | null;
}

const ChangePasswordContent: React.VFC<ResetPasswordContentProps> = function (
  props
) {
  const { appConfig, form, firstEmail } = props;

  return (
    <ScreenContent>
      <div className={styles.widget}>
        <Link to="..#account-security" className={styles.backLink}>
          <ChevronLeftIcon className={styles.backLinkIcon} />
          <span>
            <FormattedMessage id="UserDetailsScreen.title" />
          </span>
        </Link>
        <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="ChangePasswordScreen.title" />
        </Heading>
      </div>
      <ResetPasswordForm
        className={`${styles.widget} ${styles.form}`}
        submitMessageID="ChangePasswordScreen.change"
        form={form}
        appConfig={appConfig}
        firstEmail={firstEmail}
      />
    </ScreenContent>
  );
};

const ChangePasswordScreen: React.VFC = function ChangePasswordScreen() {
  const { appID } = useParams() as { appID: string };
  const { userID } = useParams() as { userID: string };

  const navigate = useNavigate();

  const {
    effectiveAppConfig,
    isLoading: loadingConfig,
    loadError: configError,
    refetch: refetchConfig,
  } = useAppAndSecretConfigQuery(appID);

  const {
    user,
    loading: loadingUser,
    error: userError,
    refetch: refetchUser,
  } = useUserQuery(userID);

  const firstEmail: string | null = useMemo(() => {
    const emailIdentities =
      user?.identities?.edges
        ?.filter((identityEdge) => {
          const identity = identityEdge?.node;
          return (
            identity?.type === "LOGIN_ID" &&
            identity.claims["https://authgear.com/claims/login_id/type"] ===
              "email"
          );
        })
        .map((identity) => identity?.node) ?? [];
    if (emailIdentities.length > 0) {
      return emailIdentities[0]?.claims.email ?? null;
    }
    return null;
  }, [user]);

  const passwordPolicy = useMemo(
    () => effectiveAppConfig?.authenticator?.password?.policy ?? {},
    [effectiveAppConfig]
  );

  const { resetPassword, error: resetPasswordError } =
    useResetPasswordMutation(userID);

  const resetPasswordErrorRules: ErrorParseRule[] = useMemo(() => {
    return [
      makeReasonErrorParseRule(
        "SendPasswordNoTarget",
        "ChangePasswordScreen.error.send-password-no-target"
      ),
    ];
  }, []);

  const validate = useCallback(
    (state: FormState) => {
      if (state.passwordCreationType === PasswordCreationType.AutoGenerate) {
        return null;
      }
      return validatePassword(state.newPassword, passwordPolicy);
    },
    [passwordPolicy]
  );

  const submit = useCallback(
    async (state: FormState) => {
      const newPassword =
        state.passwordCreationType === PasswordCreationType.AutoGenerate
          ? ""
          : state.newPassword;
      await resetPassword(
        newPassword,
        state.sendPassword,
        state.setPasswordExpired
      );
    },
    [resetPassword]
  );

  const form = useSimpleForm({
    defaultState,
    submit,
    validate,
  });

  const canSave =
    form.state.passwordCreationType === PasswordCreationType.AutoGenerate ||
    form.state.newPassword.length > 0;

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("./..#account-security");
    }
  }, [form.isSubmitted, navigate]);

  if (loadingUser || loadingConfig) {
    return <ShowLoading />;
  }

  if (configError != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={configError} onRetry={refetchConfig} />;
  }

  if (userError != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={userError} onRetry={refetchUser} />;
  }

  return (
    <FormContainerBase form={form} canSave={canSave}>
      <ChangePasswordContent
        form={form}
        appConfig={effectiveAppConfig}
        firstEmail={firstEmail}
      />
      <ErrorDialog error={resetPasswordError} rules={resetPasswordErrorRules} />
    </FormContainerBase>
  );
};

export default ChangePasswordScreen;
