import React, { useCallback, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import PortalLink from "../../Link";
import type { FormState as ApplicationsFormState } from "../../graphql/portal/ApplicationsConfigurationScreen";
import { AppConfigFormModel } from "../../hook/useAppConfigForm";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { SettingsSectionCard } from "../v2/SettingsSectionCard/SettingsSectionCard";
import { Toggle } from "../v2/Toggle/Toggle";
import { TextField } from "../v2/TextField/TextField";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { SaveFunctionBar } from "../v2/SaveFunctionBar/SaveFunctionBar";
import { InitialAccessTokenSection } from "./InitialAccessTokenSection";
import { DynamicClientAllowedResources } from "./DynamicClientAllowedResources";
import { useDynamicClientsQueryQuery } from "../../graphql/adminapi/query/dynamicClientsQuery.generated";
import styles from "./DynamicClientsTab.module.css";

export interface DynamicClientsTabProps {
  form: AppConfigFormModel<ApplicationsFormState>;
  publicOrigin: string;
  // The smallest block-action quota configured for oauth_client_dcr, or null
  // when the plan does not limit dynamic clients.
  dcrClientQuota: number | null;
}

export const DynamicClientsTab: React.VFC<DynamicClientsTabProps> =
  function DynamicClientsTab({ form, publicOrigin, dcrClientQuota }) {
    const navigate = useNavigate();
    const { appID } = useParams() as { appID: string };
    const { state, setState, isUpdating, effectiveConfig } = form;
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const anchorRef = useRef<HTMLDivElement>(null);

    // The server resolves every default_client_config field via
    // OAuthDynamicClientRegistrationDefaultClientConfig.SetDefaults(), so the
    // effective config always carries a concrete value even when authgear.yaml
    // omits the whole section. Read the placeholders from there instead of
    // duplicating the defaults here, where they silently drift.
    const effectiveDefaultClientConfig =
      effectiveConfig.oauth?.dynamic_client_registration?.default_client_config;

    const [
      isOpenRegistrationConfirmationVisible,
      setIsOpenRegistrationConfirmationVisible,
    ] = useState(false);

    const registrationEnabled = state.dynamicClientRegistrationEnabled;
    const registrationEndpoint = `${publicOrigin}/oauth2/register`;

    const { data } = useDynamicClientsQueryQuery({
      variables: { first: 1 },
      fetchPolicy: "cache-and-network",
      skip: !registrationEnabled,
    });
    const totalCount = data?.dynamicClients?.totalCount ?? null;

    const onEnabledChange = useCallback(
      (checked: boolean) => {
        // Deferred like every other control on this tab: nothing is written
        // until the admin presses Save. Saving from here would commit the
        // whole form state, including edits elsewhere on the tab (and on the
        // sibling Applications tab, which shares this form) that the admin
        // has not confirmed yet.
        setState((prev) => ({
          ...prev,
          dynamicClientRegistrationEnabled: checked,
        }));
      },
      [setState]
    );

    const onViewAllClick = useCallback(() => {
      navigate("./dcr");
    }, [navigate]);

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
      <div
        ref={anchorRef}
        className={cn(styles.root, isDirty && styles.rootWithSaveBar)}
      >
        <SettingsSectionCard
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientsTab.section.registration.title" />
          }
        >
          <div className="flex flex-col gap-1">
            <Toggle
              checked={registrationEnabled}
              disabled={isUpdating}
              onCheckedChange={onEnabledChange}
              text={
                <FormattedMessage id="DynamicClientsTab.enable.toggle.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="DynamicClientsTab.enable.description" />
            </Text>
          </div>
          {registrationEnabled ? (
            <TextField
              size="2"
              labelSize="2"
              label={
                <FormattedMessage id="DynamicClientsTab.registration-endpoint.label" />
              }
              value={registrationEndpoint}
              readOnly={true}
              suffixPlain={true}
              suffix={<CopyIconButton textToCopy={registrationEndpoint} />}
            />
          ) : null}
        </SettingsSectionCard>

        {registrationEnabled ? (
          <SettingsSectionCard
            contentClassName="gap-4"
            title={<FormattedMessage id="DynamicClientsTab.clients.title" />}
          >
            {totalCount != null ? (
              <Text as="p" size="2">
                {dcrClientQuota != null ? (
                  <FormattedMessage
                    id="DynamicClientsTab.quota"
                    values={{ count: totalCount, quota: dcrClientQuota }}
                  />
                ) : (
                  <FormattedMessage
                    id="DynamicClientsTab.clients.count"
                    values={{ count: totalCount }}
                  />
                )}
              </Text>
            ) : null}
            <div className="self-start">
              <SecondaryButton
                size="2"
                text={
                  <FormattedMessage id="DynamicClientsTab.clients.view-all" />
                }
                onClick={onViewAllClick}
              />
            </div>
          </SettingsSectionCard>
        ) : null}

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
          <InitialAccessTokenSection />
        </SettingsSectionCard>

        <SettingsSectionCard
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientsTab.default-client-config.title" />
          }
        >
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="DynamicClientsTab.default-client-config.description" />
          </Text>
          <TextField
            size="2"
            labelSize="2"
            type="text"
            label={
              <FormattedMessage id="DynamicClientsTab.access-token-lifetime.label" />
            }
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
                <FormattedMessage id="DynamicClientsTab.refresh-token-idle-timeout-enabled.label" />
              }
            />
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="DynamicClientsTab.refresh-token-idle-timeout-enabled.description" />
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
            placeholder={effectiveDefaultClientConfig?.refresh_token_idle_timeout_seconds?.toFixed(
              0
            )}
            value={state.refreshTokenIdleTimeoutSeconds?.toFixed(0) ?? ""}
            onChange={onRefreshTokenIdleTimeoutChange}
          />
        </SettingsSectionCard>

        {registrationEnabled ? (
          <SettingsSectionCard
            contentClassName="gap-4"
            title={
              <FormattedMessage id="DynamicClientsTab.allowed-resources.title" />
            }
          >
            <Text as="p" size="2" color="gray">
              <FormattedMessage
                id="DynamicClientsTab.allowed-resources.description"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  apiResourcesLink: (chunks: React.ReactNode) => (
                    <PortalLink to={`/project/${appID}/api-resources`}>
                      {chunks}
                    </PortalLink>
                  ),
                }}
              />
            </Text>
            <DynamicClientAllowedResources />
          </SettingsSectionCard>
        ) : null}

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
    );
  };
