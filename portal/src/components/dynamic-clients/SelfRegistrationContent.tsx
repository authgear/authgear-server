import React, { useCallback, useMemo, useRef, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { produce } from "immer";
import { FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { AppConfigFormModel } from "../../hook/useAppConfigForm";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { Toggle } from "../v2/Toggle/Toggle";
import { TextField } from "../v2/TextField/TextField";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { SaveFunctionBar } from "../v2/SaveFunctionBar/SaveFunctionBar";
import { useCalloutToast } from "../v2/Callout/Callout";
import { InitialAccessTokenSection } from "./InitialAccessTokenSection";
import { useDynamicClientsQueryQuery } from "../../graphql/adminapi/query/dynamicClientsQuery.generated";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./SelfRegistrationContent.module.css";

// A save the admin triggered themselves needs only a glance to confirm.
const SAVED_TOAST_DURATION_MS = 2000;

// JSON pointer of the object holding the default_client_config fields. Passing
// it (with fieldName) lets the config schema's validation errors -- e.g.
// "minimum" when a lifetime is 0 or negative -- bind to the field that caused
// them instead of only reaching the generic error bar.
const DEFAULT_CLIENT_CONFIG_JSON_POINTER =
  "/oauth/dynamic_client_registration/default_client_config";

// Only this mechanism's fields live here. The tab that renders this owns the
// form, so its save bar commits DCR and nothing else -- not the sibling CIMD
// tab's settings, and not the client list.
export interface FormState {
  dynamicClientRegistrationEnabled: boolean;
  initialAccessTokenRequired: boolean;
  accessTokenLifetimeSeconds: number | undefined;
  refreshTokenLifetimeSeconds: number | undefined;
  refreshTokenIdleTimeoutEnabled: boolean;
  refreshTokenIdleTimeoutSeconds: number | undefined;
}

export function constructFormState(config: PortalAPIAppConfig): FormState {
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

export function constructConfig(
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

export interface SelfRegistrationContentProps {
  form: AppConfigFormModel<FormState>;
}

export const SelfRegistrationContent: React.VFC<SelfRegistrationContentProps> =
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

    // The SAVED value gates the cards below, never the pending one: the
    // initial-access-token controls act through the Admin API the moment
    // they are used, while POST /oauth2/register checks the saved config
    // (handler_register.go), so a card reachable before its save landed
    // could mint a token whose curl example is refused. Turning registration
    // on is written immediately, so the cards still appear at once; turning
    // it off leaves them up until the admin saves, which is accurate --
    // registration really is still running until then.
    const savedRegistrationEnabled =
      effectiveConfig.oauth?.dynamic_client_registration?.enabled ?? false;
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

    // The two directions are deliberately asymmetric, because only one of
    // them is dangerous to defer.
    //
    // ON is written immediately. The initial-access-token controls it reveals
    // act through the Admin API the moment they are used, while
    // POST /oauth2/register checks the SAVED config (handler_register.go: 403
    // access_denied when disabled) -- a token created while this switch was
    // merely pending came with a curl example that could not work. saveOnly
    // writes the switch and nothing else, so an edit pending elsewhere on the
    // page is neither committed unreviewed nor able to fail the save.
    //
    // OFF is deferred like every other control here. Nothing on this page
    // acts ahead of a save in that direction, so there is no hazard to close
    // -- and deferring means the switch never has to discard an unfinished
    // edit to get itself written. The admin reviews both together and presses
    // Save once.
    const setRegistrationEnabled = useCallback(
      (checked: boolean) => {
        if (!checked) {
          setState((prev) => ({
            ...prev,
            dynamicClientRegistrationEnabled: false,
            // Switching registration off also resets the initial access token
            // requirement, so re-enabling always starts from the safe default
            // rather than quietly restoring open registration. Turning the
            // requirement off again is an explicit, separately confirmed act.
            initialAccessTokenRequired: true,
          }));
          return;
        }
        form
          .saveOnly((prev) => ({
            ...prev,
            dynamicClientRegistrationEnabled: true,
          }))
          .then(() => {
            showToast({
              type: "success",
              text: (
                <FormattedMessage id="SelfRegistrationContent.enable.toast.on" />
              ),
              duration: SAVED_TOAST_DURATION_MS,
            });
          })
          // performSave rethrows, and the form's error bar renders it.
          .catch(() => {});
      },
      [form, setState, showToast]
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
      <>
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
              <FormattedMessage id="SelfRegistrationContent.enable.title" />
            }
          >
            <Text as="p" size="2" color="gray">
              <FormattedMessage
                id="SelfRegistrationContent.enable.description"
                values={{
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
              checked={state.dynamicClientRegistrationEnabled}
              disabled={isUpdating}
              onCheckedChange={onEnabledChange}
              text={
                <FormattedMessage id="SelfRegistrationContent.enable.toggle.label" />
              }
            />
          </SettingsSectionCard>

          {savedRegistrationEnabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={
                <FormattedMessage id="SelfRegistrationContent.registration-endpoint.label" />
              }
              description={
                <FormattedMessage id="SelfRegistrationContent.registration-endpoint.description" />
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

          {savedRegistrationEnabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={
                <FormattedMessage id="SelfRegistrationContent.security.title" />
              }
            >
              <div className="flex flex-col gap-1">
                <Toggle
                  checked={state.initialAccessTokenRequired}
                  onCheckedChange={onInitialAccessTokenRequiredChange}
                  text={
                    <FormattedMessage id="SelfRegistrationContent.iat-required.toggle.label" />
                  }
                />
                <Text as="p" size="1" color="gray">
                  <FormattedMessage id="SelfRegistrationContent.iat-required.toggle.description" />
                </Text>
              </div>
              <InitialAccessTokenSection
                registrationEndpoint={registrationEndpoint}
              />
            </SettingsSectionCard>
          ) : null}

          {savedRegistrationEnabled ? (
            <SettingsSectionCard
              contentClassName="gap-4"
              title={<FormattedMessage id="DynamicClientConfig.title" />}
              description={
                <FormattedMessage id="SelfRegistrationContent.client-config.description" />
              }
            >
              <TextField
                size="2"
                labelSize="2"
                type="text"
                label={
                  <FormattedMessage id="DynamicClientConfig.access-token-lifetime.label" />
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
                  <FormattedMessage id="DynamicClientConfig.refresh-token-lifetime.label" />
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
                  <FormattedMessage id="DynamicClientConfig.refresh-token-idle-timeout.label" />
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
              <FormattedMessage id="SelfRegistrationContent.disable.confirm.title" />
            }
            description={
              <FormattedMessage id="SelfRegistrationContent.disable.confirm.description" />
            }
            confirmText={
              <FormattedMessage id="SelfRegistrationContent.disable.confirm.confirm" />
            }
            cancelText={<FormattedMessage id="cancel" />}
            onConfirm={onConfirmDisable}
            onCancel={onCancelDisable}
          />

          <ConfirmationDialog
            open={isOpenRegistrationConfirmationVisible}
            onOpenChange={setIsOpenRegistrationConfirmationVisible}
            title={
              <FormattedMessage id="SelfRegistrationContent.open-registration.confirm.title" />
            }
            description={
              <FormattedMessage id="SelfRegistrationContent.open-registration.confirm.description" />
            }
            confirmText={
              <FormattedMessage id="SelfRegistrationContent.open-registration.confirm.confirm" />
            }
            cancelText={
              <FormattedMessage id="SelfRegistrationContent.open-registration.confirm.cancel" />
            }
            onConfirm={onConfirmOpenRegistration}
            onCancel={onCancelOpenRegistration}
          />

          <SaveFunctionBar anchorRef={anchorRef} />
        </div>
      </>
    );
  };
