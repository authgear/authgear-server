import React, { useMemo, useCallback, useState, useContext } from "react";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import WidgetTitle from "../../WidgetTitle";
import { SearchBox, Text } from "@fluentui/react";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useAddResourceToClientIdMutation } from "../../graphql/adminapi/mutations/addResourceToClientID.generated";
import { useRemoveResourceFromClientIdMutation } from "../../graphql/adminapi/mutations/removeResourceFromClientID.generated";
import {
  ApplicationList,
  ApplicationListItem,
} from "../../components/api-resources/ApplicationList";
import { UnauthorizeApplicationDialog } from "../../components/api-resources/UnauthorizeApplicationDialog";
import { useParams, useNavigate } from "react-router-dom";
import {
  ResourceQueryDocument,
  ResourceQueryQuery,
} from "../../graphql/adminapi/query/resourceQuery.generated";
import { parseRawError } from "../../error/parse";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { PortalAPIAppConfig } from "../../types";

export function APIResourceDetailsScreenApplicationsTab({
  resource,
  effectiveAppConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
}): JSX.Element {
  const { appID } = useParams() as { appID: string };
  const [addResource] = useAddResourceToClientIdMutation();
  const [removeResource] = useRemoveResourceFromClientIdMutation();
  const { setErrors } = useErrorMessageBarContext();
  const { renderToString } = useContext(MessageContext);
  const { themes } = useSystemConfig();
  const [disabledToggleClientIDs, setDisabledToggleClientIDs] = useState<
    string[]
  >([]);

  const [applicationToUnauthorize, setApplicationToUnauthorize] =
    useState<ApplicationListItem | null>(null);

  const [searchKeyword, setSearchKeyword] = useState("");

  const applications = useMemo((): ApplicationListItem[] => {
    return (
      effectiveAppConfig.oauth?.clients
        ?.filter((clientConfig) => {
          switch (clientConfig.x_application_type) {
            case "m2m":
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
  }, [effectiveAppConfig.oauth?.clients, resource.clientIDs]);

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

  const navigate = useNavigate();

  const onManageScopes = useCallback(
    (item: ApplicationListItem) => {
      navigate(
        `/project/${appID}/api-resources/${resource.id}/applications/${item.clientID}/scopes`
      );
    },
    [appID, navigate, resource.id]
  );

  const handleOpenUnauthorizeDialog = useCallback(
    (item: ApplicationListItem) => {
      setApplicationToUnauthorize(item);
    },
    []
  );

  const handleCloseUnauthorizeDialog = useCallback(() => {
    setApplicationToUnauthorize(null);
  }, []);

  const handleConfirmUnauthorize = useCallback(async () => {
    if (!applicationToUnauthorize) {
      return;
    }
    try {
      setDisabledToggleClientIDs((prev) => [
        ...prev,
        applicationToUnauthorize.clientID,
      ]);

      const newResource = {
        ...resource,
        clientIDs: resource.clientIDs.filter(
          (clientID) => clientID !== applicationToUnauthorize.clientID
        ),
      };

      await removeResource({
        variables: {
          clientID: applicationToUnauthorize.clientID,
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
          cache.writeQuery<ResourceQueryQuery>({
            query: ResourceQueryDocument,
            variables: { id: resource.id },
            data: { node: newResource },
          });
        },
      });
    } catch (e: unknown) {
      setErrors(parseRawError(e));
    } finally {
      setDisabledToggleClientIDs((prev) =>
        prev.filter(
          (clientID) => clientID !== applicationToUnauthorize.clientID
        )
      );
      handleCloseUnauthorizeDialog();
    }
  }, [
    applicationToUnauthorize,
    resource,
    removeResource,
    setErrors,
    handleCloseUnauthorizeDialog,
  ]);

  const onToggleAuthorized = useCallback(
    async (item: ApplicationListItem, checked: boolean) => {
      if (!checked) {
        handleOpenUnauthorizeDialog(item);
        return;
      }
      try {
        setDisabledToggleClientIDs((prev) => [...prev, item.clientID]);
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
            cache.writeQuery<ResourceQueryQuery>({
              query: ResourceQueryDocument,
              variables: { id: resource.id },
              data: { node: newResource },
            });
          },
        });
      } catch (e: unknown) {
        setErrors(parseRawError(e));
      } finally {
        setDisabledToggleClientIDs((prev) =>
          prev.filter((clientID) => clientID !== item.clientID)
        );
      }
    },
    [resource, addResource, setErrors, handleOpenUnauthorizeDialog]
  );

  const isEmpty = applications.length === 0;

  return (
    <div className="pt-5 flex-1 flex flex-col space-y-4">
      <header className="space-y-2">
        <WidgetTitle>
          <FormattedMessage id="APIResourceDetailsScreen.tab.applications" />
        </WidgetTitle>
        <Text block={true}>
          <FormattedMessage id="APIResourceDetailsScreen.applications.description" />
        </Text>
        {isEmpty ? (
          <Text
            styles={{ root: { color: themes.main.palette.neutralTertiary } }}
          >
            <FormattedMessage
              id="APIResourceDetailsScreen.applications.empty"
              values={{
                to: `/project/${appID}/configuration/apps`,
              }}
            />
          </Text>
        ) : null}
      </header>

      {isEmpty ? null : (
        <>
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
              loading={false} // The app config query should always be completed
              onToggleAuthorized={onToggleAuthorized}
              onManageScopes={onManageScopes}
              disabledToggleClientIDs={disabledToggleClientIDs}
            />
          </div>
        </>
      )}
      <UnauthorizeApplicationDialog
        data={
          applicationToUnauthorize
            ? { applicationName: applicationToUnauthorize.name }
            : null
        }
        onDismiss={handleCloseUnauthorizeDialog}
        onConfirm={handleConfirmUnauthorize}
        onDismissed={handleCloseUnauthorizeDialog}
      />
    </div>
  );
}
