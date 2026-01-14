import React, { useContext, useMemo } from "react";
import styles from "./App2AppConfigurationScreen.module.css";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import {
  Context as IntlContext,
  FormattedMessage,
} from "../../intl";
import ScreenDescription from "../../ScreenDescription";
import { useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { PortalAPIAppConfig } from "../../types";
import { DetailsList, IColumn, SelectionMode, Text } from "@fluentui/react";
import Link from "../../Link";

interface App2AppRowViewModel {
  cliendID: string;
  name: string;
  isEnabled: boolean;
}

function App2AppConfigurationScreenLoaded(props: {
  appID: string;
  effectiveAppConfig: PortalAPIAppConfig;
}) {
  const { appID, effectiveAppConfig } = props;
  const { renderToString } = useContext(IntlContext);

  const columns = useMemo((): IColumn[] => {
    return [
      {
        key: "name",
        fieldName: "name",
        name: renderToString("App2AppConfigurationScreen.columns.name"),
        minWidth: 100,
      },
      {
        key: "status",
        fieldName: "isEnabled",
        name: renderToString("App2AppConfigurationScreen.columns.status"),
        minWidth: 100,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: App2AppRowViewModel) => {
          if (item.isEnabled) {
            return (
              <span className="text-status-green">
                <FormattedMessage id="App2AppConfigurationScreen.status.enabled"></FormattedMessage>
              </span>
            );
          }
          return (
            <span className="text-status-grey">
              <FormattedMessage id="App2AppConfigurationScreen.status.disabled"></FormattedMessage>
            </span>
          );
        },
      },
      {
        key: "action",
        name: renderToString("App2AppConfigurationScreen.columns.action"),
        minWidth: 150,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item: App2AppRowViewModel) => {
          return (
            <Link
              to={`/project/${appID}/configuration/apps/${item.cliendID}/edit#app2app`}
            >
              <FormattedMessage id="App2AppConfigurationScreen.action.setup"></FormattedMessage>
            </Link>
          );
        },
      },
    ];
  }, [appID, renderToString]);

  const rows = useMemo((): App2AppRowViewModel[] => {
    return (
      effectiveAppConfig.oauth?.clients
        ?.filter((client) => client.x_application_type === "native")
        .map((client) => ({
          cliendID: client.client_id,
          name: client.name ?? client.client_id,
          isEnabled: client.x_app2app_enabled ? true : false,
        })) ?? []
    );
  }, [effectiveAppConfig.oauth?.clients]);

  return (
    <ScreenContent>
      <ScreenTitle className={styles.widget}>
        <FormattedMessage id="App2AppConfigurationScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="App2AppConfigurationScreen.description" />
      </ScreenDescription>
      <div className={styles.widget}>
        <Text block={true}>
          <FormattedMessage id="App2AppConfigurationScreen.table.description" />
        </Text>
        <DetailsList
          className={styles.clientList}
          columns={columns}
          items={rows}
          selectionMode={SelectionMode.none}
        />
      </div>
    </ScreenContent>
  );
}

export default function App2AppConfigurationScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };

  const {
    isLoading,
    loadError,
    effectiveAppConfig,
    refetch: reload,
  } = useAppAndSecretConfigQuery(appID);

  if (isLoading) {
    return <ShowLoading />;
  }

  if (loadError) {
    return <ShowError error={loadError} onRetry={reload} />;
  }

  if (effectiveAppConfig != null) {
    return (
      <App2AppConfigurationScreenLoaded
        appID={appID}
        effectiveAppConfig={effectiveAppConfig}
      />
    );
  }

  return <></>;
}
