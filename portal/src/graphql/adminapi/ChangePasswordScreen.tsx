import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
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
import ScreenContent from "../../ScreenContent";
import FormTextField from "../../FormTextField";

import styles from "./ChangePasswordScreen.module.css";
import { validatePassword } from "../../error/password";

interface FormState {
  newPassword: string;
  confirmPassword: string;
}

const defaultState: FormState = {
  newPassword: "",
  confirmPassword: "",
};

interface ResetPasswordContentProps {
  appConfig: PortalAPIAppConfig | null;
  form: SimpleFormModel<FormState>;
}

const ChangePasswordContent: React.VFC<ResetPasswordContentProps> = function (
  props
) {
  const {
    appConfig,
    form: { state, setState },
  } = props;
  const { renderToString } = useContext(Context);
  const { userID } = useParams() as { userID: string };

  const navBreadcrumbItems = useMemo(() => {
    return [
      { to: "~/users", label: <FormattedMessage id="UsersScreen.title" /> },
      {
        to: `~/users/${userID}/details`,
        label: <FormattedMessage id="UserDetailsScreen.title" />,
      },
      { to: ".", label: <FormattedMessage id="ChangePasswordScreen.title" /> },
    ];
  }, [userID]);

  const { onChange: onNewPasswordChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, newPassword: value }));
  });
  const { onChange: onConfirmPasswordChange } = useTextField((value) => {
    setState((prev) => ({ ...prev, confirmPassword: value }));
  });

  return (
    <ScreenContent>
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
      <PasswordField
        className={styles.widget}
        label={renderToString("ChangePasswordScreen.new-password")}
        value={state.newPassword}
        onChange={onNewPasswordChange}
        passwordPolicy={appConfig?.authenticator?.password?.policy ?? {}}
        parentJSONPointer=""
        fieldName="password"
      />
      <FormTextField
        className={styles.widget}
        label={renderToString("ChangePasswordScreen.confirm-password")}
        type="password"
        value={state.confirmPassword}
        onChange={onConfirmPasswordChange}
        parentJSONPointer=""
        fieldName="confirm_password"
      />
    </ScreenContent>
  );
};

const ChangePasswordScreen: React.VFC = function ChangePasswordScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const { effectiveAppConfig, loading, error, refetch } =
    useAppAndSecretConfigQuery(appID);
  const passwordPolicy = useMemo(
    () => effectiveAppConfig?.authenticator?.password?.policy ?? {},
    [effectiveAppConfig]
  );

  const { userID } = useParams() as { userID: string };
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

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
  });

  const canSave =
    form.state.newPassword.length > 0 && form.state.confirmPassword.length > 0;

  useEffect(() => {
    if (form.isSubmitted) {
      navigate("./..#account-security");
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
      <ChangePasswordContent form={form} appConfig={effectiveAppConfig} />
    </FormContainer>
  );
};

export default ChangePasswordScreen;
