import React from "react";
import { useParams } from "react-router-dom";
import { useGetClientResourceScopesQuery } from "../../graphql/adminapi/query/getClientResourceScopes.generated";
import { useResourceScopesQueryQuery } from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";
import { EditApplicationScopesScreenContent } from "./EditApplicationScopesScreen";
import { FormattedMessage } from "@oursky/react-messageformat";

const pageSize = 1000;

const EditOAuthClientResourceScopeScreen: React.VFC =
  function EditOAuthClientResourceScopeScreen() {
    const { appID, clientID, resourceID } = useParams<{
      appID: string;
      clientID: string;
      resourceID: string;
    }>();

    const appConfigQuery = useAppAndSecretConfigQuery(appID ?? "");

    const scopesQuery = useResourceScopesQueryQuery({
      variables: {
        resourceID: resourceID!,
        first: pageSize,
      },
      fetchPolicy: "network-only",
    });

    const clientScopesQuery = useGetClientResourceScopesQuery({
      variables: {
        clientID: clientID!,
        resourceID: resourceID!,
        first: pageSize,
      },
      fetchPolicy: "network-only",
    });

    const scopesQueryLoadable = {
      isLoading: scopesQuery.loading,
      reload: scopesQuery.refetch,
      loadError: scopesQuery.error,
      data: scopesQuery.data,
    };

    const clientScopesQueryLoadable = {
      isLoading: clientScopesQuery.loading,
      reload: clientScopesQuery.refetch,
      loadError: clientScopesQuery.error,
      data: clientScopesQuery.data,
    };

    return useLoadableView({
      loadables: [
        appConfigQuery,
        scopesQueryLoadable,
        clientScopesQueryLoadable,
      ] as const,
      render: ([
        { effectiveAppConfig },
        { data: scopesQueryData },
        { data: clientScopesQueryData },
      ]) => {
        const client = effectiveAppConfig?.oauth?.clients?.find(
          (c) => c.client_id === clientID
        );
        const resource =
          scopesQueryData?.node?.__typename === "Resource"
            ? scopesQueryData.node
            : null;
        if (
          client == null ||
          resource == null ||
          scopesQueryData == null ||
          clientScopesQueryData == null
        ) {
          return null;
        }

        return (
          <EditApplicationScopesScreenContent
            clientConfig={client}
            resourceScopesQueryData={scopesQueryData}
            clientResourceScopesQueryData={clientScopesQueryData}
            breadcrumbItems={[
              {
                to: "~/configuration/apps",
                label: <FormattedMessage id="ScreenNav.client-applications" />,
              },
              {
                to: `~/configuration/apps/${client.client_id}/edit?tab=api-resources`,
                label: client.name ?? client.client_name,
              },
              {
                to: ``,
                label: resource.name ?? resource.resourceURI,
              },
            ]}
          />
        );
      },
    });
  };

export default EditOAuthClientResourceScopeScreen;
