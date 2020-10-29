import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ActionButton,
  DefaultButton,
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  MessageBar,
  SelectionMode,
  Text,
  VerticalDivider,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams, Link } from "react-router-dom";
import produce from "immer";
import deepEqual from "deep-equal";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import FormTextFieldList from "../../FormTextFieldList";
import { PortalAPIApp, PortalAPIAppConfig } from "../../types";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { copyToClipboard } from "../../util/clipboard";
import { clearEmptyObject, setFieldIfListNonEmpty } from "../../util/misc";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useValidationError } from "../../error/useValidationError";
import { FormContext } from "../../error/FormContext";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";

import styles from "./OAuthClientConfigurationScreen.module.scss";

interface OAuthClientConfigurationProps {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
  showNotification: (msg: string) => void;
}

interface AllowedOriginsConfigurationProps {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
}

interface OAuthClientListItem {
  name: string;
  clientId: string;
}

interface OAuthClientListActionCellProps {
  clientId: string;
  clientName: string;
  onCopyComplete: () => void;
  onRemoveClientClick: (clientId: string, clientName: string) => void;
}

interface ConfirmRemoveDialogData {
  clientId: string;
  clientName: string;
}

interface ConfirmRemoveOAuthClientDialogProps extends ConfirmRemoveDialogData {
  visible: boolean;
  updatingAppConfig: boolean;
  onDismiss: () => void;
  removeOAuthClient: (clientId: string) => void;
}

const ADD_CLIENT_BUTTON_STYLES = {
  icon: { paddingRight: "4px" },
  flexContainer: { paddingRight: "2px" },
};

const DIALOG_FOOTER_STYLES = {
  actionsRight: { display: "flex", justifyContent: "flex-end" },
};

function makeOAuthClientListColumns(
  renderToString: (messageId: string) => string
): IColumn[] {
  return [
    {
      key: "name",
      fieldName: "name",
      name: renderToString("OAuthClientConfiguration.client-list.name"),
      minWidth: 150,
      className: styles.clientListColumn,
    },

    {
      key: "clientId",
      fieldName: "clientId",
      name: renderToString("OAuthClientConfiguration.client-list.client-id"),
      minWidth: 300,
      className: styles.clientListColumn,
    },
    { key: "action", name: renderToString("action"), minWidth: 200 },
  ];
}

function getAllowedOriginJSONPointer(index: number): string {
  return `/http/allowed_origins/${index}`;
}

const AllowedOriginsConfiguration: React.FC<AllowedOriginsConfigurationProps> = function AllowedOriginsConfiguration(
  props: AllowedOriginsConfigurationProps
) {
  const { effectiveAppConfig, rawAppConfig, updateAppConfig } = props;

  const initialAllowedOrigins = useMemo(() => {
    return effectiveAppConfig?.http?.allowed_origins ?? [];
  }, [effectiveAppConfig]);

  const [allowedOrigins, setAllowedOrigins] = useState<string[]>(
    initialAllowedOrigins
  );

  const onAllowedOriginsChange = useCallback((list: string[]) => {
    setAllowedOrigins(list);
  }, []);

  const onSaveClick = useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      draftConfig.http = draftConfig.http ?? {};
      setFieldIfListNonEmpty(
        draftConfig.http,
        "allowed_origins",
        allowedOrigins
      );

      clearEmptyObject(draftConfig);
    });

    updateAppConfig(newAppConfig).catch(() => {});
  }, [rawAppConfig, updateAppConfig, allowedOrigins]);

  const isModified = useMemo(() => {
    return !deepEqual(allowedOrigins, initialAllowedOrigins);
  }, [allowedOrigins, initialAllowedOrigins]);

  const resetForm = useCallback(() => {
    setAllowedOrigins(initialAllowedOrigins);
  }, [initialAllowedOrigins]);

  return (
    <section className={styles.allowedOriginsConfiguration}>
      <ModifiedIndicatorPortal resetForm={resetForm} isModified={isModified} />
      <Text as="h2" className={styles.allowedOriginsConfigurationHeader}>
        <FormattedMessage id="OAuthClientConfigurationScreen.allowed-origins.header" />
      </Text>
      <Text className={styles.allowedOriginsConfigurationDesc}>
        <FormattedMessage id="OAuthClientConfigurationScreen.allowed-origins.desc" />
      </Text>
      <FormTextFieldList
        className={styles.allowedOriginsInputList}
        jsonPointer="/http/allowed_origins"
        parentJSONPointer="/http"
        getItemJSONPointer={getAllowedOriginJSONPointer}
        fieldName="allowed_origins"
        list={allowedOrigins}
        onListChange={onAllowedOriginsChange}
        addButtonLabelMessageID="OAuthClientConfigurationScreen.add-origin"
      />
      <ButtonWithLoading
        labelId="save"
        disabled={!isModified}
        loading={false}
        onClick={onSaveClick}
      />
    </section>
  );
};

