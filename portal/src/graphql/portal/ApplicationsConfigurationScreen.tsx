import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  DropdownMenu,
  Heading,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { DotsVerticalIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import { useNavigate, useParams } from "react-router-dom";
import { produce } from "immer";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import styles from "./ApplicationsConfigurationScreen.module.css";
import ScreenContent from "../../ScreenContent";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { getApplicationTypeMessageID } from "./EditOAuthClientForm";
import { findFramework } from "./CreateOAuthClientScreen/frameworks";
import FormContainer from "../../FormContainer";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { CardTable } from "../../components/v2/CardTable/CardTable";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { FeatureDisabledCallout } from "../../components/v2/FeatureDisabledCallout/FeatureDisabledCallout";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { useOAuthClientForm } from "../../hook/useOAuthClientForm";
import { usePivotNavigation } from "../../hook/usePivot";
import { getNextPlan } from "../../util/plan";
import { RolesAndGroupsEmptyView } from "../../components/roles-and-groups/empty-view/RolesAndGroupsEmptyView";
import { DynamicClientsTab } from "../../components/dynamic-clients/DynamicClientsTab";

export interface FormState {
  clients: OAuthClientConfig[];
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
    clients: config.oauth?.clients ?? [],
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
      config.oauth.clients = currentState.clients;

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

function stopPropagation(e: React.SyntheticEvent) {
  e.stopPropagation();
}

interface ClientRowProps {
  client: OAuthClientConfig;
  onDeleteClick: (clientID: string) => void;
}

const ClientRow: React.VFC<ClientRowProps> = function ClientRow(props) {
  const { client, onDeleteClick } = props;
  const { renderToString } = useContext(Context);
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();

  const editPath = `/project/${appID}/configuration/apps/${encodeURIComponent(
    client.client_id
  )}/edit`;

  const framework = findFramework(client.x_framework);
  const fallbackIcon =
    client.x_application_type === "m2m" ? "server" : "app-window";
  const iconName = framework?.iconName ?? fallbackIcon;

  const onRowClick = useCallback(() => {
    navigate(editPath);
  }, [navigate, editPath]);

  const onRowKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLDivElement>) => {
      // Only navigate when the row itself is focused; Enter/Space on the
      // copy button or the actions menu must not trigger navigation.
      if (e.target !== e.currentTarget) {
        return;
      }
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        navigate(editPath);
      }
    },
    [navigate, editPath]
  );

  return (
    <CardTable.Row
      className={styles.clientRow}
      role="button"
      tabIndex={0}
      onClick={onRowClick}
      onKeyDown={onRowKeyDown}
    >
      <CardTable.Cell className={styles.colName}>
        <div className={styles.clientIconWrap}>
          <i
            className={cn("ti", `ti-${iconName}`, styles.clientIcon)}
            aria-hidden={true}
          />
        </div>
        <div className={styles.clientNameBlock}>
          <Text size="2" className={styles.clientName}>
            {client.name ?? ""}
          </Text>
          <Text size="1" className={styles.clientSubtitle}>
            <FormattedMessage
              id={getApplicationTypeMessageID(client.x_application_type)}
            />
            {framework != null ? (
              <>
                {" · "}
                <FormattedMessage id={framework.displayNameMessageId} />
              </>
            ) : null}
          </Text>
          <div className={styles.compactClientId} onClick={stopPropagation}>
            <Text size="1" className={styles.clientIdText}>
              {client.client_id}
            </Text>
            <CopyIconButton textToCopy={client.client_id} />
          </div>
        </div>
      </CardTable.Cell>
      <CardTable.Cell className={styles.colClientId} onClick={stopPropagation}>
        <Text size="2" className={styles.clientIdText}>
          {client.client_id}
        </Text>
        <CopyIconButton textToCopy={client.client_id} />
      </CardTable.Cell>
      <CardTable.Cell className={styles.colActions} onClick={stopPropagation}>
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            <RadixIconButton
              className={styles.rowActionsButton}
              variant="soft"
              color="gray"
              size="2"
              aria-label={renderToString(
                "ApplicationsConfigurationScreen.client-list.row-actions"
              )}
            >
              <DotsVerticalIcon width="1rem" height="1rem" />
            </RadixIconButton>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item
              onSelect={() => {
                navigate(editPath);
              }}
            >
              <FormattedMessage id="edit" />
            </DropdownMenu.Item>
            <DropdownMenu.Item
              color="red"
              onSelect={() => {
                onDeleteClick(client.client_id);
              }}
            >
              <FormattedMessage id="ApplicationsConfigurationScreen.delete-client.label" />
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </CardTable.Cell>
    </CardTable.Row>
  );
};

type ApplicationsTabKey = "applications" | "dynamic-clients";

interface OAuthClientConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  planName: string | null;
  oauthClientsSoftMaximum: number | undefined;
  oauthClientsHardMaximum: number | undefined;
  selectedKey: ApplicationsTabKey;
  onChangeKey: (key: ApplicationsTabKey) => void;
  publicOrigin: string;
  dcrClientQuota: number | null;
}

