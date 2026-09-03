import React, { useCallback, useMemo, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import PortalLink from "../../Link";
import ExternalLink from "../../ExternalLink";
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
import { CIMDSection } from "./CIMDSection";
import { MechanismSectionHeader } from "./MechanismSectionHeader";
import { useDynamicClientsQueryQuery } from "../../graphql/adminapi/query/dynamicClientsQuery.generated";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";
import styles from "./DynamicClientsTab.module.css";

// JSON pointer of the object holding the default_client_config fields. Passing
// it (with fieldName) lets the config schema's validation errors -- e.g.
// "minimum" when a lifetime is 0 or negative -- bind to the field that caused
// them instead of only reaching the generic error bar.
const DEFAULT_CLIENT_CONFIG_JSON_POINTER =
  "/oauth/dynamic_client_registration/default_client_config";

interface ClientCountStatProps {
  labelId: string;
  // null only while the count query is in flight; the stat renders its label
  // with an em dash rather than a misleading 0.
  count: number | null;
  quota: number | null;
}

const ClientCountStat: React.VFC<ClientCountStatProps> =
  function ClientCountStat({ labelId, count, quota }) {
    return (
      <div className={styles.stat}>
        <Text as="p" size="1" color="gray" className={styles.statLabel}>
          <FormattedMessage id={labelId} />
        </Text>
        <Text as="p" size="6" weight="bold" className={styles.statValue}>
          {count == null ? (
            <FormattedMessage id="DynamicClientsTab.clients.count.unknown" />
          ) : quota != null ? (
            <FormattedMessage
              id="DynamicClientsTab.clients.count.quota"
              values={{ count, quota }}
            />
          ) : (
            <FormattedMessage
              id="DynamicClientsTab.clients.count"
              values={{ count }}
            />
          )}
        </Text>
      </div>
    );
  };

export interface DynamicClientsTabProps {
  form: AppConfigFormModel<ApplicationsFormState>;
  publicOrigin: string;
  // The smallest block-action quota configured for oauth_client_dcr, or null
  // when the plan does not limit DCR-registered clients.
  dcrClientQuota: number | null;
  // The same for oauth_client_cimd. The two quotas are counted separately
  // (docs/specs/cimd.md § Client Limit), so neither number can stand in for
  // the other.
  cimdClientQuota: number | null;
}

export const DynamicClientsTab: React.VFC<DynamicClientsTabProps> =
  function DynamicClientsTab({
    form,
    publicOrigin,
    dcrClientQuota,
    cimdClientQuota,
  }) {
    const navigate = useNavigate();
    const { appID } = useParams() as { appID: string };
    const { state, setState, isUpdating, effectiveConfig } = form;
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
    const registrationEndpoint = `${publicOrigin}/oauth2/register`;

    // One query per source, because the card states each mechanism's count
    // against its own quota and a source-less totalCount cannot be split
    // after the fact. Both are always fetched and both stats always render,
    // including at 0: turning a mechanism off neither deletes nor disables
    // the clients already resolved or registered, so the card has to keep
    // reporting them either way.
    const { data: cimdData } = useDynamicClientsQueryQuery({
      variables: { first: 1, source: OAuthClientSource.Cimd },
      fetchPolicy: "cache-and-network",
    });
    const { data: dcrData } = useDynamicClientsQueryQuery({
      variables: { first: 1, source: OAuthClientSource.Dcr },
      fetchPolicy: "cache-and-network",
    });
    const cimdCount = cimdData?.dynamicClients?.totalCount ?? null;
    const dcrCount = dcrData?.dynamicClients?.totalCount ?? null;
    // Only DCR's own count gates anything now: its disable dialog warns
    // that registered clients keep working, which is only worth saying when
    // some exist.
    const hasDCRClients = (dcrCount ?? 0) > 0;

    const setRegistrationEnabled = useCallback(
      (checked: boolean) => {
        // Deferred like every other control on this tab: nothing is written
        // until the admin presses Save. Saving from here would commit the
        // whole form state, including edits elsewhere on the tab (and on the
        // sibling Applications tab, which shares this form) that the admin
        // has not confirmed yet.
        setState((prev) => ({
          ...prev,
          dynamicClientRegistrationEnabled: checked,
          // Switching registration off resets the initial access token
          // requirement, so re-enabling always starts from the safe default
          // rather than quietly restoring open registration. Turning the
          // requirement off again is an explicit, separately confirmed act.
          initialAccessTokenRequired: checked
            ? prev.initialAccessTokenRequired
            : true,
        }));
      },
      [setState]
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
          title={<FormattedMessage id="DynamicClientsTab.clients.title" />}
        >
          <div className={styles.stats}>
            <ClientCountStat
              labelId="DynamicClientSource.cimd"
              count={cimdCount}
              quota={cimdClientQuota}
            />
            <ClientCountStat
              labelId="DynamicClientSource.dcr"
              count={dcrCount}
              quota={dcrClientQuota}
            />
          </div>
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

        <SettingsSectionCard
          contentClassName="gap-4"
          title={
            <FormattedMessage id="DynamicClientsTab.allowed-resources.title" />
          }
          description={
            <FormattedMessage id="DynamicClientsTab.allowed-resources.description" />
          }
        >
          <DynamicClientAllowedResources />
          <Text as="p" size="2" color="gray">
            <FormattedMessage
              id="DynamicClientsTab.allowed-resources.manage"
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
        </SettingsSectionCard>

        <CIMDSection form={form} />

        <MechanismSectionHeader
          title={
            <FormattedMessage id="DynamicClientsTab.section.registration.title" />
          }
          description={
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
          }
          toggleLabel={
            <FormattedMessage id="DynamicClientsTab.enable.toggle.label" />
          }
          checked={registrationEnabled}
          disabled={isUpdating}
          onCheckedChange={onEnabledChange}
        />

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
    );
  };
