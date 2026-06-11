import React, { useCallback, useMemo, useRef } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
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
import { TextField } from "../../components/v2/TextField/TextField";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";
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

interface CookieLifetimeConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const CookieLifetimeConfigurationScreenContent: React.VFC<CookieLifetimeConfigurationScreenContentProps> =
  function CookieLifetimeConfigurationScreenContent(props) {
    const { form } = props;
    const { state, setState } = form;
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const hostname = useMemo(
      () => getHostname(state.publicOrigin),
      [state.publicOrigin]
    );

    const onSessionLifetimeSecondsChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          sessionLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onIdleTimeoutEnabledChange = useCallback(
      (checked: boolean) => {
        setState((state) => ({
          ...state,
          idleTimeoutEnabled: checked,
        }));
      },
      [setState]
    );

    const onIdleTimeoutSecondsChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          idleTimeoutSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="CookieLifetimeConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage
              id="CookieLifetimeConfigurationScreen.description"
              values={{
                hostname,
                // eslint-disable-next-line react/no-unstable-nested-components
                b: (chunks: React.ReactNode) => <b>{chunks}</b>,
              }}
            />
          </Text>
        </div>

        <div
          className={cn(
            styles.widget,
            "border border-[var(--gray-5)] rounded-lg p-6 flex gap-8 bg-white",
            isDirty && styles.settingsCardSaveBarClearance
          )}
        >
          <Text as="p" size="3" weight="medium" className="shrink-0 w-[200px]">
            <FormattedMessage id="CookieLifetimeConfigurationScreen.settings.label" />
          </Text>
          <div className="flex-1 flex flex-col gap-4 min-w-0">
          <TextField
            size="2"
            labelSize="2"
            type="text"
            label={
              <FormattedMessage id="CookieLifetimeConfigurationScreen.session-lifetime.label" />
            }
            hint={
              <FormattedMessage
                id="CookieLifetimeConfigurationScreen.session-lifetime.description"
                values={{ hostname }}
              />
            }
            value={state.sessionLifetimeSeconds?.toFixed(0) ?? ""}
            onChange={onSessionLifetimeSecondsChange}
          />
          <div className="flex flex-col gap-1">
            <Toggle
              checked={state.idleTimeoutEnabled}
              onCheckedChange={onIdleTimeoutEnabledChange}
              text={
                <FormattedMessage id="CookieLifetimeConfigurationScreen.invalidate-session-after-idling.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="CookieLifetimeConfigurationScreen.invalidate-session-after-idling.description" />
            </Text>
          </div>
          <TextField
            size="2"
            labelSize="2"
            type="text"
            disabled={!state.idleTimeoutEnabled}
            label={
              <FormattedMessage id="CookieLifetimeConfigurationScreen.idle-timeout.label" />
            }
            hint={
              <FormattedMessage id="CookieLifetimeConfigurationScreen.idle-timeout.description" />
            }
            value={state.idleTimeoutSeconds?.toFixed(0) ?? ""}
            onChange={onIdleTimeoutSecondsChange}
          />
          </div>
        </div>

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
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
      <FormContainer form={form} hideFooterComponent={true}>
        <CookieLifetimeConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default CookieLifetimeConfigurationScreen;
