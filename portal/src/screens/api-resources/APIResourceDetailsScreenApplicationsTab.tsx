import React, { useMemo } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import WidgetTitle from "../../WidgetTitle";
import { Text } from "@fluentui/react";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import {
  ApplicationList,
  ApplicationListItem,
} from "../../components/api-resources/ApplicationList";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useParams } from "react-router-dom";
import ShowError from "../../ShowError";

export function APIResourceDetailsScreenApplicationsTab({}: {
  resource: Resource;
}): JSX.Element {
  const { appID } = useParams() as { appID: string };
  const appConfigQuery = useAppAndSecretConfigQuery(appID);

  const isLoading = appConfigQuery.isLoading;

  const applications = useMemo((): ApplicationListItem[] => {
    return (
      appConfigQuery.effectiveAppConfig?.oauth?.clients?.map(
        (clientConfig) => ({
          clientID: clientConfig.client_id,
          authorized: true,
          name: clientConfig.name ?? clientConfig.client_name ?? "",
        })
      ) ?? []
    );
  }, [appConfigQuery.effectiveAppConfig?.oauth?.clients]);

  if (appConfigQuery.loadError) {
    return <ShowError error={appConfigQuery.loadError} />;
  }

  return (
    <div className="pt-5 flex-1 flex flex-col space-y-2">
      <header>
        <WidgetTitle className="mb-2">
          <FormattedMessage id="APIResourceDetailsScreen.tab.applications" />
        </WidgetTitle>
        <Text>
          <FormattedMessage id="APIResourceDetailsScreen.applications.description" />
        </Text>
      </header>
      <div className="flex-1 flex flex-col max-w-180">
        <ApplicationList
          applications={applications}
          className="flex-1 min-h-0"
          loading={isLoading}
        />
      </div>
    </div>
  );
}
