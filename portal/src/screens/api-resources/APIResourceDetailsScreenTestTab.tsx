import React, { useContext, useMemo } from "react";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import WidgetTitle from "../../WidgetTitle";
import { Text, Dropdown } from "@fluentui/react";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";
import { PortalAPIAppConfig } from "../../types";

export function APIResourceDetailsScreenTestTab({
  resource,
}: {
  resource: Resource;
}): React.ReactElement | null {
  const { appID } = useParams() as { appID: string };
  const appConfigQuery = useAppAndSecretConfigQuery(appID);

  return useLoadableView({
    loadables: [appConfigQuery] as const,
    render: ([{ effectiveAppConfig }]) => {
      return (
        <APIResourceDetailsScreenTestTabContent
          resource={resource}
          appConfig={effectiveAppConfig!}
        />
      );
    },
  });
}

function APIResourceDetailsScreenTestTabContent({
  resource,
  appConfig,
}: {
  resource: Resource;
  appConfig: PortalAPIAppConfig;
}) {
  const { renderToString } = useContext(MessageContext);

  const authorizedApplicationsOptions = useMemo(() => {
    const authorizedClientIDs = new Set(resource.clientIDs);
    return (
      appConfig.oauth?.clients
        ?.filter((clientConfig) => {
          return authorizedClientIDs.has(clientConfig.client_id);
        })
        .map((clientConfig) => ({
          key: clientConfig.client_id,
          text: clientConfig.name ?? clientConfig.client_name ?? "",
        })) ?? []
    );
  }, [appConfig.oauth?.clients, resource.clientIDs]);

  return (
    <div className="pt-5 flex-1 flex flex-col space-y-4">
      <header className="space-y-2">
        <WidgetTitle>
          <FormattedMessage id="APIResourceDetailsScreen.tab.test" />
        </WidgetTitle>
        <Text block={true}>
          <FormattedMessage id="APIResourceDetailsScreen.test.description" />
        </Text>
      </header>
      <div className="flex-1 flex flex-col max-w-180">
        <Dropdown
          label={renderToString(
            "APIResourceDetailsScreen.authorizedApplications"
          )}
          placeholder={
            authorizedApplicationsOptions.length === 0
              ? renderToString(
                  "APIResourceDetailsScreen.selectApplication.empty"
                )
              : renderToString("APIResourceDetailsScreen.selectApplication")
          }
          options={authorizedApplicationsOptions}
          disabled={authorizedApplicationsOptions.length === 0}
        />
      </div>
    </div>
  );
}
