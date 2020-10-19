import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Label, TextField, Toggle } from "@fluentui/react";
import cn from "classnames";
import produce from "immer";
import deepEqual from "deep-equal";

import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { useAppConfigQuery } from "./query/appConfigQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ToggleWithContent from "../../ToggleWithContent";
import { PortalAPIAppConfig, PortalAPIApp } from "../../types";
import {
  clearEmptyObject,
  setFieldIfChanged,
  setNumericFieldIfChanged,
} from "../../util/misc";

import styles from "./SessionSettings.module.scss";

interface SessionSettingsProps {
  className?: string;
}

interface SessionProps {
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
}

interface SessionState {
  cookiePersistent: boolean;
  idleTimeoutEnabled: boolean;
  idleTimeoutSeconds: string;
  lifetimeSeconds: string;
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): SessionState {
  return {
    cookiePersistent: !(appConfig?.session?.cookie_non_persistent ?? false),
    idleTimeoutEnabled: appConfig?.session?.idle_timeout_enabled ?? false,
    idleTimeoutSeconds:
      appConfig?.session?.idle_timeout_seconds?.toString() ?? "",
    lifetimeSeconds: appConfig?.session?.lifetime_seconds?.toString() ?? "",
  };
}

function constructAppConfigFromState(
  state: SessionState,
  initialState: SessionState,
  appConfig: PortalAPIAppConfig
) {
  return produce(appConfig, (draftConfig) => {
    draftConfig.session = draftConfig.session ?? {};

    setFieldIfChanged(
      draftConfig.session,
      "cookie_non_persistent",
      !initialState.cookiePersistent,
      !state.cookiePersistent
    );

    setFieldIfChanged(
      draftConfig.session,
      "idle_timeout_enabled",
      initialState.idleTimeoutEnabled,
      state.idleTimeoutEnabled
    );

    setNumericFieldIfChanged(
      draftConfig.session,
      "idle_timeout_seconds",
      initialState.idleTimeoutSeconds,
      state.idleTimeoutSeconds
    );

    setNumericFieldIfChanged(
      draftConfig.session,
      "lifetime_seconds",
      initialState.lifetimeSeconds,
      state.lifetimeSeconds
    );

    clearEmptyObject(draftConfig);
  });
}

const SessionForm: React.FC<SessionProps> = function SessionForm(props) {
  const {
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
  } = props;

  const { renderToString } = useContext(Context);

  const initialState = useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [state, setState] = useState(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onCookiePersistentChange = useCallback((_event, checked?: boolean) => {
    if (checked === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      cookiePersistent: checked,
    }));
  }, []);

  const onLifetimeSecondsChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      lifetimeSeconds: value,
    }));
  }, []);

  const onIdleTimeoutEnabledChange = useCallback(
    (_event, checked?: boolean) => {
      if (checked === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        idleTimeoutEnabled: checked,
      }));
    },
    []
  );

  const onIdleTimeoutSecondsChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      idleTimeoutSeconds: value,
    }));
  }, []);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      if (rawAppConfig == null) {
        return;
      }

      const newAppConfig = constructAppConfigFromState(
        state,
        initialState,
        rawAppConfig
      );

      updateAppConfig(newAppConfig).catch(() => {});
    },
    [state, rawAppConfig, updateAppConfig, initialState]
  );

  return (
    <form onSubmit={onFormSubmit}>
      <Toggle
        className={styles.toggle}
        inlineLabel={true}
        label={renderToString("SessionSettings.persistent-cookie.label")}
        checked={state.cookiePersistent}
        onChange={onCookiePersistentChange}
      />
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("SessionSettings.session-lifetime.label")}
        value={state.lifetimeSeconds}
        onChange={onLifetimeSecondsChange}
      />
      <ToggleWithContent
        className={styles.toggleWithContent}
        inlineLabel={true}
        checked={state.idleTimeoutEnabled}
        onChange={onIdleTimeoutEnabledChange}
      >
        <Label className={styles.toggleLabel}>
          <FormattedMessage id="SessionSettings.invalidate-session-after-idling.label" />
        </Label>
        <TextField
          className={styles.textField}
          type="number"
          min="1"
          step="1"
          disabled={!state.idleTimeoutEnabled}
          label={renderToString("SessionSettings.idle-timeout.label")}
          value={state.idleTimeoutSeconds}
          onChange={onIdleTimeoutSecondsChange}
        />
      </ToggleWithContent>

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified}
          loading={updatingAppConfig}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </form>
  );
};

const SessionSettings: React.FC<SessionSettingsProps> = function SessionSettings(
  props
) {
  const { className } = props;

  const { appID } = useParams();

  const {
    loading,
    error,
    effectiveAppConfig,
    rawAppConfig,
    refetch,
  } = useAppConfigQuery(appID);
  const {
    loading: updatingAppConfig,
    error: updateAppConfigError,
    updateAppConfig,
  } = useUpdateAppConfigMutation(appID);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main
      className={cn(styles.root, className, {
        [styles.loading]: updatingAppConfig,
      })}
    >
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <SessionForm
        effectiveAppConfig={effectiveAppConfig}
        rawAppConfig={rawAppConfig}
        updateAppConfig={updateAppConfig}
        updatingAppConfig={updatingAppConfig}
      />
    </main>
  );
};

export default SessionSettings;
