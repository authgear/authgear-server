import React, { useMemo, useRef } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Heading, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { formatOptionalDuration, parseDuration } from "../../util/duration";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import {
  useAppConfigForm,
  AppConfigFormModel,
} from "../../hook/useAppConfigForm";
import LockoutSettings, { State as LockoutFormState } from "./LockoutSettings";
import { APIError } from "../../error/error";
import {
  LocalValidationError,
  makeLocalValidationError,
} from "../../error/validation";
import styles from "./AccountLockoutScreen.module.css";

interface FormState extends LockoutFormState {}

function parseOptionalDurationIntoMinutes(s: string | undefined) {
  if (s == null || s === "") {
    return undefined;
  }
  return parseDuration(s) / 60;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const isLockoutEnabled =
    (config.authentication?.lockout?.max_attempts ?? 0) > 0;

  return {
    isEnabled: isLockoutEnabled,
    maxAttempts: isLockoutEnabled
      ? config.authentication?.lockout?.max_attempts
      : 10,
    historyDurationMins: isLockoutEnabled
      ? parseOptionalDurationIntoMinutes(
          config.authentication?.lockout?.history_duration
        )
      : 1440,
    minimumDurationMins: isLockoutEnabled
      ? parseOptionalDurationIntoMinutes(
          config.authentication?.lockout?.minimum_duration
        )
      : 1,
    maximumDurationMins: isLockoutEnabled
      ? parseOptionalDurationIntoMinutes(
          config.authentication?.lockout?.maximum_duration
        )
      : 60,
    backoffFactorRaw: isLockoutEnabled
      ? config.authentication?.lockout?.backoff_factor?.toString()
      : "2",
    lockoutType: config.authentication?.lockout?.lockout_type ?? "per_user",
    isEnabledForPassword: isLockoutEnabled
      ? config.authentication?.lockout?.password?.enabled ?? false
      : true,
    isEnabledForTOTP: isLockoutEnabled
      ? config.authentication?.lockout?.totp?.enabled ?? false
      : true,
    isEnabledForOOBOTP: isLockoutEnabled
      ? config.authentication?.lockout?.oob_otp?.enabled ?? false
      : true,
    isEnabledForRecoveryCode: isLockoutEnabled
      ? config.authentication?.lockout?.recovery_code?.enabled ?? false
      : true,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.authentication ??= {};
    config.authentication.lockout ??= {};

    if (!currentState.isEnabled) {
      config.authentication.lockout = undefined;
    } else {
      const backoffFactor = Number(currentState.backoffFactorRaw);
      config.authentication.lockout.backoff_factor = Number.isFinite(
        backoffFactor
      )
        ? backoffFactor
        : undefined;
      config.authentication.lockout.history_duration = formatOptionalDuration(
        currentState.historyDurationMins,
        "m"
      );
      config.authentication.lockout.lockout_type = currentState.lockoutType;
      config.authentication.lockout.max_attempts = currentState.maxAttempts;
      config.authentication.lockout.maximum_duration = formatOptionalDuration(
        currentState.maximumDurationMins,
        "m"
      );
      config.authentication.lockout.minimum_duration = formatOptionalDuration(
        currentState.minimumDurationMins,
        "m"
      );
      if (currentState.isEnabledForOOBOTP) {
        config.authentication.lockout.oob_otp = { enabled: true };
      } else {
        config.authentication.lockout.oob_otp = undefined;
      }
      if (currentState.isEnabledForPassword) {
        config.authentication.lockout.password = { enabled: true };
      } else {
        config.authentication.lockout.password = undefined;
      }
      if (currentState.isEnabledForRecoveryCode) {
        config.authentication.lockout.recovery_code = { enabled: true };
      } else {
        config.authentication.lockout.recovery_code = undefined;
      }
      if (currentState.isEnabledForTOTP) {
        config.authentication.lockout.totp = { enabled: true };
      } else {
        config.authentication.lockout.totp = undefined;
      }
    }

    clearEmptyObject(config);
  });
}

function validateFormState(state: FormState): APIError | null {
  if (!state.isEnabled) {
    return null;
  }

  const errors: LocalValidationError[] = [];

  if ((state.maxAttempts ?? 0) < 1) {
    errors.push({
      messageID: "errors.validation.minimum",
      arguments: {
        minimum: 1,
      },
      location: "/authentication/lockout/max_attempts",
    });
  }

  if (
    [
      state.isEnabledForOOBOTP,
      state.isEnabledForPassword,
      state.isEnabledForRecoveryCode,
      state.isEnabledForTOTP,
    ].every((enabled) => !enabled)
  ) {
    errors.push({
      messageID:
        "LoginMethodConfigurationScreen.lockout.errors.mustEnableForAtLeastOneAuthenticator",
    });
  }

  if (errors.length < 1) {
    return null;
  }

  return makeLocalValidationError(errors);
}

interface AccountLockoutContentProps {
  form: AppConfigFormModel<FormState>;
}

const AccountLockoutContent: React.VFC<AccountLockoutContentProps> =
  function AccountLockoutContent(props) {
    const { form } = props;
    const { state, setState } = form;

    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="AccountLockoutScreen.title" />
          </Heading>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="AccountLockoutScreen.description" />
          </Text>
        </div>
        <LockoutSettings
          className={styles.widget}
          {...state}
          setState={setState}
        />
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const AccountLockoutScreen: React.VFC = function AccountLockoutScreen() {
  const { appID } = useParams() as { appID: string };

  const form = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
    validate: validateFormState,
  });

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <AccountLockoutContent form={form} />
    </FormContainer>
  );
};

export default AccountLockoutScreen;
