import React, { useCallback, useContext, useMemo } from "react";
import { Text } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";
import { produce } from "immer";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import FormContainer from "../../FormContainer";

import styles from "./CookieLifetimeConfigurationScreen.module.css";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import Widget from "../../Widget";
import Toggle from "../../Toggle";
import TextField from "../../TextField";

function getHostname(publicOrigin: string): string {
  try {
    return new URL(publicOrigin).hostname;
  } catch (_: unknown) {
    return "";
  }
}

interface FormState {
  publicOrigin: string;
  sessionLifetimeSeconds: number | undefined;
  idleTimeoutEnabled: boolean;
  idleTimeoutSeconds: number | undefined;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
    sessionLifetimeSeconds: config.session?.lifetime_seconds,
    idleTimeoutEnabled: config.session?.idle_timeout_enabled ?? false,
    idleTimeoutSeconds: config.session?.idle_timeout_seconds,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.session = config.session ?? {};
    config.session.lifetime_seconds = currentState.sessionLifetimeSeconds;
    config.session.idle_timeout_enabled = currentState.idleTimeoutEnabled;
    config.session.idle_timeout_seconds = currentState.idleTimeoutSeconds;
    clearEmptyObject(config);
  });
}

interface SessionConfigurationWidgetProps {
  form: AppConfigFormModel<FormState>;
}

const SessionConfigurationWidget: React.VFC<SessionConfigurationWidgetProps> =
  function SessionConfigurationWidget(props: SessionConfigurationWidgetProps) {
    const { state, setState } = props.form;

    const { renderToString } = useContext(Context);

    const hostname = useMemo(
      () => getHostname(state.publicOrigin),
      [state.publicOrigin]
    );

    const onSessionLifetimeSecondsChange = useCallback(
      (_, value?: string) => {
        setState((prev) => ({
          ...prev,
          sessionLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
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
        setState((prev) => ({
          ...prev,
          idleTimeoutSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    return (
      <Widget className={styles.widget}>
        <TextField
          type="text"
          label={renderToString(
            "CookieLifetimeConfigurationScreen.session-lifetime.label"
          )}
          description={renderToString(
            "CookieLifetimeConfigurationScreen.session-lifetime.description",
            { hostname }
          )}
          value={state.sessionLifetimeSeconds?.toFixed(0) ?? ""}
          onChange={onSessionLifetimeSecondsChange}
        />
        <Toggle
          label={renderToString(
            "CookieLifetimeConfigurationScreen.invalidate-session-after-idling.label"
          )}
          description={renderToString(
            "CookieLifetimeConfigurationScreen.invalidate-session-after-idling.description"
          )}
          checked={state.idleTimeoutEnabled}
          onChange={onIdleTimeoutEnabledChange}
        />
        <TextField
          type="text"
          disabled={!state.idleTimeoutEnabled}
          label={renderToString(
            "CookieLifetimeConfigurationScreen.idle-timeout.label"
          )}
          description={renderToString(
            "CookieLifetimeConfigurationScreen.idle-timeout.description"
          )}
          value={state.idleTimeoutSeconds?.toFixed(0) ?? ""}
          onChange={onIdleTimeoutSecondsChange}
        />
      </Widget>
    );
  };

interface CookieLifetimeConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const CookieLifetimeConfigurationScreenContent: React.VFC<CookieLifetimeConfigurationScreenContentProps> =
  function CookieLifetimeConfigurationScreenContent(props) {
    const { form } = props;
    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="CookieLifetimeConfigurationScreen.title" />
        </ScreenTitle>
        <Widget className={styles.widget}>
          <Text>
            <FormattedMessage
              id="CookieLifetimeConfigurationScreen.description"
              values={{
                hostname: getHostname(form.state.publicOrigin),
              }}
            />
          </Text>
        </Widget>
        <SessionConfigurationWidget form={form} />
      </ScreenContent>
    );
  };

const CookieLifetimeConfigurationScreen: React.VFC =
  function CookieLifetimeConfigurationScreen() {
    const { appID } = useParams() as { appID: string };

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form}>
        <CookieLifetimeConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default CookieLifetimeConfigurationScreen;