const OAuthClientConfigurationContent: React.VFC<OAuthClientConfigurationContentProps> =
  function OAuthClientConfigurationContent(props) {
    const {
      form: { state, reload },
      planName,
      oauthClientsHardMaximum,
      oauthClientsSoftMaximum,
      selectedKey,
      onChangeKey,
      publicOrigin,
      dcrClientQuota,
    } = props;
    const navigate = useNavigate();
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const deleteForm = useOAuthClientForm(appID, null);

    const [isRemoveDialogVisible, setIsRemoveDialogVisible] = useState(false);

    const hardLimitReached = useMemo(() => {
      if (oauthClientsHardMaximum == null) {
        return false;
      }
      return state.clients.length >= oauthClientsHardMaximum;
    }, [oauthClientsHardMaximum, state.clients.length]);

    const displayedClientMaximum = useMemo<number | undefined>(() => {
      return oauthClientsSoftMaximum ?? oauthClientsHardMaximum;
    }, [oauthClientsHardMaximum, oauthClientsSoftMaximum]);

    const goToCreateApp = useCallback(() => {
      navigate(`/project/${appID}/configuration/apps/add`);
    }, [appID, navigate]);

    const showDialogAndSetRemoveClientByID = useCallback(
      (clientID: string) => {
        deleteForm.setState((state) => ({
          ...state,
          removeClientByID: clientID,
        }));
        setIsRemoveDialogVisible(true);
      },
      [deleteForm, setIsRemoveDialogVisible]
    );

    const dismissDialogAndResetRemoveClientByID = useCallback(() => {
      setIsRemoveDialogVisible(false);
      deleteForm.setState((state) => {
        return {
          ...state,
          removeClientByID: undefined,
        };
      });
    }, [deleteForm, setIsRemoveDialogVisible]);

    const onRemoveDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open && !deleteForm.isUpdating) {
          dismissDialogAndResetRemoveClientByID();
        }
      },
      [deleteForm.isUpdating, dismissDialogAndResetRemoveClientByID]
    );

    const onConfirmRemove = useCallback(() => {
      deleteForm.save().then(
        () => {
          dismissDialogAndResetRemoveClientByID();
          reload();
        },
        () => {
          dismissDialogAndResetRemoveClientByID();
        }
      );
    }, [deleteForm, reload, dismissDialogAndResetRemoveClientByID]);

    const canUpgradePlan = useMemo(() => {
      return getNextPlan(planName ?? "") != null;
    }, [planName]);

    const displayMaximumWarning = useMemo(() => {
      if (displayedClientMaximum == null) {
        return false;
      }
      return state.clients.length >= displayedClientMaximum;
    }, [state, displayedClientMaximum]);

    const onTabChange = useCallback(
      (value: string) => {
        onChangeKey(value as ApplicationsTabKey);
      },
      [onChangeKey]
    );

    const tabOptions = useMemo(
      () => [
        {
          value: "applications",
          label: renderToString(
            "ApplicationsConfigurationScreen.tab.applications"
          ),
        },
        {
          value: "dynamic-clients",
          label: renderToString(
            "ApplicationsConfigurationScreen.tab.dynamic-clients"
          ),
        },
      ],
      [renderToString]
    );

    const isEmpty = state.clients.length === 0;

    return (
      <ScreenContent layout="list">
        <div className={styles.pageHeader}>
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          </Heading>
        </div>
        <OverflowTabs
          className={styles.widget}
          value={selectedKey}
          onValueChange={onTabChange}
          tabs={tabOptions}
        />
        {selectedKey === "applications" ? (
          isEmpty ? (
            <div className={cn(styles.widget, styles.emptyState)}>
              <RolesAndGroupsEmptyView
                icon={
                  <span className={styles.emptyStateIconWrap}>
                    <i
                      className={cn("ti", "ti-apps", styles.emptyStateIcon)}
                      aria-hidden={true}
                    />
                  </span>
                }
                title={
                  <FormattedMessage id="ApplicationsConfigurationScreen.empty-state.title" />
                }
                description={
                  <FormattedMessage id="ApplicationsConfigurationScreen.empty-state.description" />
                }
                button={
                  <RolesAndGroupsEmptyView.CreateButton
                    onClick={goToCreateApp}
                    text={
                      <FormattedMessage id="ApplicationsConfigurationScreen.add-client-button" />
                    }
                  />
                }
              />
            </div>
          ) : (
            <>
              <div className={cn(styles.widget, styles.listHeader)}>
                <Text as="p" size="2" className={styles.pageDescription}>
                  <FormattedMessage id="ApplicationsConfigurationScreen.description" />
                </Text>
                <Tooltip
                  content={
                    <FormattedMessage
                      id="ApplicationsConfigurationScreen.add-client-button.hard-limit-tooltip"
                      values={{ maximum: oauthClientsHardMaximum ?? 0 }}
                    />
                  }
                  disabled={!hardLimitReached}
                >
                  {/* The tooltip must still fire when the button is disabled,
                    so it anchors on a wrapper span instead of the button
                    itself. */}
                  <span>
                    <PrimaryButton
                      size="2"
                      text={
                        <FormattedMessage id="ApplicationsConfigurationScreen.add-client-button" />
                      }
                      onClick={goToCreateApp}
                      disabled={hardLimitReached}
                    />
                  </span>
                </Tooltip>
              </div>
              <div className={cn(styles.widget, styles.listSection)}>
                {displayMaximumWarning ? (
                  <FeatureDisabledCallout
                    messageID={
                      canUpgradePlan
                        ? "FeatureConfig.oauth-clients.maximum.upgrade"
                        : "FeatureConfig.oauth-clients.maximum.contact-us"
                    }
                    messageValues={{ maximum: displayedClientMaximum! }}
                  />
                ) : null}
                <CardTable>
                  <CardTable.Header>
                    <CardTable.HeaderCell className={styles.colName}>
                      <FormattedMessage id="ApplicationsConfigurationScreen.client-list.name" />
                    </CardTable.HeaderCell>
                    <CardTable.HeaderCell className={styles.colClientId}>
                      <FormattedMessage id="ApplicationsConfigurationScreen.client-list.client-id" />
                    </CardTable.HeaderCell>
                    <CardTable.HeaderCell className={styles.colActions} />
                  </CardTable.Header>
                  {state.clients.map((client) => (
                    <ClientRow
                      key={client.client_id}
                      client={client}
                      onDeleteClick={showDialogAndSetRemoveClientByID}
                    />
                  ))}
                </CardTable>
              </div>
            </>
          )
        ) : (
          <div className={styles.widget}>
            <DynamicClientsTab
              form={props.form}
              publicOrigin={publicOrigin}
              dcrClientQuota={dcrClientQuota}
            />
          </div>
        )}
        <ConfirmationDialog
          open={isRemoveDialogVisible}
          onOpenChange={onRemoveDialogOpenChange}
          title={
            <FormattedMessage id="ApplicationsConfigurationScreen.delete-client-dialog.title" />
          }
          description={
            <FormattedMessage id="ApplicationsConfigurationScreen.delete-client-dialog.description" />
          }
          confirmText={<FormattedMessage id="confirm" />}
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={onConfirmRemove}
          onCancel={dismissDialogAndResetRemoveClientByID}
          loading={deleteForm.isUpdating}
          confirmColor="red"
        />
      </ScreenContent>
    );
  };

