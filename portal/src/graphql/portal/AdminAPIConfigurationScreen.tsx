import React, { useContext, useMemo, useCallback } from "react";
import { useParams } from "react-router-dom";
import {
  DetailsList,
  IColumn,
  SelectionMode,
  ActionButton,
  MessageBar,
  MessageBarType,
  TextField,
  PrimaryButton,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  useAppAndSecretConfigQuery,
  AppAndSecretConfigQueryResult,
} from "./query/appAndSecretConfigQuery";
import { formatDatetime } from "../../util/formatDatetime";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { downloadStringAsFile } from "../../util/download";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { makeGraphQLEndpoint } from "../adminapi/apollo";
import styles from "./AdminAPIConfigurationScreen.module.scss";
import { useCopyFeedback } from "../../hook/useCopyFeedback";

interface AdminAPIConfigurationScreenContentProps {
  appID: string;
  queryResult: AppAndSecretConfigQueryResult;
}

interface Item {
  keyID: string;
  createdAt: string | null;
  publicKeyPEM: string;
  privateKeyPEM?: string | null;
}

interface LocationState {
  keyID: string;
}

const messageBarStyles = {
  root: {
    width: "auto",
  },
};

const EXAMPLE_QUERY = `# The GraphQL schema follows the Relay GraphQL Server convention.
# If you find the terms like "Node", "Edge", "Connection" strange to you, you can learn about them
# at the related documentation of Relay at https://relay.dev/docs/guides/graphql-server-specification/
#
# For those who are curious, this is a more formal documentation https://relay.dev/assets/files/connections-932f4f2cdffd79724ac76373deb30dc8.htm
#
# Here is an example query of fetching a list of users with page size equal to 2.
query {
  users(first: 2) {
    pageInfo {
      hasNextPage
    }
    edges {
      cursor
      node {
        id
        createdAt
      }
    }
  }
}
`;

const AdminAPIConfigurationScreenContent: React.FC<AdminAPIConfigurationScreenContentProps> =
  function AdminAPIConfigurationScreenContent(props) {
    const { appID, queryResult } = props;
    const { locale, renderToString } = useContext(Context);
    const { effectiveAppConfig } = useAppAndSecretConfigQuery(appID);
    const { appHostSuffix, themes } = useSystemConfig();

    const rawAppID = effectiveAppConfig?.id;
    const adminAPIEndpoint =
      rawAppID != null
        ? "https://" + rawAppID + appHostSuffix + "/_api/admin/graphql"
        : "";

    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: adminAPIEndpoint,
    });

    const graphqlEndpoint = useMemo(() => {
      const base = makeGraphQLEndpoint(appID);
      return base + "?query=" + encodeURIComponent(EXAMPLE_QUERY);
    }, [appID]);

    const adminAPISecrets = useMemo(() => {
      return queryResult.secretConfig?.adminAPISecrets ?? [];
    }, [queryResult.secretConfig?.adminAPISecrets]);

    const items: Item[] = useMemo(() => {
      const items = [];
      for (const adminAPISecret of adminAPISecrets) {
        items.push({
          keyID: adminAPISecret.keyID,
          createdAt: formatDatetime(locale, adminAPISecret.createdAt),
          publicKeyPEM: adminAPISecret.publicKeyPEM,
          privateKeyPEM: adminAPISecret.privateKeyPEM,
        });
      }
      return items;
    }, [locale, adminAPISecrets]);

    const downloadItem = useCallback(
      (keyID: string) => {
        const item = items.find((a) => a.keyID === keyID);
        if (item == null) {
          return;
        }
        if (item.privateKeyPEM != null) {
          downloadStringAsFile({
            content: item.privateKeyPEM,
            mimeType: "application/x-pem-file",
            filename: `${item.keyID}.pem`,
          });
          return;
        }

        const state: LocationState = {
          keyID,
        };

        startReauthentication(state).catch((e) => {
          console.error(e);
        });
      },
      [items]
    );

    useLocationEffect((state: LocationState) => {
      downloadItem(state.keyID);
    });

    const actionColumnOnRender = useCallback(
      (item?: Item) => {
        return (
          <ActionButton
            className={styles.actionButton}
            theme={themes.actionButton}
            onClick={(e: React.MouseEvent<unknown>) => {
              e.preventDefault();
              e.stopPropagation();
              if (item != null) {
                downloadItem(item.keyID);
              }
            }}
          >
            <FormattedMessage id="download" />
          </ActionButton>
        );
      },
      [downloadItem, themes.actionButton]
    );

    const columns: IColumn[] = useMemo(() => {
      return [
        {
          key: "keyID",
          fieldName: "keyID",
          name: renderToString("AdminAPIConfigurationScreen.column.key-id"),
          minWidth: 150,
        },
        {
          key: "createdAt",
          fieldName: "createdAt",
          name: renderToString("AdminAPIConfigurationScreen.column.created-at"),
          minWidth: 150,
        },
        {
          key: "action",
          name: renderToString("action"),
          minWidth: 150,
          onRender: actionColumnOnRender,
        },
      ];
    }, [renderToString, actionColumnOnRender]);

    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="AdminAPIConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="AdminAPIConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AdminAPIConfigurationScreen.api-endpoint.title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="AdminAPIConfigurationScreen.api-endpoint.description" />
          </WidgetDescription>
          <div className={styles.copyButtonGroup}>
            <TextField
              type="text"
              readOnly={true}
              value={adminAPIEndpoint}
              className={styles.copyTextField}
            />
            <PrimaryButton
              {...copyButtonProps}
              className={styles.copyButton}
              iconProps={undefined}
            />
            <Feedback />
          </div>
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AdminAPIConfigurationScreen.graphiql.title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage
              id="AdminAPIConfigurationScreen.graphiql.description"
              values={{ graphqlEndpoint }}
            />
          </WidgetDescription>
          <MessageBar
            className={styles.messageBar}
            messageBarType={MessageBarType.warning}
            styles={messageBarStyles}
          >
            <FormattedMessage id="AdminAPIConfigurationScreen.graphiql.warning" />
          </MessageBar>
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="AdminAPIConfigurationScreen.keys.title" />
          </WidgetTitle>
          <DetailsList
            items={items}
            columns={columns}
            selectionMode={SelectionMode.none}
          />
        </Widget>
      </ScreenContent>
    );
  };

const AdminAPIConfigurationScreen: React.FC =
  function AdminAPIConfigurationScreen() {
    const { appID } = useParams();
    const queryResult = useAppAndSecretConfigQuery(appID);

    if (queryResult.loading) {
      return <ShowLoading />;
    }

    if (queryResult.error) {
      return (
        <ShowError error={queryResult.error} onRetry={queryResult.refetch} />
      );
    }

    return (
      <AdminAPIConfigurationScreenContent
        appID={appID}
        queryResult={queryResult}
      />
    );
  };

export default AdminAPIConfigurationScreen;
