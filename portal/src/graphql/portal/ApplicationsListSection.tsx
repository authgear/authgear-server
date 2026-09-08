import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { DotsVerticalIcon } from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import { produce } from "immer";
import { Context, FormattedMessage } from "../../intl";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { AppConfigFormModel } from "../../hook/useAppConfigForm";
import { getApplicationTypeMessageID } from "./EditOAuthClientForm";
import { findFramework } from "./CreateOAuthClientScreen/frameworks";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { CardTable } from "../../components/v2/CardTable/CardTable";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { FeatureDisabledCallout } from "../../components/v2/FeatureDisabledCallout/FeatureDisabledCallout";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { useOAuthClientForm } from "../../hook/useOAuthClientForm";
import { getNextPlan } from "../../util/plan";
import { RolesAndGroupsEmptyView } from "../../components/roles-and-groups/empty-view/RolesAndGroupsEmptyView";
import styles from "./ApplicationsListSection.module.css";

// Only the static client list. CIMD and DCR each own their own form, so
// neither mechanism's fields belong here: a save from this form must never
// carry an edit made on one of them.
export interface FormState {
  clients: OAuthClientConfig[];
}

export function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    clients: config.oauth?.clients ?? [],
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

export interface ApplicationsListSectionProps {
  form: AppConfigFormModel<FormState>;
  planName: string | null;
  oauthClientsSoftMaximum: number | undefined;
  oauthClientsHardMaximum: number | undefined;
}

/**
 * The static OAuth client list: the add button and its plan-limit tooltip,
 * the feature-config callout, the CardTable of clients, the empty state, and
 * the delete confirmation.
 *
 * Rendered as direct children of a `ScreenContent layout="list"` grid, so
 * every top-level element carries the grid-spanning `widget` class.
 */
export const ApplicationsListSection: React.VFC<ApplicationsListSectionProps> =
  function ApplicationsListSection(props) {
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
      <>
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
      </>
    );
  };