const ApplicationsConfigurationScreen: React.VFC =
  function ApplicationsConfigurationScreen() {
    const { appID } = useParams() as { appID: string };

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);
    const { selectedKey, onChangeKey } = usePivotNavigation<ApplicationsTabKey>(
      ["applications", "dynamic-clients"]
    );

    const publicOrigin = useMemo(() => {
      return form.effectiveConfig.http?.public_origin ?? "";
    }, [form.effectiveConfig]);

    const dcrClientQuota = useMemo<number | null>(() => {
      const limits =
        featureConfig.effectiveFeatureConfig?.usage?.limits?.oauth_client_dcr;
      const blockQuotas = (limits ?? [])
        .filter((limit) => limit.action === "block")
        .map((limit) => limit.quota)
        .filter((quota): quota is number => quota != null);
      if (blockQuotas.length === 0) {
        return null;
      }
      return Math.min(...blockQuotas);
    }, [featureConfig.effectiveFeatureConfig]);

    const oauthClientsHardMaximum = useMemo<number | undefined>(() => {
      return featureConfig.effectiveFeatureConfig?.oauth?.client?.maximum;
    }, [featureConfig]);

    const oauthClientsSoftMaximum = useMemo(() => {
      return featureConfig.effectiveFeatureConfig?.oauth?.client?.soft_maximum;
    }, [featureConfig]);

    const isLoading = useMemo(
      () => form.isLoading || featureConfig.isLoading,
      [form.isLoading, featureConfig.isLoading]
    );

    const error = useMemo(
      () => form.loadError ?? featureConfig.loadError,
      [form.loadError, featureConfig.loadError]
    );

    const onRetry = useCallback(() => {
      if (form.loadError) {
        form.reload();
      }

      if (featureConfig.loadError) {
        featureConfig.refetch().finally(() => {});
      }
    }, [form, featureConfig]);

    if (isLoading) {
      return <ShowLoading />;
    }

    if (error) {
      return <ShowError error={error} onRetry={onRetry} />;
    }

    return (
      <FormContainer form={form}>
        <OAuthClientConfigurationContent
          form={form}
          planName={featureConfig.planName}
          oauthClientsHardMaximum={oauthClientsHardMaximum}
          oauthClientsSoftMaximum={oauthClientsSoftMaximum}
          selectedKey={selectedKey}
          onChangeKey={onChangeKey}
          publicOrigin={publicOrigin}
          dcrClientQuota={dcrClientQuota}
        />
      </FormContainer>
    );
  };

export default ApplicationsConfigurationScreen;
