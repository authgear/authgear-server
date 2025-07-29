import React, { useState, useMemo, useCallback } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import { FormattedMessage } from "@oursky/react-messageformat";
import {
  GetClientResourceScopesQuery,
  useGetClientResourceScopesQuery,
} from "../../graphql/adminapi/query/getClientResourceScopes.generated";
import {
  useResourceScopesQueryQuery,
  ResourceScopesQueryQuery,
} from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import { useParams } from "react-router-dom";
import {
  EditApplicationScopesList,
  EditApplicationScopesListItem,
} from "../../components/api-resources/EditApplicationScopesList";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { OAuthClientConfig } from "../../types";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";

const pageSize = 1000;

interface FormState {
  assignedScopes: string[];
}

function EditApplicationScopesScreenContent({
  clientConfig,
  resourceScopesQueryData,
  clientResourceScopesQueryData,
}: {
  clientConfig: OAuthClientConfig;
  resourceScopesQueryData: ResourceScopesQueryQuery;
  clientResourceScopesQueryData: GetClientResourceScopesQuery;
}) {
  const resource =
    resourceScopesQueryData.node &&
    resourceScopesQueryData.node.__typename === "Resource"
      ? resourceScopesQueryData.node
      : null;

  const remoteClientScopes = useMemo(() => {
    return clientResourceScopesQueryData.node?.__typename === "Resource"
      ? clientResourceScopesQueryData.node.scopes?.edges
          ?.map((edge) => edge?.node?.scope)
          .filter((scopeStr): scopeStr is string => !!scopeStr)
      : undefined;
  }, [clientResourceScopesQueryData]);

  const [initialState] = useState<FormState>({
    assignedScopes: remoteClientScopes ?? [],
  });
  const form = useSimpleForm<FormState>({
    defaultState: initialState,
    submit: async () => {
      // TODO
    },
    stateMode: "UpdateInitialStateWithUseEffect",
  });

  const assignedScopes = useMemo((): Set<string> => {
    const scopes = new Set<string>(form.state.assignedScopes);
    return scopes;
  }, [form.state.assignedScopes]);

  const scopes = useMemo((): EditApplicationScopesListItem[] => {
    const allScopes =
      resourceScopesQueryData.node &&
      resourceScopesQueryData.node.__typename === "Resource"
        ? resourceScopesQueryData.node.scopes?.edges
            ?.map((edge) => edge?.node)
            .filter((n) => !!n) ?? []
        : [];
    return allScopes.map((scope) => ({
      scope: scope.scope,
      isAssigned: assignedScopes.has(scope.scope),
    }));
  }, [assignedScopes, resourceScopesQueryData.node]);

  return (
    <APIResourceScreenLayout
      breadcrumbItems={[
        {
          to: "~/api-resources",
          label: <FormattedMessage id="ScreenNav.api-resources" />,
        },
        {
          to: `~/api-resources/${resource?.id}`,
          label: resource?.name ?? resource?.resourceURI ?? "",
        },
        {
          to: ``,
          label: clientConfig.name ?? clientConfig.client_name,
        },
      ]}
    >
      <FormContainerBase form={form}>
        <EditApplicationScopesList
          className="flex-1 min-h-0 col-span-full"
          scopes={scopes}
          onToggleAssignedScopes={useCallback(
            (
              updatedScopes: EditApplicationScopesListItem[],
              isAssigned: boolean
            ) => {
              form.setState((state) => {
                const currentAssignedScopes = state.assignedScopes;
                const newSet = new Set(currentAssignedScopes);

                updatedScopes.forEach((scopeItem) => {
                  if (isAssigned) {
                    newSet.add(scopeItem.scope);
                  } else {
                    newSet.delete(scopeItem.scope);
                  }
                });

                return {
                  assignedScopes: Array.from(newSet),
                };
              });
            },
            [form]
          )}
        />
      </FormContainerBase>
    </APIResourceScreenLayout>
  );
}

const EditApplicationScopesScreen: React.VFC =
  function EditApplicationScopesScreen() {
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
    });

    const clientScopesQuery = useGetClientResourceScopesQuery({
      variables: {
        clientID: clientID!,
        resourceID: resourceID!,
        first: pageSize,
      },
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
        if (
          client == null ||
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
          />
        );
      },
    });
  };

export default EditApplicationScopesScreen;
