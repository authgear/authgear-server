import React, { useMemo, useCallback } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormattedMessage } from "../../intl";
import {
  GetClientResourceScopesDocument,
  GetClientResourceScopesQuery,
  useGetClientResourceScopesQuery,
} from "../../graphql/adminapi/query/getClientResourceScopes.generated";
import {
  useResourceScopesQueryQuery,
  ResourceScopesQueryQuery,
} from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import { useReplaceScopesOfClientIdMutation } from "../../graphql/adminapi/mutations/replaceScopesOfClientID.generated";
import { useParams } from "react-router-dom";
import {
  EditApplicationScopesList,
  EditApplicationScopesListItem,
} from "../../components/api-resources/EditApplicationScopesList";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { OAuthClientConfig } from "../../types";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";
import FormContainer from "../../FormContainer";
import { BreadcrumbItem } from "../../NavBreadcrumb";

const pageSize = 1000;

interface FormState {
  assignedScopes: string[];
}

export function EditApplicationScopesScreenContent({
  clientConfig,
  resourceScopesQueryData,
  clientResourceScopesQueryData,
  breadcrumbItems,
}: {
  clientConfig: OAuthClientConfig;
  resourceScopesQueryData: ResourceScopesQueryQuery;
  clientResourceScopesQueryData: GetClientResourceScopesQuery;
  breadcrumbItems: BreadcrumbItem[];
}): React.ReactElement {
  const [replaceScopesOfClientIdMutation] =
    useReplaceScopesOfClientIdMutation();

  const resource =
    resourceScopesQueryData.node?.__typename === "Resource"
      ? resourceScopesQueryData.node
      : null;

  if (resource == null) {
    throw new Error("unexpected type of node");
  }

  const remoteClientScopes = useMemo(() => {
    return clientResourceScopesQueryData.node?.__typename === "Resource"
      ? clientResourceScopesQueryData.node.scopes?.edges
          ?.map((edge) => edge?.node?.scope)
          .filter((scopeStr): scopeStr is string => !!scopeStr)
      : undefined;
  }, [clientResourceScopesQueryData]);

  const initialState = useMemo(
    () => ({
      assignedScopes: remoteClientScopes ?? [],
    }),
    [remoteClientScopes]
  );

  const form = useSimpleForm<FormState>({
    defaultState: initialState,
    submit: async () => {
      await replaceScopesOfClientIdMutation({
        variables: {
          clientID: clientConfig.client_id,
          resourceURI: resource.resourceURI,
          scopes: form.state.assignedScopes,
        },
        awaitRefetchQueries: true,
        refetchQueries: [GetClientResourceScopesDocument],
      });
    },
    stateMode: "UpdateInitialStateWithUseEffect",
  });

  const assignedScopes = useMemo((): Set<string> => {
    const scopes = new Set<string>(form.state.assignedScopes);
    return scopes;
  }, [form.state.assignedScopes]);

  const scopes = useMemo((): EditApplicationScopesListItem[] => {
    const allScopes =
      resource.scopes?.edges?.map((edge) => edge?.node).filter((n) => !!n) ??
      [];
    return allScopes.map((scope) => ({
      scope: scope.scope,
      isAssigned: assignedScopes.has(scope.scope),
    }));
  }, [assignedScopes, resource]);

  return (
    <FormContainer
      form={form}
      className="flex-1-0-auto flex flex-col"
      stickyFooterComponent={true}
      showDiscardButton={true}
    >
      <APIResourceScreenLayout breadcrumbItems={breadcrumbItems}>
        <EditApplicationScopesList
          className="flex-1-0-auto col-span-full"
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
      </APIResourceScreenLayout>
    </FormContainer>
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
                to: "~/api-resources",
                label: <FormattedMessage id="ScreenNav.api-resources" />,
              },
              {
                to: `~/api-resources/${resource.id}#applications`,
                label: resource.name ?? resource.resourceURI,
              },
              {
                to: ``,
                label: client.name ?? client.client_name,
              },
            ]}
          />
        );
      },
    });
  };

export default EditApplicationScopesScreen;
