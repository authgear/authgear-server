import React, { useCallback, useMemo, useRef, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import FormContainer from "../../FormContainer";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { TextField } from "../../components/v2/TextField/TextField";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useCalloutToast } from "../../components/v2/Callout/Callout";
import { InitialAccessTokenSection } from "../../components/dynamic-clients/InitialAccessTokenSection";
import { useDynamicClientsQueryQuery } from "../../graphql/adminapi/query/dynamicClientsQuery.generated";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./SelfRegistrationScreen.module.css";

// JSON pointer of the object holding the default_client_config fields. Passing
// it (with fieldName) lets the config schema's validation errors -- e.g.
// "minimum" when a lifetime is 0 or negative -- bind to the field that caused
// them instead of only reaching the generic error bar.
// A save the admin triggered themselves needs only a glance to confirm.
const SAVED_TOAST_DURATION_MS = 2000;

const DEFAULT_CLIENT_CONFIG_JSON_POINTER =
  "/oauth/dynamic_client_registration/default_client_config";

// Only this mechanism's fields live here. The form is this screen's own, so
// its save bar commits DCR and nothing else -- not the sibling CIMD screen's
// settings, and not the Applications screen's client list.
interface FormState {
  dynamicClientRegistrationEnabled: boolean;
  initialAccessTokenRequired: boolean;
  accessTokenLifetimeSeconds: number | undefined;
  refreshTokenLifetimeSeconds: number | undefined;
  refreshTokenIdleTimeoutEnabled: boolean;
  refreshTokenIdleTimeoutSeconds: number | undefined;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const dcr = config.oauth?.dynamic_client_registration;
  return {
    dynamicClientRegistrationEnabled: dcr?.enabled ?? false,
    // Absent means required — the spec default. The requirement only means
    // anything while registration is enabled, so normalise it back to required
    // whenever registration is off: that way enabling registration can never
    // silently open it, including for a config that arrived with the
    // requirement already turned off. constructConfig then drops the key.
    initialAccessTokenRequired:
      dcr?.enabled ?? false ? dcr?.initial_access_token_required ?? true : true,
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
  const [newConfig, _] = produce(
    [config, currentState],
    ([config, currentState]) => {
      config.oauth ??= {};
      config.oauth.dynamic_client_registration ??= {};
      const dcr = config.oauth.dynamic_client_registration;

      if (currentState.dynamicClientRegistrationEnabled) {
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

      clearEmptyObject(config);
    }
  );
  return newConfig;
}

interface SelfRegistrationContentProps {
  form: AppConfigFormModel<FormState>;
}

const SelfRegistrationContent: React.VFC<SelfRegistrationContentProps> =
  function SelfRegistrationContent({ form }) {
    const { state, setState, isUpdating, effectiveConfig } = form;
    const { showToast } = useCalloutToast();
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const anchorRef = useRef<HTMLDivElement>(null);

    // The server resolves every default_client_config field via
    // OAuthDynamicClientTokenLifetimesConfig.SetDefaults(), so the
    // effective config always carries a concrete value even when authgear.yaml
    // omits the whole section. Read the placeholders from there instead of
    // duplicating the defaults here, where they silently drift.
    const effectiveDefaultClientConfig =
      effectiveConfig.oauth?.dynamic_client_registration?.default_client_config;

    const [
      isOpenRegistrationConfirmationVisible,
      setIsOpenRegistrationConfirmationVisible,
    ] = useState(false);
    const [isDisableConfirmationVisible, setIsDisableConfirmationVisible] =
      useState(false);

    const registrationEnabled = state.dynamicClientRegistrationEnabled;
    const publicOrigin = effectiveConfig.http?.public_origin ?? "";
    const registrationEndpoint = `${publicOrigin}/oauth2/register`;

    // The count gates the disable dialog: it warns that registered clients
    // keep working, which is only worth saying when some exist. first: 1
    // because only totalCount is read.
    const { data: dcrData } = useDynamicClientsQueryQuery({
      variables: { first: 1, source: OAuthClientSource.Dcr },
      fetchPolicy: "cache-and-network",
    });
    const hasDCRClients = (dcrData?.dynamicClients?.totalCount ?? 0) > 0;

    // Saved immediately, unlike every other control on this screen. The
    // initial-access-token controls below act through the Admin API the moment
    // they are used, while POST /oauth2/register checks the SAVED config
    // (handler_register.go: 403 access_denied when disabled) -- so a token
    // created while this switch was merely pending came with a curl example
    // that could not work. This screen's form covers only
    // oauth.dynamic_client_registration, so the save cannot carry an edit
    // belonging to another screen; it does commit anything else pending here,
    // which is what the toast reports.
    const setRegistrationEnabled = useCallback(
      (checked: boolean) => {
        form
          .saveWith((prev) => ({
            ...prev,
            dynamicClientRegistrationEnabled: checked,
            // Switching registration off resets the initial access token
            // requirement, so re-enabling always starts from the safe default
            // rather than quietly restoring open registration. Turning the
            // requirement off again is an explicit, separately confirmed act.
            initialAccessTokenRequired: checked
              ? prev.initialAccessTokenRequired
              : true,
          }))
          .then(() => {
            showToast({
              type: "success",
              text: <FormattedMessage id="changes-saved" />,
              duration: SAVED_TOAST_DURATION_MS,
            });
          })
          // performSave rethrows, and the form's error bar renders it.
          .catch(() => {});
      },
      [form, showToast]
    );

    const onEnabledChange = useCallback(
      (checked: boolean) => {
        // Turning registration off does not disable already-registered
        // clients — remind the admin of that before reflecting it in the
        // form state. The reminder is only true while clients exist, so
        // skip it otherwise.
        if (!checked && hasDCRClients) {
          setIsDisableConfirmationVisible(true);
          return;
        }
        setRegistrationEnabled(checked);
      },
      [hasDCRClients, setRegistrationEnabled]
    );

    const onConfirmDisable = useCallback(() => {
      setIsDisableConfirmationVisible(false);
      setRegistrationEnabled(false);
    }, [setRegistrationEnabled]);

    const onCancelDisable = useCallback(() => {
      setIsDisableConfirmationVisible(false);
    }, []);

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
      <ScreenContent layout="list">
        <div className={cn(styles.widget, styles.pageHeader)}>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="SelfRegistrationScreen.title" />
          </Text>
        </div>
        <div
          ref={anchorRef}
          className={cn(
            styles.widget,
            styles.content,
            isDirty && styles.contentWithSaveBar
          )}
        >
          <SettingsSectionCard
            contentClassName="gap-4"
            title={
              <FormattedMessage id="SelfRegistrationScreen.enable.title" />
            }
          >
            <Text as="p" size="2" color="gray">
              <FormattedMessage
                id="DynamicClientsTab.enable.description"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  mcpLink: (chunks: React.ReactNode) => (
                    <ExternalLink href="https://docs.authgear.com/get-started/auth-for-mcp">
                      {chunks}
                    </ExternalLink>
                  ),
                  // eslint-disable-next-line react/no-unstable-nested-components
                  dcrLink: (chunks: React.ReactNode) => (
                    <ExternalLink href="https://docs.authgear.com/integration/dynamic-client-registration">
                      {chunks}
                    </ExternalLink>
                  ),
                }}
              />
            </Text>
            <Toggle
              checked={registrationEnabled}
              disabled={isUpdating}
              onCheckedChange={onEnabledChange}
              text={
                <FormattedMessage id="DynamicClientsTab.enable.toggle.label" />
              }
            />
          </SettingsSectionCard>

