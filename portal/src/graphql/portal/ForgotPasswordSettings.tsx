import React, { useMemo, useContext, useState, useCallback } from "react";
import { Context } from "@oursky/react-messageformat";
import { TextField } from "@fluentui/react";
import cn from "classnames";
import deepEqual from "deep-equal";
import produce from "immer";

import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import { setFieldIfChanged, clearEmptyObject } from "../../util/misc";
import { PortalAPIApp, PortalAPIAppConfig } from "../../types";

import styles from "./ForgotPasswordSettings.module.scss";

interface ForgotPasswordSettingsProps {
  className?: string;
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
  resetForm: () => void;
}

interface ForgotPasswordSettingsState {
  resetCodeExpirySeconds: number;
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): ForgotPasswordSettingsState {
  const forgot_password = appConfig?.forgot_password;

  return {
    resetCodeExpirySeconds: forgot_password?.reset_code_expiry_seconds ?? 0,
  };
}

function constructAppConfigFromState(
  rawAppConfig: PortalAPIAppConfig,
  initialScreenState: ForgotPasswordSettingsState,
  screenState: ForgotPasswordSettingsState
): PortalAPIAppConfig {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    draftConfig.forgot_password = draftConfig.forgot_password ?? {};

    const forgotPassword = draftConfig.forgot_password;

    setFieldIfChanged(
      forgotPassword,
      "reset_code_expiry_seconds",
      initialScreenState.resetCodeExpirySeconds,
      screenState.resetCodeExpirySeconds
    );

    clearEmptyObject(draftConfig);
  });

  return newAppConfig;
}

const ForgotPasswordSettings: React.FC<ForgotPasswordSettingsProps> = function ForgotPasswordSettings(
  props
) {
  const {
    className,
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
    resetForm,
  } = props;

  const { renderToString } = useContext(Context);

  const initialState = useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [state, setState] = useState(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onResetCodeExpirySecondsChange = useCallback(
    (_event, value?: string) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        resetCodeExpirySeconds: parseInt(value, 10),
      }));
    },
    []
  );

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      if (rawAppConfig == null) {
        return;
      }

      const newAppConfig = constructAppConfigFromState(
        rawAppConfig,
        initialState,
        state
      );

      updateAppConfig(newAppConfig).catch(() => {});
    },
    [rawAppConfig, initialState, state, updateAppConfig]
  );

  return (
    <form className={cn(styles.root, className)} onSubmit={onFormSubmit}>
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <TextField
        className={styles.textField}
        type="number"
        min="0"
        step="1"
        label={renderToString(
          "PasswordsScreen.forgot-password.time-to-invalid-reset-code.label"
        )}
        value={`${state.resetCodeExpirySeconds}`}
        onChange={onResetCodeExpirySecondsChange}
      />

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified}
          loading={updatingAppConfig}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
    </form>
  );
};

export default ForgotPasswordSettings;
