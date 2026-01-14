import React, { useState, useCallback, useMemo, useContext } from "react";
import { encodeOffsetToCursor } from "../../util/pagination";
import ScreenDescription from "../../ScreenDescription";
import {
  FormattedMessage,
  Context as MessageContext,
} from "../../intl";
import { ResourceList } from "../../components/api-resources/ResourceList";
import { useResourcesQueryQuery } from "../../graphql/adminapi/query/resourcesQuery.generated";
import { useDeleteResourceMutation } from "../../graphql/adminapi/mutations/deleteResourceMutation.generated";
import ShowError from "../../ShowError";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { PaginationProps } from "../../PaginationWidget";
import { CreateResourceButton } from "../../components/api-resources/CreateResourceButton";
import {
  DeleteResourceDialog,
  DeleteResourceDialogData,
} from "../../components/api-resources/DeleteResourceDialog";
import { SearchBox } from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { useDebounced } from "../../hook/useDebounced";

const PAGE_SIZE = 10;

const APIResourcesScreen: React.VFC = function APIResourcesScreen() {
  const [offset, setOffset] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState("");
  const [deleteDialogData, setDeleteDialogData] =
    useState<DeleteResourceDialogData | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);

  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();
  const { appID } = useParams<{ appID: string }>();

  const onSearchQueryChange = useCallback(
    (_, newValue) => {
      setOffset(0); // Reset offset to 0 on search query change
      setSearchKeyword(newValue ?? "");
    },
    [setOffset, setSearchKeyword]
  );
  const { data, loading, error, refetch } = useResourcesQueryQuery({
    variables: {
      first: PAGE_SIZE,
      after: encodeOffsetToCursor(offset),
      searchKeyword:
        debouncedSearchKeyword === "" ? undefined : debouncedSearchKeyword,
    },
    fetchPolicy: "network-only",
  });

  const [deleteResource] = useDeleteResourceMutation();

  const onConfirmDelete = useCallback(
    async (data: DeleteResourceDialogData) => {
      setIsDeleting(true);
      try {
        await deleteResource({
          variables: {
            input: {
              resourceURI: data.resourceURI,
            },
          },
        });
        setDeleteDialogData(null);
        await refetch();
      } finally {
        setIsDeleting(false);
      }
    },
    [deleteResource, refetch, setIsDeleting, setDeleteDialogData]
  );

  const onDismissDeleteDialog = useCallback(() => {
    setDeleteDialogData(null);
  }, [setDeleteDialogData]);

  const onDelete = useCallback(
    (resource: Resource) => {
      setDeleteDialogData({
        resourceURI: resource.resourceURI,
        resourceName: resource.name ?? null,
      });
    },
    [setDeleteDialogData]
  );

  const onEdit = useCallback(
    (resource) => {
      navigate(
        `/project/${appID}/api-resources/${encodeURIComponent(resource.id)}`
      );
    },
    [navigate, appID]
  );

  const resources = useMemo(() => {
    return (
      data?.resources?.edges
        ?.map((edge) => edge?.node)
        .filter(
          (resource): resource is NonNullable<typeof resource> => !!resource
        ) ?? []
    );
  }, [data]);

  const onChangeOffset = useCallback((offset: number) => {
    setOffset(offset);
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
    <>
      <APIResourceScreenLayout
        breadcrumbItems={[
          {
            to: "",
            label: <FormattedMessage id="APIResourcesScreen.title" />,
          },
        ]}
        headerDescription={
          <ScreenDescription>
            <FormattedMessage id="APIResourcesScreen.description" />
          </ScreenDescription>
        }
        headerSuffix={
          resources.length !== 0 ? (
            <CreateResourceButton className="self-start" />
          ) : null
        }
      >
        <div className="col-span-full flex flex-col">
          <SearchBox
            className="mb-4"
            styles={{ root: { width: 300 } }}
            onChange={onSearchQueryChange}
            value={searchKeyword}
            placeholder={renderToString("search")}
          />
          <ResourceList
            className="flex-1"
            resources={resources}
            loading={loading}
            pagination={pagination}
            onDelete={onDelete}
            onItemClicked={onEdit}
          />
        </div>
      </APIResourceScreenLayout>
      <DeleteResourceDialog
        data={deleteDialogData}
        isLoading={isDeleting}
        onConfirm={onConfirmDelete}
        onDismiss={onDismissDeleteDialog}
      />
    </>
  );
};

export default APIResourcesScreen;
