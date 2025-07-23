import React, { useState, useCallback, useMemo, useContext } from "react";
import ScreenContent from "../../ScreenContent";
import { encodeOffsetToCursor } from "../../util/pagination";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import ScreenContentHeader from "../../ScreenContentHeader";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
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

const PAGE_SIZE = 20;

const APIResourcesScreen: React.VFC = function APIResourcesScreen() {
  const [offset, setOffset] = useState(0);
  const [searchQuery, setSearchQuery] = useState("");
  const [deleteDialogData, setDeleteDialogData] =
    useState<DeleteResourceDialogData | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { renderToString } = useContext(MessageContext);

  const onSearchQueryChange = useCallback(
    (_, newValue) => {
      setOffset(0); // Reset offset to 0 on search query change
      setSearchQuery(newValue ?? "");
    },
    [setOffset, setSearchQuery]
  );
  const { data, loading, error, refetch } = useResourcesQueryQuery({
    variables: {
      first: PAGE_SIZE,
      after: useMemo(() => {
        if (offset === 0) {
          return undefined;
        }
        return encodeOffsetToCursor(offset - 1);
      }, [offset]),
      searchKeyword: searchQuery === "" ? undefined : searchQuery,
    },
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
      <ScreenContent className="flex-1" layout="list">
        <ScreenContentHeader
          title={
            <ScreenTitle>
              <FormattedMessage id="APIResourcesScreen.title" />
            </ScreenTitle>
          }
          description={
            <ScreenDescription>
              <FormattedMessage id="APIResourcesScreen.description" />
            </ScreenDescription>
          }
          suffix={
            resources.length !== 0 ? (
              <CreateResourceButton
                onClick={() => {
                  // TODO
                }}
                className="self-start"
              />
            ) : null
          }
        />
        <div className="col-span-full flex flex-col">
          <SearchBox
            className="mb-4"
            styles={{ root: { width: 300 } }}
            onChange={onSearchQueryChange}
            value={searchQuery}
            placeholder={renderToString("search")}
          />
          <ResourceList
            className="flex-1"
            resources={resources}
            loading={loading}
            pagination={pagination}
            onDelete={onDelete}
            onEdit={() => {
              // TODO
            }}
          />
        </div>
      </ScreenContent>
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