          {registrationEnabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={
                <FormattedMessage id="DynamicClientsTab.registration-endpoint.label" />
              }
              description={
                <FormattedMessage id="DynamicClientsTab.registration-endpoint.description" />
              }
            >
              <TextField
                size="2"
                labelSize="2"
                value={registrationEndpoint}
                readOnly={true}
                suffixPlain={true}
                suffix={<CopyIconButton textToCopy={registrationEndpoint} />}
              />
            </SettingsSectionCard>
          ) : null}

          {registrationEnabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={<FormattedMessage id="DynamicClientsTab.security.title" />}
            >
              <div className="flex flex-col gap-1">
                <Toggle
                  checked={state.initialAccessTokenRequired}
                  onCheckedChange={onInitialAccessTokenRequiredChange}
                  text={
                    <FormattedMessage id="DynamicClientsTab.iat-required.toggle.label" />
                  }
                />
                <Text as="p" size="1" color="gray">
                  <FormattedMessage id="DynamicClientsTab.iat-required.toggle.description" />
                </Text>
              </div>
              <InitialAccessTokenSection
                registrationEndpoint={registrationEndpoint}
              />
            </SettingsSectionCard>
          ) : null}

          {registrationEnabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={
                <FormattedMessage id="DynamicClientsTab.client-config.title" />
              }
              description={
                <FormattedMessage id="DynamicClientsTab.client-config.description" />
              }
            >
              <TextField
                size="2"
                labelSize="2"
                type="text"
                label={
                  <FormattedMessage id="DynamicClientsTab.access-token-lifetime.label" />
                }
                parentJSONPointer={DEFAULT_CLIENT_CONFIG_JSON_POINTER}
                fieldName="access_token_lifetime_seconds"
                placeholder={effectiveDefaultClientConfig?.access_token_lifetime_seconds?.toFixed(
                  0
                )}
                value={state.accessTokenLifetimeSeconds?.toFixed(0) ?? ""}
                onChange={onAccessTokenLifetimeChange}
              />
              <TextField
                size="2"
                labelSize="2"
                type="text"
                label={
                  <FormattedMessage id="DynamicClientsTab.refresh-token-lifetime.label" />
                }
                parentJSONPointer={DEFAULT_CLIENT_CONFIG_JSON_POINTER}
                fieldName="refresh_token_lifetime_seconds"
                placeholder={effectiveDefaultClientConfig?.refresh_token_lifetime_seconds?.toFixed(
                  0
                )}
                value={state.refreshTokenLifetimeSeconds?.toFixed(0) ?? ""}
                onChange={onRefreshTokenLifetimeChange}
              />
              <div className="flex flex-col gap-1">
                <Toggle
                  checked={state.refreshTokenIdleTimeoutEnabled}
                  onCheckedChange={onRefreshTokenIdleTimeoutEnabledChange}
                  text={
                    <FormattedMessage id="EditOAuthClientForm.refresh-token-idle-timeout-enabled.label" />
                  }
                />
                <Text as="p" size="1" color="gray">
                  <FormattedMessage id="EditOAuthClientForm.refresh-token-idle-timeout-enabled.description" />
                </Text>
              </div>
              <TextField
                size="2"
                labelSize="2"
                type="text"
                disabled={!state.refreshTokenIdleTimeoutEnabled}
                label={
                  <FormattedMessage id="DynamicClientsTab.refresh-token-idle-timeout.label" />
                }
                parentJSONPointer={DEFAULT_CLIENT_CONFIG_JSON_POINTER}
                fieldName="refresh_token_idle_timeout_seconds"
                placeholder={effectiveDefaultClientConfig?.refresh_token_idle_timeout_seconds?.toFixed(
                  0
                )}
                value={state.refreshTokenIdleTimeoutSeconds?.toFixed(0) ?? ""}
                onChange={onRefreshTokenIdleTimeoutChange}
              />
            </SettingsSectionCard>
          ) : null}

          <ConfirmationDialog
            open={isDisableConfirmationVisible}
            onOpenChange={setIsDisableConfirmationVisible}
            title={
              <FormattedMessage id="DynamicClientsTab.disable.confirm.title" />
            }
            description={
              <FormattedMessage id="DynamicClientsTab.disable.confirm.description" />
            }
            confirmText={
              <FormattedMessage id="DynamicClientsTab.disable.confirm.confirm" />
            }
            cancelText={<FormattedMessage id="cancel" />}
            onConfirm={onConfirmDisable}
            onCancel={onCancelDisable}
          />

          <ConfirmationDialog
            open={isOpenRegistrationConfirmationVisible}
            onOpenChange={setIsOpenRegistrationConfirmationVisible}
            title={
              <FormattedMessage id="DynamicClientsTab.open-registration.confirm.title" />
            }
            description={
              <FormattedMessage id="DynamicClientsTab.open-registration.confirm.description" />
            }
            confirmText={
              <FormattedMessage id="DynamicClientsTab.open-registration.confirm.confirm" />
            }
            cancelText={
              <FormattedMessage id="DynamicClientsTab.open-registration.confirm.cancel" />
            }
            onConfirm={onConfirmOpenRegistration}
            onCancel={onCancelOpenRegistration}
          />

          <SaveFunctionBar anchorRef={anchorRef} />
        </div>
      </ScreenContent>
    );
  };

const SelfRegistrationScreen: React.VFC = function SelfRegistrationScreen() {
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
      <SelfRegistrationContent form={form} />
    </FormContainer>
  );
};

export default SelfRegistrationScreen;
