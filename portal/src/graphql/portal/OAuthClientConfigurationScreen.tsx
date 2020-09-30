import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { DateTime } from "luxon";
import {
  DetailsList,
  IColumn,
  IconButton,
  MessageBar,
  PrimaryButton,
  SelectionMode,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { formatDatetime } from "../../util/formatDatetime";

import styles from "./OAuthClientConfigurationScreen.module.scss";

interface OAuthClientConfigurationProps {
  appConfig: PortalAPIAppConfig | null;
  showNotification: (msg: string) => void;
}

interface OAuthClientListItem {
  name: string;
  creationDate: string;
  clientId: string;
}

interface OAuthClientIdCellProps {
  clientId: string;
  onCopyComplete: () => void;
}

const ADD_CLIENT_BUTTON_STYLES = {
  icon: { paddingRight: "4px" },
  flexContainer: { paddingRight: "2px" },
};

const ICON_BUTTON_STYLES = { flexContainer: { color: "#504e4c" } };

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
      key: "creationDate",
      fieldName: "creationDate",
      name: renderToString(
        "OAuthClientConfiguration.client-list.creation-date"
      ),
      minWidth: 150,
      className: styles.clientListColumn,
    },

    {
      key: "clientId",
      fieldName: "clientId",
      name: renderToString("OAuthClientConfiguration.client-list.client-id"),
      minWidth: 350,
    },
  ];
}

const OAuthClientIdCell: React.FC<OAuthClientIdCellProps> = function OAuthClientIdCell(
  props: OAuthClientIdCellProps
) {
  const { clientId, onCopyComplete } = props;
  const navigate = useNavigate();

  const onEditClick = useCallback(() => {
    navigate(`./${clientId}/edit`);
  }, [navigate, clientId]);

  const onCopyClick = useCallback(() => {
    const el = document.createElement("textarea");
    el.value = clientId;
    // Set non-editable to avoid focus and move outside of view
    el.setAttribute("readonly", "");
    el.setAttribute("style", "position: absolute; left: -9999px");
    document.body.appendChild(el);
    // Select text inside element
    el.select();
    el.setSelectionRange(0, 100); // for mobile device
    document.execCommand("copy");
    // Remove temporary element
    document.body.removeChild(el);

    // Invoke callback
    onCopyComplete();
  }, [clientId, onCopyComplete]);

  return (
    <div className={styles.clientListColumn}>
      <span
        className={cn(
          styles.clientListColumnContent,
          styles.clientIdColumnContent
        )}
      >
        {clientId}
      </span>
      <IconButton
        className={styles.editButton}
        styles={ICON_BUTTON_STYLES}
        onClick={onEditClick}
        iconProps={{ iconName: "Edit" }}
      />
      <IconButton
        className={styles.copyButton}
        styles={ICON_BUTTON_STYLES}
        onClick={onCopyClick}
        iconProps={{ iconName: "Copy" }}
      />
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
      // TODO: replace with actual data
      const creationDateString =
        formatDatetime(locale, "1970-01-01T00:00:00.000Z", DateTime.DATE_MED) ??
        "---";
      return {
        name: client.client_id,
        creationDate: creationDateString,
        clientId: client.client_id,
      };
    });
  }, [oauthClients, locale]);

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
        case "clientId":
          return (
            <OAuthClientIdCell
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
        <PrimaryButton
          className={styles.addClientButton}
          onClick={onAddOAuthClientClick}
          iconProps={{ iconName: "CirclePlus" }}
          styles={ADD_CLIENT_BUTTON_STYLES}
        >
          <FormattedMessage id="OAuthClientConfiguration.add-client-button" />
        </PrimaryButton>
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
