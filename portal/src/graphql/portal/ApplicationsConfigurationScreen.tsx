import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import cn from "classnames";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { DotsVerticalIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import { useLocation, useNavigate, useParams } from "react-router-dom";
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
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { useOAuthClientForm } from "../../hook/useOAuthClientForm";
import { getNextPlan } from "../../util/plan";
import { RolesAndGroupsEmptyView } from "../../components/roles-and-groups/empty-view/RolesAndGroupsEmptyView";

// Navigation state the create-application screens attach when they send the
// user back to this list (breadcrumb / Cancel). It suppresses the empty-list
// auto-redirect to ./add, so cancelling the create flow can actually land on
// the empty list instead of bouncing back to the wizard (DEV-3810).
export const FROM_CREATE_APPLICATION_FLOW_STATE = {
  fromCreateApplicationFlow: true,
};

interface FormState {
  clients: OAuthClientConfig[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    clients: config.oauth?.clients ?? [],
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

interface OAuthClientConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  planName: string | null;
  oauthClientsSoftMaximum: number | undefined;
  oauthClientsHardMaximum: number | undefined;
}

const OAuthClientConfigurationContent: React.VFC<OAuthClientConfigurationContentProps> =
  function OAuthClientConfigurationContent(props) {
    const {
      form: { state, reload },
      planName,
      oauthClientsHardMaximum,
      oauthClientsSoftMaximum,
    } = props;
    const navigate = useNavigate();
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

    const isEmpty = state.clients.length === 0;

    return (
      <ScreenContent layout="list">
        <div className={styles.pageHeader}>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          </Text>
          {isEmpty ? null : (
            <Tooltip
              content={
                <FormattedMessage
                  id="ApplicationsConfigurationScreen.add-client-button.hard-limit-tooltip"
                  values={{ maximum: oauthClientsHardMaximum ?? 0 }}
                />
              }
              disabled={!hardLimitReached}
            >
              {/* The tooltip must still fire when the button is disabled, so it
                anchors on a wrapper span instead of the button itself. */}
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
          )}
        </div>
        {isEmpty ? (
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
            <Text
              as="p"
              size="2"
              className={cn(styles.widget, styles.pageDescription)}
            >
              <FormattedMessage id="ApplicationsConfigurationScreen.description" />
            </Text>
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
    const navigate = useNavigate();
    const location = useLocation();

    const fromCreateApplicationFlow =
      (location.state as { fromCreateApplicationFlow?: boolean } | null)
        ?.fromCreateApplicationFlow === true;

    // Whether this mounted screen has ever shown a non-empty list. Deleting
    // the last application should reveal the empty state, not throw the user
    // into the create wizard.
    const [hadClients, setHadClients] = useState(false);

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
    const featureConfig = useAppFeatureConfigQuery(appID);

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

    if (!isLoading && !error && form.state.clients.length > 0 && !hadClients) {
      // Adjust-state-during-render pattern; React re-renders immediately
      // without committing the stale output.
      setHadClients(true);
    }

    // A visit with no application still lands directly on the create wizard
    // (e.g. clicking Applications in the nav), unless the user is coming back
    // from that wizard — then show the empty list so the flow can be quit.
    const shouldRedirectToAdd =
      !isLoading &&
      !error &&
      form.state.clients.length === 0 &&
      !fromCreateApplicationFlow &&
      !hadClients;

    useEffect(() => {
      if (shouldRedirectToAdd) {
        navigate("./add", { replace: true });
      }
    }, [shouldRedirectToAdd, navigate]);

    if (isLoading || shouldRedirectToAdd) {
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
        />
      </FormContainer>
    );
  };

export default ApplicationsConfigurationScreen;
