import cn from "classnames";
import {
  Context as MessageContext,
  FormattedMessage,
} from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import React, { useCallback, useContext, useMemo } from "react";
import { SimpleFormModel } from "../../hook/useSimpleForm";
import { PortalAPIAppConfig } from "../../types";
import { Checkbox, ChoiceGroup, IChoiceGroupOption } from "@fluentui/react";
import { useCheckbox, useTextField } from "../../hook/useInput";
import styles from "./ResetPasswordForm.module.css";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import TextField from "../../TextField";
import PasswordField from "../../PasswordField";

export enum PasswordCreationType {
  ManualEntry = "manual_entry",
  AutoGenerate = "auto_generate",
}

export interface FormState {
  newPassword: string;
  passwordCreationType: PasswordCreationType;
  sendPassword: boolean;
  setPasswordExpired: boolean;
}

interface ResetPasswordFormProps {
  className?: string;
  appConfig: PortalAPIAppConfig | null;
  form: SimpleFormModel<FormState>;
  firstEmail: string | null;
  submitMessageID: string;
}

export const ResetPasswordForm: React.VFC<ResetPasswordFormProps> = function (
  props
) {
  const {
    className,
    appConfig,
    form: { state, setState },
    firstEmail,
    submitMessageID,
  } = props;
  const { renderToString } = useContext(MessageContext);

  const { canSave, isUpdating, onSubmit } =
    useFormContainerBaseContext<SimpleFormModel<FormState, string | null>>();

  const passwordCreateionTypeOptions = useMemo(() => {
    return [
      {
        key: PasswordCreationType.ManualEntry,
        text: renderToString("ResetPasswordForm.password-creation-type.manual"),
      },
      {
        key: PasswordCreationType.AutoGenerate,
        text: renderToString("ResetPasswordForm.password-creation-type.auto"),
      },
    ];
  }, [renderToString]);

  const onChangePasswordCreationType = useCallback(
    (_e, option: IChoiceGroupOption | undefined) => {
      if (option != null) {
        setState((prev) => ({
          ...prev,
          newPassword:
            option.key === PasswordCreationType.AutoGenerate
              ? ""
              : prev.newPassword,
          passwordCreationType: option.key as PasswordCreationType,
          sendPassword:
            prev.sendPassword ||
            option.key === PasswordCreationType.AutoGenerate,
        }));
      }
    },
    [setState]
  );

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
    <form
      className={cn(className, styles.form)}
      onSubmit={onSubmit}
      noValidate={true}
    >
      {firstEmail != null ? (
        <div>
          <TextField
            label={renderToString("ResetPasswordForm.email")}
            type="email"
            value={firstEmail}
            disabled={true}
          />
          <ChoiceGroup
            selectedKey={state.passwordCreationType}
            options={passwordCreateionTypeOptions}
            onChange={onChangePasswordCreationType}
          />
        </div>
      ) : null}
      <div>
        <PasswordField
          label={renderToString("ResetPasswordForm.new-password")}
          value={state.newPassword}
          onChange={onNewPasswordChange}
          passwordPolicy={appConfig?.authenticator?.password?.policy ?? {}}
          parentJSONPointer=""
          fieldName="password"
          canRevealPassword={true}
          canGeneratePassword={true}
          disabled={
            state.passwordCreationType === PasswordCreationType.AutoGenerate
          }
        />
        <Checkbox
          className={styles.checkbox}
          label={renderToString("ResetPasswordForm.send-password")}
          checked={state.sendPassword}
          onChange={onChangeSendPassword}
          disabled={
            firstEmail == null ||
            state.passwordCreationType === PasswordCreationType.AutoGenerate
          }
        />
        <Checkbox
          className={styles.checkbox}
          label={renderToString("ResetPasswordForm.force-change-on-login")}
          checked={state.setPasswordExpired}
          onChange={onChangeForceChangeOnLogin}
        />
      </div>
      <div>
        <PrimaryButton
          disabled={!canSave || isUpdating}
          type="submit"
          text={<FormattedMessage id={submitMessageID} />}
        />
      </div>
    </form>
  );
};
