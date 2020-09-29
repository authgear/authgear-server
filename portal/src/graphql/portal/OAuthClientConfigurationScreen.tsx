import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { DateTime } from "luxon";
import {
  DetailsList,
  IColumn,
  IconButton,
  IconNames,
  IDetailsList,
  IDetailsListProps,
  IRenderFunction,
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
}

type ApplicationType = "ios" | "android" | "web-app";
interface OAuthClientListItem {
  name: string;
  creationDate: string;
  applicationType: string;
  clientId: string;
}

const ADD_CLIENT_BUTTON_STYLES = {
  icon: { paddingRight: "4px" },
  flexContainer: { paddingRight: "2px" },
};

const ICON_BUTTON_STYLES = { flexContainer: { color: "#504e4c" } };

const applicationTypeMessageID: Record<ApplicationType, string> = {
  ios: "OAuthClientConfiguration.application-type.ios",
  android: "OAuthClientConfiguration.application-type.android",
  "web-app": "OAuthClientConfiguration.application-type.web-app",
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
      key: "creationDate",
      fieldName: "creationDate",
      name: renderToString(
        "OAuthClientConfiguration.client-list.creation-date"
      ),
      minWidth: 150,
      className: styles.clientListColumn,
    },

    {
      key: "applicationType",
      fieldName: "applicationType",
      name: renderToString(
        "OAuthClientConfiguration.client-list.application-type"
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

const OAuthClientConfiguration: React.FC<OAuthClientConfigurationProps> = function OAuthClientConfiguration(
  props: OAuthClientConfigurationProps
) {
  const { appConfig } = props;
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
        // TODO: replace with actual data
        applicationType: renderToString(applicationTypeMessageID["ios"]),
        clientId: client.client_id,
      };
    });
  }, [oauthClients, locale, renderToString]);

  const onAddOAuthClientClick = useCallback(() => {
    navigate("./add");
  }, [navigate]);

  const onEditOAuthClientClick = useCallback(
    (clientId: string) => {
      navigate(`./${clientId}/edit`);
    },
    [navigate]
  );

  const onCopyClientIdClick = useCallback((_clientId: string) => {
    // TODO: to be implemented
  }, []);

  const onRenderOAuthClientColumns = useCallback(
    (item?: OAuthClientListItem, _index?: number, column?: IColumn) => {
      if (item == null || column == null) {
        return null;
      }
      const fieldContent = item[column.fieldName as keyof OAuthClientListItem];
      switch (column.key) {
        case "clientId":
          return (
            <div className={styles.clientListColumn}>
              <span
                className={cn(
                  styles.clientListColumnContent,
                  styles.clientIdColumnContent
                )}
              >
                {fieldContent}
              </span>
              <IconButton
                className={styles.editButton}
                styles={ICON_BUTTON_STYLES}
                onClick={() => onEditOAuthClientClick(item.clientId)}
                iconProps={{ iconName: "Edit" }}
              />
              <IconButton
                className={styles.copyButton}
                styles={ICON_BUTTON_STYLES}
                onClick={() => onCopyClientIdClick(item.clientId)}
                iconProps={{ iconName: "Copy" }}
              />
            </div>
          );
        default:
          return (
            <span className={styles.clientListColumnContent}>
              {fieldContent}
            </span>
          );
      }
    },
    [onCopyClientIdClick, onEditOAuthClientClick]
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
      <Text as="h1" className={styles.header}>
        <FormattedMessage id="OAuthClientConfiguration.title" />
      </Text>
      <OAuthClientConfiguration appConfig={appConfig} />
    </main>
  );
};

export default OAuthClientConfigurationScreen;
