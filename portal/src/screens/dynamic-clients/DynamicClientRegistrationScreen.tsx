import React, { useCallback, useMemo, useRef, useState } from "react";
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
import ScreenContent from "../../ScreenContent";
import { TextField } from "../../components/v2/TextField/TextField";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { InitialAccessTokenSection } from "../../components/dynamic-clients/InitialAccessTokenSection";
import { useFormContainerBaseContext } from "../../FormContainerBase";

import styles from "./DynamicClientRegistrationScreen.module.css";

// Effective defaults applied by the server when the corresponding
// default_client_config field is absent. Mirrors
// OAuthDynamicClientRegistrationDefaultClientConfig.SetDefaults() in
// pkg/lib/config/oauth_dynamic_client_registration.go.
const DEFAULT_ACCESS_TOKEN_LIFETIME_SECONDS = 1800;
const DEFAULT_REFRESH_TOKEN_LIFETIME_SECONDS = 2592000;
const DEFAULT_REFRESH_TOKEN_IDLE_TIMEOUT_SECONDS = 1209600;

interface FormState {
  enabled: boolean;
  initialAccessTokenRequired: boolean;
  accessTokenLifetimeSeconds: number | undefined;
  refreshTokenLifetimeSeconds: number | undefined;
  refreshTokenIdleTimeoutEnabled: boolean;
  refreshTokenIdleTimeoutSeconds: number | undefined;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const dcr = config.oauth?.dynamic_client_registration;
  return {
    enabled: dcr?.enabled ?? false,
    // Absent means required — the spec default.
    initialAccessTokenRequired: dcr?.initial_access_token_required ?? true,
    accessTokenLifetimeSeconds:
      dcr?.default_client_config?.access_token_lifetime_seconds,
    refreshTokenLifetimeSeconds:
      dcr?.default_client_config?.refresh_token_lifetime_seconds,
    refreshTokenIdleTimeoutEnabled:
      dcr?.default_client_config?.refresh_token_idle_timeout_enabled ?? true,
    refreshTokenIdleTimeoutSeconds:
      dcr?.default_client_config?.refresh_token_idle_timeout_seconds,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (draft) => {
    draft.oauth ??= {};
    draft.oauth.dynamic_client_registration ??= {};
    const dcr = draft.oauth.dynamic_client_registration;

    if (currentState.enabled) {
      dcr.enabled = true;
    } else {
      delete dcr.enabled;
    }

    if (currentState.initialAccessTokenRequired) {
      // Absent means required — keep the config minimal.
      delete dcr.initial_access_token_required;
    } else {
      dcr.initial_access_token_required = false;
    }

    dcr.default_client_config ??= {};
    const defaultClientConfig = dcr.default_client_config;

    if (currentState.accessTokenLifetimeSeconds != null) {
      defaultClientConfig.access_token_lifetime_seconds =
        currentState.accessTokenLifetimeSeconds;
    } else {
      delete defaultClientConfig.access_token_lifetime_seconds;
    }

    if (currentState.refreshTokenLifetimeSeconds != null) {
      defaultClientConfig.refresh_token_lifetime_seconds =
        currentState.refreshTokenLifetimeSeconds;
    } else {
      delete defaultClientConfig.refresh_token_lifetime_seconds;
    }

    if (currentState.refreshTokenIdleTimeoutEnabled) {
      // Absent means enabled — the server default.
      delete defaultClientConfig.refresh_token_idle_timeout_enabled;
    } else {
      defaultClientConfig.refresh_token_idle_timeout_enabled = false;
    }

    if (currentState.refreshTokenIdleTimeoutSeconds != null) {
      defaultClientConfig.refresh_token_idle_timeout_seconds =
        currentState.refreshTokenIdleTimeoutSeconds;
    } else {
      delete defaultClientConfig.refresh_token_idle_timeout_seconds;
    }

    clearEmptyObject(draft);
  });
}

