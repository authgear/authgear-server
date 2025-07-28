import React, { useMemo, useCallback, useState, useContext } from "react";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import WidgetTitle from "../../WidgetTitle";
import { SearchBox, Text } from "@fluentui/react";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useAddResourceToClientIdMutation } from "../../graphql/adminapi/mutations/addResourceToClientID.generated";
import { useRemoveResourceFromClientIdMutation } from "../../graphql/adminapi/mutations/removeResourceFromClientID.generated";
import {
  ApplicationList,
  ApplicationListItem,
} from "../../components/api-resources/ApplicationList";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useParams } from "react-router-dom";
import ShowError from "../../ShowError";
import {
  ResourceQueryDocument,
  ResourceQueryQuery,
} from "../../graphql/adminapi/query/resourceQuery.generated";
import { parseRawError } from "../../error/parse";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";

export function APIResourceDetailsScreenApplicationsTab({
  resource,
}: {
  resource: Resource;
}): JSX.Element {
  const { appID } = useParams() as { appID: string };
  const appConfigQuery = useAppAndSecretConfigQuery(appID);
  const [addResource] = useAddResourceToClientIdMutation();
  const [removeResource] = useRemoveResourceFromClientIdMutation();
  const { setErrors } = useErrorMessageBarContext();
  const { renderToString } = useContext(MessageContext);

  const isLoading = appConfigQuery.isLoading;
  const [searchKeyword, setSearchKeyword] = useState("");

  const applications = useMemo((): ApplicationListItem[] => {
    return (
      appConfigQuery.effectiveAppConfig?.oauth?.clients
        ?.filter((clientConfig) => {
          switch (clientConfig.x_application_type) {
            case "m2m":
            case "confidential":
              return true;
            default:
              return false;
          }
        })
        .map((clientConfig) => ({
          clientID: clientConfig.client_id,
          authorized: resource.clientIDs.includes(clientConfig.client_id),
          name: clientConfig.name ?? clientConfig.client_name ?? "",
        })) ?? []
    );
  }, [appConfigQuery.effectiveAppConfig?.oauth?.clients, resource.clientIDs]);

  const filteredApplications = useMemo(() => {
    if (!searchKeyword) {
      return applications;
    }
    return applications.filter((app) =>
      app.name.toLowerCase().includes(searchKeyword.toLowerCase())
    );
  }, [applications, searchKeyword]);

  const onSearchQueryChange = useCallback(
    (
      _event: React.ChangeEvent<HTMLInputElement> | undefined,
      newValue: string | undefined
    ) => {
      setSearchKeyword(newValue ?? "");
    },
    []
  );

  const onToggleAuthorized = useCallback(
    async (item: ApplicationListItem, checked: boolean) => {
      try {
        if (checked) {
          const newResource = {
            ...resource,
            clientIDs: [...resource.clientIDs, item.clientID],
          };
          await addResource({
            variables: {
              clientID: item.clientID,
              resourceURI: resource.resourceURI,
            },
            refetchQueries: [ResourceQueryDocument],
            awaitRefetchQueries: true,
            optimisticResponse: {
              addResourceToClientID: {
                resource: newResource,
              },
            },
            update: (cache) => {
              // optimistic update
              cache.writeQuery<ResourceQueryQuery>({
                query: ResourceQueryDocument,
                variables: { id: resource.id },
                data: { node: newResource },
              });
            },
          });
        } else {
          const newResource = {
            ...resource,
            clientIDs: resource.clientIDs.filter(
              (clientID) => clientID !== item.clientID
            ),
          };
          await removeResource({
            variables: {
              clientID: item.clientID,
              resourceURI: resource.resourceURI,
            },
            refetchQueries: [ResourceQueryDocument],
            awaitRefetchQueries: true,
            optimisticResponse: {
              removeResourceFromClientID: {
                resource: newResource,
              },
            },
            update: (cache) => {
              // optimistic update
              cache.writeQuery<ResourceQueryQuery>({
                query: ResourceQueryDocument,
                variables: { id: resource.id },
                data: { node: newResource },
              });
            },
          });
        }
      } catch (e: unknown) {
        setErrors(parseRawError(e));
      }
    },
    [resource, addResource, removeResource, setErrors]
  );

  if (appConfigQuery.loadError) {
    return <ShowError error={appConfigQuery.loadError} />;
  }

  return (
    <div className="pt-5 flex-1 flex flex-col space-y-4">
      <header>
        <WidgetTitle className="mb-2">
          <FormattedMessage id="APIResourceDetailsScreen.tab.applications" />
        </WidgetTitle>
        <Text>
          <FormattedMessage id="APIResourceDetailsScreen.applications.description" />
        </Text>
      </header>
      <SearchBox
        onChange={onSearchQueryChange}
        styles={{ root: { width: 300 } }}
        value={searchKeyword}
        placeholder={renderToString("search")}
      />
      <div className="flex-1 flex flex-col max-w-180">
        <ApplicationList
          applications={filteredApplications}
          className="flex-1 min-h-0"
          loading={isLoading}
          onToggleAuthorized={onToggleAuthorized}
        />
      </div>
    </div>
  );
}