const ConfirmRemoveOAuthClientDialog: React.FC<ConfirmRemoveOAuthClientDialogProps> = function ConfirmRemoveOAuthClientDialog(
  props: ConfirmRemoveOAuthClientDialogProps
) {
  const {
    visible,
    updatingAppConfig,
    onDismiss: onDismissProps,
    removeOAuthClient,
    clientId,
    clientName,
  } = props;
  const { renderToString } = useContext(Context);

  const onConfirm = useCallback(() => {
    removeOAuthClient(clientId);
  }, [clientId, removeOAuthClient]);

  const onDismiss = useCallback(() => {
    if (!updatingAppConfig) {
      onDismissProps();
    }
  }, [onDismissProps, updatingAppConfig]);

  const confirmRemoveDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="OAuthClientConfigurationScreen.confirm-remove-dialog.title" />
      ),

      subText: renderToString(
        "OAuthClientConfigurationScreen.confirm-remove-dialog.message",
        { clientName }
      ),
    };
  }, [renderToString, clientName]);

  return (
    <Dialog
      hidden={!visible}
      dialogContentProps={confirmRemoveDialogContentProps}
      modalProps={{ isBlocking: updatingAppConfig }}
      onDismiss={onDismiss}
    >
      <DialogFooter
        styles={DIALOG_FOOTER_STYLES}
        className={styles.confirmDialogButtons}
      >
        <ButtonWithLoading
          onClick={onConfirm}
          labelId="confirm"
          loading={updatingAppConfig}
          disabled={!visible}
        />
        <DefaultButton
          disabled={updatingAppConfig || !visible}
          onClick={onDismiss}
        >
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

const OAuthClientListActionCell: React.FC<OAuthClientListActionCellProps> = function OAuthClientListActionCell(
  props: OAuthClientListActionCellProps
) {
  const { clientId, clientName, onCopyComplete, onRemoveClientClick } = props;
  const navigate = useNavigate();
  const { themes } = useSystemConfig();

  const onEditClick = useCallback(() => {
    navigate(`./${clientId}/edit`);
  }, [navigate, clientId]);

  const onCopyClick = useCallback(() => {
    copyToClipboard(clientId);

    // Invoke callback
    onCopyComplete();
  }, [clientId, onCopyComplete]);

  const onRemoveClick = useCallback(() => {
    onRemoveClientClick(clientId, clientName);
  }, [clientId, clientName, onRemoveClientClick]);

  return (
    <div className={styles.clientListColumn}>
      <ActionButton
        className={styles.listActionButton}
        theme={themes.actionButton}
        onClick={onEditClick}
      >
        <FormattedMessage id="edit" />
      </ActionButton>
      <VerticalDivider className={styles.listActionButtonDivider} />
      <ActionButton
        className={styles.listActionButton}
        theme={themes.actionButton}
        onClick={onCopyClick}
      >
        <FormattedMessage id="copy" />
      </ActionButton>
      <VerticalDivider className={styles.listActionButtonDivider} />
      <ActionButton
        className={styles.listActionButton}
        theme={themes.actionButton}
        onClick={onRemoveClick}
      >
        <FormattedMessage id="remove" />
      </ActionButton>
    </div>
  );
};

const OAuthClientConfiguration: React.FC<OAuthClientConfigurationProps> = function OAuthClientConfiguration(
  props: OAuthClientConfigurationProps
) {
  const { rawAppConfig, effectiveAppConfig, showNotification } = props;
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const navigate = useNavigate();
  const { themes, authgearEndpoint } = useSystemConfig();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const [confirmRemoveDialogData, setConfirmRemoveDialogData] = useState<
    ConfirmRemoveDialogData
  >({
    clientName: "",
    clientId: "",
  });
  const [confirmRemoveDialogVisible, setConfirmRemoveDialogVisible] = useState(
    false
  );

  const oauthClients = useMemo(() => {
    return effectiveAppConfig?.oauth?.clients ?? [];
  }, [effectiveAppConfig]);

  const oauthClientListColumns = useMemo(() => {
    return makeOAuthClientListColumns(renderToString);
  }, [renderToString]);

  const oauthClientListItems: OAuthClientListItem[] = useMemo(() => {
    return oauthClients.map((client) => {
      return {
        name: client.name!,
        clientId: client.client_id,
      };
    });
  }, [oauthClients]);

  const onAddOAuthClientClick = useCallback(() => {
    navigate("./add");
  }, [navigate]);

  const onClientIdCopied = useCallback(() => {
    showNotification(
      renderToString("OAuthClientConfiguration.client-id-copied")
    );
  }, [showNotification, renderToString]);

  const onRemoveClientClick = useCallback(
    (clientId: string, clientName: string) => {
      setConfirmRemoveDialogData({
        clientId,
        clientName,
      });
      setConfirmRemoveDialogVisible(true);
    },
    []
  );

  const dismissConfirmRemoveDialog = useCallback(() => {
    setConfirmRemoveDialogVisible(false);
  }, []);

  const removeOAuthClient = useCallback(
    (clientId: string) => {
      if (rawAppConfig == null) {
        return;
      }

      const newAppConfig = produce(rawAppConfig, (draftConfig) => {
        const clients = draftConfig.oauth!.clients!;
        const updatedClients = clients.filter(
          (client) => client.client_id !== clientId
        );
        if (clients.length === updatedClients.length) {
          console.warn("[Remove OAuth Client]: OAuth client not found");
          return;
        }

        draftConfig.oauth!.clients = updatedClients;
      });

      updateAppConfig(newAppConfig)
        .catch(() => {})
        .finally(() => {
          dismissConfirmRemoveDialog();
        });
    },
    [rawAppConfig, updateAppConfig, dismissConfirmRemoveDialog]
  );

  const onRenderOAuthClientColumns = useCallback(
    (item?: OAuthClientListItem, _index?: number, column?: IColumn) => {
      if (item == null || column == null) {
        return null;
      }
      const fieldContent = item[column.fieldName as keyof OAuthClientListItem];
      switch (column.key) {
        case "action":
          return (
            <OAuthClientListActionCell
              clientId={item.clientId}
              clientName={item.name}
              onCopyComplete={onClientIdCopied}
              onRemoveClientClick={onRemoveClientClick}
            />
          );
        default:
          return (
            <span className={styles.clientListColumnContent}>
              {fieldContent}
            </span>
          );
      }
    },
    [onClientIdCopied, onRemoveClientClick]
  );

  const {
    otherError,
    unhandledCauses,
    value: formContextValue,
  } = useValidationError(updateAppConfigError);

  return (
    <FormContext.Provider value={formContextValue}>
      <section className={styles.content}>
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        {(unhandledCauses ?? []).length === 0 && otherError && (
          <ShowError error={otherError} />
        )}
        <AllowedOriginsConfiguration
          effectiveAppConfig={effectiveAppConfig}
          rawAppConfig={rawAppConfig}
          updateAppConfig={updateAppConfig}
        />
        <section className={styles.clientEndpointSession}>
          <Text as="h2" className={styles.clientEndpointSessionHeader}>
            <FormattedMessage id="OAuthClientConfigurationScreen.client-endpoint.header" />
          </Text>
          <Text className={styles.clientEndpointSessionDesc}>
            <FormattedMessage
              id="OAuthClientConfigurationScreen.client-endpoint.desc"
              values={{
                clientEndpoint: authgearEndpoint,
                dnsUrl: "../dns",
              }}
              components={{
                Link,
              }}
            />
          </Text>
        </section>
        <section className={styles.controlButtons}>
          <ActionButton
            theme={themes.actionButton}
            className={styles.addClientButton}
            onClick={onAddOAuthClientClick}
            iconProps={{ iconName: "CirclePlus" }}
            styles={ADD_CLIENT_BUTTON_STYLES}
          >
            <FormattedMessage id="OAuthClientConfiguration.add-client-button" />
          </ActionButton>
        </section>
        <DetailsList
          columns={oauthClientListColumns}
          items={oauthClientListItems}
          selectionMode={SelectionMode.none}
          onRenderItemColumn={onRenderOAuthClientColumns}
        />
        <ConfirmRemoveOAuthClientDialog
          visible={confirmRemoveDialogVisible}
          updatingAppConfig={updatingAppConfig}
          onDismiss={dismissConfirmRemoveDialog}
          removeOAuthClient={removeOAuthClient}
          clientName={confirmRemoveDialogData.clientName}
          clientId={confirmRemoveDialogData.clientId}
        />
      </section>
    </FormContext.Provider>
  );
};

const OAuthClientConfigurationScreen: React.FC = function OAuthClientConfigurationScreen() {
  const { appID } = useParams();
  const {
    effectiveAppConfig,
    rawAppConfig,
    loading,
    error,
    refetch,
  } = useAppConfigQuery(appID);

  const [isNotificationVisible, setIsNotificationVisible] = useState(false);
  const [notificationMsg, setNotificationMsg] = useState("");

  const showNotification = useCallback((msg: string) => {
    setIsNotificationVisible(true);
    setNotificationMsg(msg);
  }, []);

  const dismissNotification = useCallback(() => {
    setIsNotificationVisible(false);
  }, []);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      <ModifiedIndicatorWrapper className={styles.wrapper}>
        {isNotificationVisible && (
          <MessageBar onDismiss={dismissNotification}>
            <p>{notificationMsg}</p>
          </MessageBar>
        )}
        <Text as="h1" className={styles.header}>
          <FormattedMessage id="OAuthClientConfiguration.title" />
        </Text>
        <OAuthClientConfiguration
          rawAppConfig={rawAppConfig}
          effectiveAppConfig={effectiveAppConfig}
          showNotification={showNotification}
        />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default OAuthClientConfigurationScreen;
