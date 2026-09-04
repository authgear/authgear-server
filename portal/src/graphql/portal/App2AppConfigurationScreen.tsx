import React, { useMemo } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import styles from "./App2AppConfigurationScreen.module.css";
import ScreenContent from "../../ScreenContent";
import { FormattedMessage } from "../../intl";
import { useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { PortalAPIAppConfig } from "../../types";
import { CardTable } from "../../components/v2/CardTable/CardTable";
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
      <Text
        as="p"
        size="5"
        weight="bold"
        className={cn(styles.widget, styles.pageTitle)}
      >
        <FormattedMessage id="App2AppConfigurationScreen.title" />
      </Text>
      <Text
        as="p"
        size="2"
        className={cn(styles.widget, styles.pageDescription)}
      >
        <FormattedMessage id="App2AppConfigurationScreen.description" />
      </Text>
      <div className={cn(styles.widget, styles.listSection)}>
        <Text as="p" size="2" className={styles.tableDescription}>
          <FormattedMessage id="App2AppConfigurationScreen.table.description" />
        </Text>
        <CardTable>
          <CardTable.Header>
            <CardTable.HeaderCell className={styles.colName}>
              <FormattedMessage id="App2AppConfigurationScreen.columns.name" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colStatus}>
              <FormattedMessage id="App2AppConfigurationScreen.columns.status" />
            </CardTable.HeaderCell>
            <CardTable.HeaderCell className={styles.colAction}>
              <FormattedMessage id="App2AppConfigurationScreen.columns.action" />
            </CardTable.HeaderCell>
          </CardTable.Header>
          {rows.map((row) => (
            <CardTable.Row key={row.cliendID}>
              <CardTable.Cell className={styles.colName}>
                <Text size="2" className={styles.clientName}>
                  {row.name}
                </Text>
              </CardTable.Cell>
              <CardTable.Cell className={styles.colStatus}>
                {row.isEnabled ? (
                  <span className={styles.statusEnabled}>
                    <FormattedMessage id="App2AppConfigurationScreen.status.enabled" />
                  </span>
                ) : (
                  <span className={styles.statusDisabled}>
                    <FormattedMessage id="App2AppConfigurationScreen.status.disabled" />
                  </span>
                )}
              </CardTable.Cell>
              <CardTable.Cell className={styles.colAction}>
                <Link
                  to={`/project/${appID}/configuration/apps/${row.cliendID}/edit#app2app`}
                >
                  <FormattedMessage id="App2AppConfigurationScreen.action.setup" />
                </Link>
              </CardTable.Cell>
            </CardTable.Row>
          ))}
        </CardTable>
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
    // eslint-disable-next-line @typescript-eslint/strict-void-return
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
