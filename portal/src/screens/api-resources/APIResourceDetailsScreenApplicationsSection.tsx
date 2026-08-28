import React, { useMemo, useCallback, useState, useContext } from "react";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useAddResourceToClientIdMutation } from "../../graphql/adminapi/mutations/addResourceToClientID.generated";
import { useRemoveResourceFromClientIdMutation } from "../../graphql/adminapi/mutations/removeResourceFromClientID.generated";
import {
  ApplicationList,
  ApplicationListItem,
} from "../../components/api-resources/ApplicationList";
import { UnauthorizeApplicationDialog } from "../../components/api-resources/UnauthorizeApplicationDialog";
import { DynamicClientsAccessRow } from "../../components/api-resources/DynamicClientsAccessRow";
import { useParams, useNavigate } from "react-router-dom";
import {
  ResourceQueryDocument,
  ResourceQueryQuery,
} from "../../graphql/adminapi/query/resourceQuery.generated";
import { parseRawError } from "../../error/parse";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { PortalAPIAppConfig } from "../../types";
import Link from "../../Link";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import styles from "./APIResourceDetailsApplicationsSection.module.css";

export function APIResourceDetailsScreenApplicationsSection({
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
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchKeyword(e.target.value);
    },
    []
  );

  const onClearSearchKeyword = useCallback(() => {
    setSearchKeyword("");
  }, []);

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

  const handleConfirmUnauthorize = useCallback(() => {
    if (!applicationToUnauthorize) {
      return;
    }
    const unauthorize = async () => {
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
    };
    void unauthorize();
  }, [
    applicationToUnauthorize,
    resource,
    removeResource,
    setErrors,
    handleCloseUnauthorizeDialog,
  ]);

  const onToggleAuthorized = useCallback(
    (item: ApplicationListItem, checked: boolean) => {
      if (!checked) {
        handleOpenUnauthorizeDialog(item);
        return;
      }
      const authorize = async () => {
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
      };
      void authorize();
    },
    [resource, addResource, setErrors, handleOpenUnauthorizeDialog]
  );

  const isEmpty = applications.length === 0;

  return (
    <div className={styles.root}>
      <div className={styles.body}>
        <SettingsSectionCard
          title={
            <FormattedMessage id="APIResourceDetailsScreen.section.applications" />
          }
          description={
            <FormattedMessage id="APIResourceDetailsScreen.applications.description" />
          }
          contentClassName={styles.cardContent}
        >
          <DynamicClientsAccessRow resource={resource} />
          {isEmpty ? (
            <Text as="p" size="2" color="gray" className={styles.empty}>
              <FormattedMessage
                id="APIResourceDetailsScreen.applications.empty"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  ReactRouterLink: (chunks: React.ReactNode) => (
                    <Link to={`/project/${appID}/configuration/apps`}>
                      {chunks}
                    </Link>
                  ),
                }}
              />
            </Text>
          ) : (
            <>
              <div className={styles.searchField}>
                <TextField
                  size="2"
                  type="search"
                  onChange={onSearchQueryChange}
                  value={searchKeyword}
                  placeholder={renderToString("search")}
                  iconStart={TextFieldIcon.MagnifyingGlass}
                  suffixPlain={true}
                  suffix={
                    searchKeyword !== "" ? (
                      <button
                        type="button"
                        className={styles.searchClearButton}
                        aria-label={renderToString(
                          "APIResourcesScreen.clear-search"
                        )}
                        onClick={onClearSearchKeyword}
                      >
                        <Cross2Icon className={styles.searchClearIcon} />
                      </button>
                    ) : undefined
                  }
                />
              </div>
              <div className={styles.listContainer}>
                <ApplicationList
                  applications={filteredApplications}
                  className={styles.list}
                  loading={false}
                  onToggleAuthorized={onToggleAuthorized}
                  onManageScopes={onManageScopes}
                  disabledToggleClientIDs={disabledToggleClientIDs}
                />
              </div>
            </>
          )}
        </SettingsSectionCard>
      </div>
      <UnauthorizeApplicationDialog
        data={
          applicationToUnauthorize
            ? { applicationName: applicationToUnauthorize.name }
            : null
        }
        onDismiss={handleCloseUnauthorizeDialog}
        onConfirm={handleConfirmUnauthorize}
      />
    </div>
  );
}
