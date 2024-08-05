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
import { useCheckbox, useTextField } from "../../hook/useInput";
import { PortalAPIAppConfig } from "../../types";
import { SimpleFormModel, useSimpleForm } from "../../hook/useSimpleForm";
import ScreenContent from "../../ScreenContent";

import styles from "./ChangePasswordScreen.module.css";
import { validatePassword } from "../../error/password";
import { Checkbox } from "@fluentui/react";
import TextField from "../../TextField";
import { useUserQuery } from "./query/userQuery";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
import PrimaryButton from "../../PrimaryButton";
import ErrorDialog from "../../error/ErrorDialog";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";

interface FormState {
  newPassword: string;
  sendPassword: boolean;
  setPasswordExpired: boolean;
}

const defaultState: FormState = {
  newPassword: "",
  sendPassword: false,
  setPasswordExpired: false,
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

  const { user } = useUserQuery(userID);

  const { canSave, isUpdating, onSubmit } =
    useFormContainerBaseContext<SimpleFormModel<FormState, string | null>>();

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
  const { onChange: onChangeSendPassword } = useCheckbox((value) => {
    setState((prev) => ({ ...prev, sendPassword: value }));
  });
  const { onChange: onChangeForceChangeOnLogin } = useCheckbox((value) => {
    setState((prev) => ({ ...prev, setPasswordExpired: value }));
  });

  return (
    <ScreenContent>
      <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
      <form
        className={cn(styles.widget, styles.form)}
        onSubmit={onSubmit}
        noValidate={true}
      >
        <TextField
          label={renderToString("ChangePasswordScreen.email")}
          type="email"
          value={user?.standardAttributes?.email ?? ""}
          disabled={true}
        />
        <div>
          <PasswordField
            label={renderToString("ChangePasswordScreen.new-password")}
            value={state.newPassword}
            onChange={onNewPasswordChange}
            passwordPolicy={appConfig?.authenticator?.password?.policy ?? {}}
            parentJSONPointer=""
            fieldName="password"
          />
          <Checkbox
            className={styles.checkbox}
            label={renderToString("ChangePasswordScreen.send-password")}
            checked={state.sendPassword}
            onChange={onChangeSendPassword}
          />
          <Checkbox
            className={styles.checkbox}
            label={renderToString("ChangePasswordScreen.force-change-on-login")}
            checked={state.setPasswordExpired}
            onChange={onChangeForceChangeOnLogin}
          />
        </div>
        <div>
          <PrimaryButton
            disabled={!canSave || isUpdating}
            type="submit"
            text={<FormattedMessage id="ChangePasswordScreen.change" />}
          />
        </div>
      </form>
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
  const { resetPassword, error: resetPasswordError } =
    useResetPasswordMutation(userID);

  const { refetch: refetchUser } = useUserQuery(userID, { skip: true });

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
      return validatePassword(state.newPassword, passwordPolicy);
    },
    [passwordPolicy]
  );

  const submit = useCallback(
    async (state: FormState) => {
      await resetPassword(
        state.newPassword,
        state.sendPassword,
        state.setPasswordExpired
      );
      await refetchUser();
    },
    [resetPassword, refetchUser]
  );

  const form = useSimpleForm({
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
    defaultState,
    submit,
    validate,
  });

  const canSave = form.state.newPassword.length > 0;

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
    <FormContainerBase form={form} canSave={canSave}>
      <ChangePasswordContent form={form} appConfig={effectiveAppConfig} />
      <ErrorDialog error={resetPasswordError} rules={resetPasswordErrorRules} />
    </FormContainerBase>
  );
};

export const ChangePasswordVeriticalFormLayout: React.VFC<
  React.PropsWithChildren<Record<never, never>>
> = function ChangePasswordVeriticalFormLayout({ children }) {
  return <div className={styles.verticalForm}>{children}</div>;
};

export default ChangePasswordScreen;
