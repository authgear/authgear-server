import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import { TextField, Toggle } from "@fluentui/react";
import cn from "classnames";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";

import styles from "./SessionConfigurationScreen.module.scss";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";

interface FormState {
  persistentCookie: boolean;
  sessionLifetimeSeconds: number;
  idleTimeoutEnabled: boolean;
  idleTimeoutSeconds: number;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    persistentCookie: !(config.session?.cookie_non_persistent ?? false),
    sessionLifetimeSeconds: config.session?.lifetime_seconds ?? 0,
    idleTimeoutEnabled: config.session?.idle_timeout_enabled ?? false,
    idleTimeoutSeconds: config.session?.idle_timeout_seconds ?? 0,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.session = config.session ?? {};
    if (initialState.persistentCookie !== currentState.persistentCookie) {
      config.session.cookie_non_persistent = !currentState.persistentCookie;
    }
    if (
      initialState.sessionLifetimeSeconds !==
      currentState.sessionLifetimeSeconds
    ) {
      config.session.lifetime_seconds = currentState.sessionLifetimeSeconds;
    }
    if (initialState.idleTimeoutEnabled !== currentState.idleTimeoutEnabled) {
      config.session.idle_timeout_enabled = currentState.idleTimeoutEnabled;
      if (
        currentState.idleTimeoutEnabled &&
        initialState.idleTimeoutSeconds !== currentState.idleTimeoutSeconds
      ) {
        config.session.idle_timeout_seconds = currentState.idleTimeoutSeconds;
      }
    }
    clearEmptyObject(config);
  });
}

interface HooksSettingsContentProps {
  form: AppConfigFormModel<FormState>;
}

const SessionConfigurationScreenContent: React.FC<HooksSettingsContentProps> = function SessionConfigurationScreenContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const onPersistentCookieChange = useCallback(
    (_, value?: boolean) => {
      setState((state) => ({
        ...state,
        persistentCookie: value ?? false,
      }));
    },
    [setState]
  );

  const onSessionLifetimeSecondsChange = useCallback(
    (_, value?: string) => {
      setState((state) => ({
        ...state,
        sessionLifetimeSeconds: Number(value),
      }));
    },
    [setState]
  );

  const onIdleTimeoutEnabledChange = useCallback(
    (_, value?: boolean) => {
      setState((state) => ({
        ...state,
        idleTimeoutEnabled: value ?? false,
      }));
    },
    [setState]
  );

  const onIdleTimeoutSecondsChange = useCallback(
    (_, value?: string) => {
      setState((state) => ({
        ...state,
        idleTimeoutSeconds: Number(value),
      }));
    },
    [setState]
  );

  return (
    <ScreenContent className={styles.root}>
      <ScreenTitle>
        <FormattedMessage id="SessionConfigurationScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="SessionConfigurationScreen.description" />
      </ScreenDescription>
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle>
          <FormattedMessage id="SessionConfigurationScreen.session-settings" />
        </WidgetTitle>
        <Toggle
          className={styles.control}
          inlineLabel={true}
          label={renderToString(
            "SessionConfigurationScreen.persistent-cookie.label"
          )}
          checked={state.persistentCookie}
          onChange={onPersistentCookieChange}
        />
        <TextField
          className={styles.control}
          type="number"
          min="1"
          step="1"
          label={renderToString(
            "SessionConfigurationScreen.session-lifetime.label"
          )}
          value={String(state.sessionLifetimeSeconds)}
          onChange={onSessionLifetimeSecondsChange}
        />
        <Toggle
          className={styles.control}
          inlineLabel={true}
          label={renderToString(
            "SessionConfigurationScreen.invalidate-session-after-idling.label"
          )}
          checked={state.idleTimeoutEnabled}
          onChange={onIdleTimeoutEnabledChange}
        />
        <TextField
          className={styles.control}
          type="number"
          min="1"
          step="1"
          disabled={!state.idleTimeoutEnabled}
          label={renderToString(
            "SessionConfigurationScreen.idle-timeout.label"
          )}
          value={String(state.idleTimeoutSeconds)}
          onChange={onIdleTimeoutSecondsChange}
        />
      </Widget>
    </ScreenContent>
  );
};

const SessionConfigurationScreen: React.FC = function SessionConfigurationScreen() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <SessionConfigurationScreenContent form={form} />
    </FormContainer>
  );
};

export default SessionConfigurationScreen;