interface DynamicClientRegistrationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const DynamicClientRegistrationScreenContent: React.VFC<DynamicClientRegistrationScreenContentProps> =
  function DynamicClientRegistrationScreenContent(props) {
    const { form } = props;
    const { state, setState } = form;
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const [
      isOpenRegistrationConfirmationVisible,
      setIsOpenRegistrationConfirmationVisible,
    ] = useState(false);

    const onEnabledChange = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          enabled: checked,
        }));
      },
      [setState]
    );

    const onInitialAccessTokenRequiredChange = useCallback(
      (checked: boolean) => {
        if (!checked) {
          // Turning the requirement off enables open registration — confirm
          // before reflecting it in the form state.
          setIsOpenRegistrationConfirmationVisible(true);
          return;
        }
        setState((prev) => ({
          ...prev,
          initialAccessTokenRequired: true,
        }));
      },
      [setState]
    );

    const onConfirmOpenRegistration = useCallback(() => {
      setIsOpenRegistrationConfirmationVisible(false);
      setState((prev) => ({
        ...prev,
        initialAccessTokenRequired: false,
      }));
    }, [setState]);

    const onCancelOpenRegistration = useCallback(() => {
      setIsOpenRegistrationConfirmationVisible(false);
    }, []);

    const onAccessTokenLifetimeChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          accessTokenLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onRefreshTokenLifetimeChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          refreshTokenLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onRefreshTokenIdleTimeoutEnabledChange = useCallback(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          refreshTokenIdleTimeoutEnabled: checked,
        }));
      },
      [setState]
    );

    const onRefreshTokenIdleTimeoutChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((prev) => ({
          ...prev,
          refreshTokenIdleTimeoutSeconds: parseIntegerAllowLeadingZeros(value),
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
            <FormattedMessage id="DynamicClientRegistrationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="DynamicClientRegistrationScreen.description" />
          </Text>
        </div>

        <SettingsSectionCard
          className={styles.widget}
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientRegistrationScreen.enable.title" />
          }
        >
          <div className="flex flex-col gap-1">
            <Toggle
              checked={state.enabled}
              onCheckedChange={onEnabledChange}
              text={
                <FormattedMessage id="DynamicClientRegistrationScreen.enable.toggle.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="DynamicClientRegistrationScreen.enable.toggle.description" />
            </Text>
          </div>
        </SettingsSectionCard>

        <SettingsSectionCard
          className={styles.widget}
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientRegistrationScreen.security.title" />
          }
        >
          <div className="flex flex-col gap-1">
            <Toggle
              checked={state.initialAccessTokenRequired}
              onCheckedChange={onInitialAccessTokenRequiredChange}
              text={
                <FormattedMessage id="DynamicClientRegistrationScreen.iat-required.toggle.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="DynamicClientRegistrationScreen.iat-required.toggle.description" />
            </Text>
          </div>
          <InitialAccessTokenSection />
        </SettingsSectionCard>

        <SettingsSectionCard
          className={cn(
            styles.widget,
            isDirty && styles.settingsCardSaveBarClearance
          )}
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientRegistrationScreen.default-client-config.title" />
          }
        >
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="DynamicClientRegistrationScreen.default-client-config.description" />
          </Text>
          <TextField
            size="2"
            labelSize="2"
            type="text"
            label={
              <FormattedMessage id="DynamicClientRegistrationScreen.access-token-lifetime.label" />
            }
            placeholder={DEFAULT_ACCESS_TOKEN_LIFETIME_SECONDS.toFixed(0)}
            value={state.accessTokenLifetimeSeconds?.toFixed(0) ?? ""}
            onChange={onAccessTokenLifetimeChange}
          />
          <TextField
            size="2"
            labelSize="2"
            type="text"
            label={
              <FormattedMessage id="DynamicClientRegistrationScreen.refresh-token-lifetime.label" />
            }
            placeholder={DEFAULT_REFRESH_TOKEN_LIFETIME_SECONDS.toFixed(0)}
            value={state.refreshTokenLifetimeSeconds?.toFixed(0) ?? ""}
            onChange={onRefreshTokenLifetimeChange}
          />
          <div className="flex flex-col gap-1">
            <Toggle
              checked={state.refreshTokenIdleTimeoutEnabled}
              onCheckedChange={onRefreshTokenIdleTimeoutEnabledChange}
              text={
                <FormattedMessage id="DynamicClientRegistrationScreen.refresh-token-idle-timeout-enabled.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="DynamicClientRegistrationScreen.refresh-token-idle-timeout-enabled.description" />
            </Text>
          </div>
          <TextField
            size="2"
            labelSize="2"
            type="text"
            disabled={!state.refreshTokenIdleTimeoutEnabled}
            label={
              <FormattedMessage id="DynamicClientRegistrationScreen.refresh-token-idle-timeout.label" />
            }
            placeholder={DEFAULT_REFRESH_TOKEN_IDLE_TIMEOUT_SECONDS.toFixed(0)}
            value={state.refreshTokenIdleTimeoutSeconds?.toFixed(0) ?? ""}
            onChange={onRefreshTokenIdleTimeoutChange}
          />
        </SettingsSectionCard>

        <ConfirmationDialog
          open={isOpenRegistrationConfirmationVisible}
          onOpenChange={setIsOpenRegistrationConfirmationVisible}
          title={
            <FormattedMessage id="DynamicClientRegistrationScreen.open-registration.confirm.title" />
          }
          description={
            <FormattedMessage id="DynamicClientRegistrationScreen.open-registration.confirm.description" />
          }
          confirmText={
            <FormattedMessage id="DynamicClientRegistrationScreen.open-registration.confirm.confirm" />
          }
          cancelText={
            <FormattedMessage id="DynamicClientRegistrationScreen.open-registration.confirm.cancel" />
          }
          onConfirm={onConfirmOpenRegistration}
          onCancel={onCancelOpenRegistration}
        />

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const DynamicClientRegistrationScreen: React.VFC =
  function DynamicClientRegistrationScreen() {
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
        <DynamicClientRegistrationScreenContent form={form} />
      </FormContainer>
    );
  };

export default DynamicClientRegistrationScreen;
