import React, { useContext, useState, useCallback, useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import cn from "classnames";
import styles from "./EditOAuthClientFormResourcesContent.module.css";
import WidgetTitle from "../../WidgetTitle";
import {
  Context as MessageContext,
  FormattedMessage,
} from "../../intl";
import { SearchBox } from "@fluentui/react/lib/SearchBox";
import {
  useResourcesQueryQuery,
  ResourcesQueryDocument,
} from "../adminapi/query/resourcesQuery.generated";
import {
  ApplicationResourcesList,
  ApplicationResourceListItem,
} from "../../components/api-resources/ApplicationResourcesList";
import { UnauthorizeApplicationDialog } from "../../components/api-resources/UnauthorizeApplicationDialog";
import { OAuthClientConfig } from "../../types";
import { encodeOffsetToCursor } from "../../util/pagination";
import { PaginationProps } from "../../PaginationWidget";
import { useDebounced } from "../../hook/useDebounced";
import ShowError from "../../ShowError";
import { useAddResourceToClientIdMutation } from "../adminapi/mutations/addResourceToClientID.generated";
import { useRemoveResourceFromClientIdMutation } from "../adminapi/mutations/removeResourceFromClientID.generated";
import { parseRawError } from "../../error/parse";
import { produce } from "immer";
import { useErrorState } from "../../hook/error";
import { Resource } from "../adminapi/globalTypes.generated";

const PAGE_SIZE = 10;

export const EditOAuthClientFormResourcesContent: React.FC<{
  className?: string;
  client: OAuthClientConfig;
}> = ({ className, client }) => {
  const { renderToString } = useContext(MessageContext);
  const [_, setError] = useErrorState();
  const [searchKeyword, setSearchKeyword] = useState("");
  const [offset, setOffset] = useState(0);
  const [resourceToUnauthorize, setResourceToUnauthorize] =
    useState<Resource | null>(null);

  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);
  const [disabledToggleClientIDs, setDisabledToggleClientIDs] = useState<
    string[]
  >([]);

  const navigate = useNavigate();
  const { appID } = useParams<{ appID: string }>();

  const onManageScopes = useCallback(
    (item: ApplicationResourceListItem) => {
      navigate(
        `/project/${appID}/configuration/apps/${client.client_id}/edit/resource/${item.id}/scopes`
      );
    },
    [navigate, appID, client.client_id]
  );

  const handleSearchChange = useCallback(
    (_event?: React.ChangeEvent<HTMLInputElement>, newValue?: string): void => {
      setSearchKeyword(newValue ?? "");
      setOffset(0); // Reset offset on search change
    },
    [setSearchKeyword, setOffset]
  );

  const [addResource] = useAddResourceToClientIdMutation();
  const [removeResource] = useRemoveResourceFromClientIdMutation();

  const { data, loading, error, refetch } = useResourcesQueryQuery({
    variables: {
      first: PAGE_SIZE,
      after: encodeOffsetToCursor(offset),
      searchKeyword:
        debouncedSearchKeyword === "" ? undefined : debouncedSearchKeyword,
    },
    fetchPolicy: "cache-and-network",
  });

  const resourceListData: ApplicationResourceListItem[] = useMemo(() => {
    const resources =
      data?.resources?.edges
        ?.map((edge) => edge?.node)
        .filter((node) => !!node) ?? [];
    return resources.map((resource) => {
      const isAuthorized = resource.clientIDs.includes(client.client_id);
      return {
        id: resource.id,
        name: resource.name,
        resourceURI: resource.resourceURI,
        isAuthorized: isAuthorized,
      };
    });
  }, [client.client_id, data?.resources?.edges]);

  const handleCloseUnauthorizeDialog = useCallback(() => {
    setResourceToUnauthorize(null);
  }, []);

  const handleConfirmUnauthorize = useCallback(async () => {
    if (!resourceToUnauthorize) {
      return;
    }
    const clientId = client.client_id;
    const resourceUri = resourceToUnauthorize.resourceURI;

    const newResource = produce(resourceToUnauthorize, (draft) => {
      const newClientIds = draft.clientIDs;
      newClientIds.push(clientId);
      draft.clientIDs = newClientIds;
      return draft;
    });

    setDisabledToggleClientIDs((prev) => [...prev, resourceToUnauthorize.id]);
    setError(null);

    try {
      await removeResource({
        variables: {
          clientID: clientId,
          resourceURI: resourceUri,
        },
        refetchQueries: [ResourcesQueryDocument],
        awaitRefetchQueries: true,
        optimisticResponse: {
          removeResourceFromClientID: {
            resource: newResource,
          },
        },
        update: (cache) => {
          cache.modify<Resource>({
            id: cache.identify(resourceToUnauthorize),
            fields: {
              clientIDs: () => {
                return newResource.clientIDs;
              },
            },
          });
        },
      });
    } catch (e) {
      setError(parseRawError(e));
    } finally {
      setDisabledToggleClientIDs((prev) =>
        prev.filter((id) => id !== resourceToUnauthorize.id)
      );
      handleCloseUnauthorizeDialog();
    }
  }, [
    resourceToUnauthorize,
    client.client_id,
    removeResource,
    setError,
    handleCloseUnauthorizeDialog,
    setDisabledToggleClientIDs,
  ]);

  const handleToggleAuthorization = useCallback(
    async (item: ApplicationResourceListItem, isAuthorized: boolean) => {
      const clientId = client.client_id;
      const resourceUri = item.resourceURI;
      try {
        const resourceNode = data?.resources?.edges?.find(
          (edge) => edge?.node?.id === item.id
        )?.node;
        if (resourceNode == null) {
          throw new Error("unexpected: cannot find the origin resource node");
        }

        setDisabledToggleClientIDs((prev) => [...prev, item.id]);
        setError(null);
        if (!isAuthorized) {
          setResourceToUnauthorize(resourceNode);
        } else {
          const newResource = produce(resourceNode, (draft) => {
            const newClientIds = draft.clientIDs;
            newClientIds.push(clientId);
            draft.clientIDs = newClientIds;
            return draft;
          });
          await addResource({
            variables: {
              clientID: clientId,
              resourceURI: resourceUri,
            },
            refetchQueries: [ResourcesQueryDocument],
            awaitRefetchQueries: true,
            optimisticResponse: {
              addResourceToClientID: {
                resource: newResource,
              },
            },
            update: (cache) => {
              cache.modify<Resource>({
                id: cache.identify(resourceNode),
                fields: {
                  clientIDs: () => {
                    return newResource.clientIDs;
                  },
                },
              });
            },
          });
        }
      } catch (e) {
        setError(parseRawError(e));
      } finally {
        setDisabledToggleClientIDs((prev) =>
          prev.filter((id) => id !== item.id)
        );
      }
    },
    [
      client.client_id,
      data?.resources?.edges,
      setError,
      addResource,
      setDisabledToggleClientIDs,
    ]
  );

  const onChangeOffset = useCallback((newOffset: number) => {
    setOffset(newOffset);
  }, []);

  const pagination: PaginationProps = {
    offset,
    pageSize: PAGE_SIZE,
    totalCount: data?.resources?.totalCount ?? undefined,
    onChangeOffset,
  };
  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }
  return (
    <section className={cn(styles.resourcesSection, className)}>
      <WidgetTitle id="resoucres">
        <FormattedMessage id="EditOAuthClientForm.resources.title" />
      </WidgetTitle>
      <SearchBox
        placeholder={renderToString("search")}
        styles={{
          root: {
            width: 300,
          },
        }}
        onChange={handleSearchChange}
      />
      <ApplicationResourcesList
        resources={resourceListData}
        loading={loading}
        pagination={pagination}
        onToggleAuthorization={handleToggleAuthorization}
        disabledToggleClientIDs={disabledToggleClientIDs}
        onManageScopes={onManageScopes}
      />
      <UnauthorizeApplicationDialog
        data={
          resourceToUnauthorize
            ? { applicationName: client.client_name ?? client.name ?? "" }
            : null
        }
        onDismiss={handleCloseUnauthorizeDialog}
        onConfirm={handleConfirmUnauthorize}
        onDismissed={handleCloseUnauthorizeDialog}
      />
    </section>
  );
};
