import cn from "classnames";
import { FormattedMessage } from "../../intl";
import React, { useCallback } from "react";
import { SimpleFormModel } from "../../hook/useSimpleForm";
import { PortalAPIAppConfig } from "../../types";
import { Checkbox, Flex, RadioGroup, Text } from "@radix-ui/themes";
import styles from "./ResetPasswordForm.module.css";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { TextField } from "../v2/TextField/TextField";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";
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
  const { canSave, isUpdating, onSubmit } =
    useFormContainerBaseContext<SimpleFormModel<FormState, string | null>>();

  const onChangePasswordCreationType = useCallback(
    (value: string) => {
      const passwordCreationType = value as PasswordCreationType;
      setState((prev) => ({
        ...prev,
        newPassword:
          passwordCreationType === PasswordCreationType.AutoGenerate
            ? ""
            : prev.newPassword,
        passwordCreationType,
        sendPassword:
          prev.sendPassword ||
          passwordCreationType === PasswordCreationType.AutoGenerate,
      }));
    },
    [setState]
  );

  const onNewPasswordChange = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, newPassword: value }));
    },
    [setState]
  );
  const onChangeSendPassword = useCallback(
    (checked: boolean | "indeterminate") => {
      setState((prev) => ({
        ...prev,
        sendPassword: checked === true,
      }));
    },
    [setState]
  );
  const onChangeForceChangeOnLogin = useCallback(
    (checked: boolean | "indeterminate") => {
      setState((prev) => ({
        ...prev,
        setPasswordExpired: checked === true,
      }));
    },
    [setState]
  );

  return (
    <form
      className={cn(className, styles.form)}
      onSubmit={onSubmit}
      noValidate={true}
    >
      {firstEmail != null ? (
        <div className={styles.section}>
          <TextField
            size="2"
            label={<FormattedMessage id="ResetPasswordForm.email" />}
            type="email"
            value={firstEmail}
            disabled={true}
          />
          <RadioGroup.Root
            value={state.passwordCreationType}
            onValueChange={onChangePasswordCreationType}
          >
            <Flex direction="column" gap="3">
              <Text as="label" size="2" className={styles.optionLabel}>
                <Flex align="center" gap="2">
                  <RadioGroup.Item value={PasswordCreationType.ManualEntry} />
                  <FormattedMessage id="ResetPasswordForm.password-creation-type.manual" />
                </Flex>
              </Text>
              <Text as="label" size="2" className={styles.optionLabel}>
                <Flex align="center" gap="2">
                  <RadioGroup.Item value={PasswordCreationType.AutoGenerate} />
                  <FormattedMessage id="ResetPasswordForm.password-creation-type.auto" />
                </Flex>
              </Text>
            </Flex>
          </RadioGroup.Root>
        </div>
      ) : null}
      <div className={styles.section}>
        <PasswordField
          label={<FormattedMessage id="ResetPasswordForm.new-password" />}
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
        <label className={styles.checkboxRow}>
          <Checkbox
            checked={state.sendPassword}
            onCheckedChange={onChangeSendPassword}
            disabled={
              firstEmail == null ||
              state.passwordCreationType === PasswordCreationType.AutoGenerate
            }
          />
          <Text size="2">
            <FormattedMessage id="ResetPasswordForm.send-password" />
          </Text>
        </label>
        <label className={styles.checkboxRow}>
          <Checkbox
            checked={state.setPasswordExpired}
            onCheckedChange={onChangeForceChangeOnLogin}
          />
          <Text size="2">
            <FormattedMessage id="ResetPasswordForm.force-change-on-login" />
          </Text>
        </label>
      </div>
      <div className={styles.actions}>
        <PrimaryButton
          disabled={!canSave || isUpdating}
          loading={isUpdating}
          size="2"
          type="submit"
          text={<FormattedMessage id={submitMessageID} />}
        />
      </div>
    </form>
  );
};
