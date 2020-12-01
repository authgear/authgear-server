import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { TextField, Toggle } from "@fluentui/react";
import cn from "classnames";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";

import styles from "./SessionSettings.module.scss";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

interface FormState {
  persistentCookie: boolean;
  sessionLifetimeSeconds: number;
  idleTimeoutEnabled: boolean;
  idleTimeoutSeconds: number;
}

const emptyFormState: FormState = {
  persistentCookie: false,
  sessionLifetimeSeconds: 0,
  idleTimeoutEnabled: false,
  idleTimeoutSeconds: 0,
};

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    persistentCookie: !(config.session?.cookie_non_persistent ?? true),
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

const SessionSettingsContent: React.FC<HooksSettingsContentProps> = function SessionSettingsContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      { to: ".", label: <FormattedMessage id="SessionSettings.title" /> },
    ];
  }, []);

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
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <section className={styles.section}>
        <Toggle
          inlineLabel={true}
          label={renderToString("SessionSettings.persistent-cookie.label")}
          checked={state.persistentCookie}
          onChange={onPersistentCookieChange}
        />
      </section>
      <section className={styles.section}>
        <TextField
          className={styles.textField}
          type="number"
          min="1"
          step="1"
          label={renderToString("SessionSettings.session-lifetime.label")}
          value={String(state.sessionLifetimeSeconds)}
          onChange={onSessionLifetimeSecondsChange}
        />
      </section>
      <section className={styles.section}>
        <Toggle
          inlineLabel={true}
          label={renderToString(
            "SessionSettings.invalidate-session-after-idling.label"
          )}
          checked={state.idleTimeoutEnabled}
          onChange={onIdleTimeoutEnabledChange}
        />
        <TextField
          className={cn(styles.textField, styles.toggleContent)}
          type="number"
          min="1"
          step="1"
          disabled={!state.idleTimeoutEnabled}
          label={renderToString("SessionSettings.idle-timeout.label")}
          value={String(state.idleTimeoutSeconds)}
          onChange={onIdleTimeoutSecondsChange}
        />
      </section>
    </div>
  );
};

const SessionSettings: React.FC = function SessionSettings() {
  const { appID } = useParams();
  const form = useAppConfigForm(
    appID,
    emptyFormState,
    constructFormState,
    constructConfig
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <SessionSettingsContent form={form} />
    </FormContainer>
  );
};

export default SessionSettings;
