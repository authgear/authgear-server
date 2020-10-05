import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ActionButton,
  DetailsList,
  IColumn,
  MessageBar,
  SelectionMode,
  Text,
  VerticalDivider,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { copyToClipboard } from "../../util/clipboard";
import { actionButtonTheme } from "../../theme";

import styles from "./OAuthClientConfigurationScreen.module.scss";

interface OAuthClientConfigurationProps {
  appConfig: PortalAPIAppConfig | null;
  showNotification: (msg: string) => void;
}

interface OAuthClientListItem {
  name: string;
  clientId: string;
}

interface OAuthClientListActionCellProps {
  clientId: string;
  onCopyComplete: () => void;
}

const ADD_CLIENT_BUTTON_STYLES = {
  icon: { paddingRight: "4px" },
  flexContainer: { paddingRight: "2px" },
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

const OAuthClientListActionCell: React.FC<OAuthClientListActionCellProps> = function OAuthClientListActionCell(
  props: OAuthClientListActionCellProps
) {
  const { clientId, onCopyComplete } = props;
  const navigate = useNavigate();

  const onEditClick = useCallback(() => {
    navigate(`./${clientId}/edit`);
  }, [navigate, clientId]);

  const onCopyClick = useCallback(() => {
    copyToClipboard(clientId);

    // Invoke callback
    onCopyComplete();
  }, [clientId, onCopyComplete]);

  const onRemoveClick = useCallback(() => {
    // TODO: to be implemented
  }, []);

  return (
    <div className={styles.clientListColumn}>
      <ActionButton
        className={styles.listActionButton}
        theme={actionButtonTheme}
        onClick={onEditClick}
      >
        <FormattedMessage id="edit" />
      </ActionButton>
      <VerticalDivider className={styles.listActionButtonDivider} />
      <ActionButton
        className={styles.listActionButton}
        theme={actionButtonTheme}
        onClick={onCopyClick}
      >
        <FormattedMessage id="copy" />
      </ActionButton>
      <VerticalDivider className={styles.listActionButtonDivider} />
      <ActionButton
        className={styles.listActionButton}
        theme={actionButtonTheme}
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
  const { appConfig, showNotification } = props;
  const { locale, renderToString } = useContext(Context);
  const navigate = useNavigate();

  const oauthClients = useMemo(() => {
    return appConfig?.oauth?.clients ?? [];
  }, [appConfig]);

  const oauthClientListColumns = useMemo(() => {
    return makeOAuthClientListColumns(renderToString);
  }, [renderToString]);

  const oauthClientListItems: OAuthClientListItem[] = useMemo(() => {
    return oauthClients.map((client) => {
      return {
        name: client.client_id,
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
              onCopyComplete={onClientIdCopied}
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
    [onClientIdCopied]
  );

  return (
    <section className={styles.content}>
      <section className={styles.controlButtons}>
        <ActionButton
          theme={actionButtonTheme}
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
    </section>
  );
};

const OAuthClientConfigurationScreen: React.FC = function OAuthClientConfigurationScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppConfigQuery(appID);

  const [isNotificationVisible, setIsNotificationVisible] = useState(false);
  const [notificationMsg, setNotificationMsg] = useState("");

  const showNotification = useCallback((msg: string) => {
    setIsNotificationVisible(true);
    setNotificationMsg(msg);
  }, []);

  const dismissNotification = useCallback(() => {
    setIsNotificationVisible(false);
  }, []);

  const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
  const appConfig = appConfigNode?.effectiveAppConfig ?? null;

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      {isNotificationVisible && (
        <MessageBar onDismiss={dismissNotification}>
          <p>{notificationMsg}</p>
        </MessageBar>
      )}
      <Text as="h1" className={styles.header}>
        <FormattedMessage id="OAuthClientConfiguration.title" />
      </Text>
      <OAuthClientConfiguration
        appConfig={appConfig}
        showNotification={showNotification}
      />
    </main>
  );
};

export default OAuthClientConfigurationScreen;
