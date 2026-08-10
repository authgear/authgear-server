import React, {
  useState,
  useCallback,
  useMemo,
  useContext,
  useEffect,
} from "react";
import { Text } from "@radix-ui/themes";
import { Cross2Icon } from "@radix-ui/react-icons";
import { encodeOffsetToCursor } from "../../util/pagination";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import { ResourceList } from "../../components/api-resources/ResourceList";
import { useResourcesQueryQuery } from "../../graphql/adminapi/query/resourcesQuery.generated";
import { useDeleteResourceMutation } from "../../graphql/adminapi/mutations/deleteResourceMutation.generated";
import ShowError from "../../ShowError";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { PaginationProps } from "../../PaginationWidget";
import { CreateResourceButton } from "../../components/api-resources/CreateResourceButton";
import { CreateAPIResourceDialog } from "../../components/api-resources/CreateAPIResourceDialog";
import {
  DeleteResourceDialog,
  DeleteResourceDialogData,
} from "../../components/api-resources/DeleteResourceDialog";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { useDebounced } from "../../hook/useDebounced";
import ExternalLink from "../../ExternalLink";
import styles from "./APIResourcesScreen.module.css";

const PAGE_SIZE = 10;

const APIResourcesScreen: React.VFC = function APIResourcesScreen() {
  const [offset, setOffset] = useState(0);
  const [searchKeyword, setSearchKeyword] = useState("");
  const [deleteDialogData, setDeleteDialogData] =
    useState<DeleteResourceDialogData | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);
  const [searchParams, setSearchParams] = useSearchParams();

  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();
  const { appID } = useParams<{ appID: string }>();

  const wantCreateDialog = searchParams.get("create") === "1";
  const [prevWantCreateDialog, setPrevWantCreateDialog] = useState(false);
  if (prevWantCreateDialog !== wantCreateDialog) {
    setPrevWantCreateDialog(wantCreateDialog);
    if (wantCreateDialog) {
      setCreateDialogOpen(true);
    }
  }

  useEffect(() => {
    if (searchParams.get("create") === "1") {
      const next = new URLSearchParams(searchParams);
      next.delete("create");
      setSearchParams(next, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const onSearchQueryChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setOffset(0);
      setSearchKeyword(e.target.value);
    },
    []
  );

  const onClearSearchKeyword = useCallback(() => {
    setOffset(0);
    setSearchKeyword("");
  }, []);

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
    [deleteResource, refetch]
  );

  const onDismissDeleteDialog = useCallback(() => {
    setDeleteDialogData(null);
  }, []);

  const onDelete = useCallback((resource: Resource) => {
    setDeleteDialogData({
      resourceURI: resource.resourceURI,
      resourceName: resource.name ?? null,
    });
  }, []);

  const onEdit = useCallback(
    (resource: ResourceListItem) => {
      navigate(
        `/project/${appID}/api-resources/${encodeURIComponent(resource.id)}`
      );
    },
    [navigate, appID]
  );

  const onOpenCreateDialog = useCallback(() => {
    setCreateDialogOpen(true);
  }, []);

  const onDismissCreateDialog = useCallback(() => {
    setCreateDialogOpen(false);
  }, []);

  const resources = useMemo(() => {
    return (
      data?.resources?.edges
        ?.map((edge) => edge?.node)
        .filter(
          (resource): resource is NonNullable<typeof resource> => !!resource
        ) ?? []
    );
  }, [data]);

  const isEmpty =
    !loading &&
    searchKeyword === "" &&
    (data?.resources?.totalCount ?? 0) === 0;

  const onChangeOffset = useCallback((nextOffset: number) => {
    setOffset(nextOffset);
  }, []);

  const pagination: PaginationProps = {
    offset,
    pageSize: PAGE_SIZE,
    totalCount: data?.resources?.totalCount ?? undefined,
    onChangeOffset,
  };

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
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
          <Text as="p" size="2" className={styles.headerDescription}>
            <FormattedMessage
              id="APIResourcesScreen.description"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                DocLink: (chunks: React.ReactNode) => (
                  <ExternalLink href="https://docs.authgear.com/get-started/m2m-applications">
                    {chunks}
                  </ExternalLink>
                ),
              }}
            />
          </Text>
        }
        headerSuffix={
          !isEmpty ? (
            <CreateResourceButton
              className="self-start"
              onClick={onOpenCreateDialog}
            />
          ) : null
        }
      >
        <div className={styles.content}>
          {!isEmpty ? (
            <div className={styles.searchField}>
              <TextField
                size="2"
                type="search"
                value={searchKeyword}
                placeholder={renderToString("search")}
                iconStart={TextFieldIcon.MagnifyingGlass}
                onChange={onSearchQueryChange}
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
          ) : null}
          <ResourceList
            className={styles.list}
            resources={resources}
            loading={loading}
            pagination={pagination}
            onDelete={onDelete}
            onItemClicked={onEdit}
            onCreateClick={onOpenCreateDialog}
          />
        </div>
      </APIResourceScreenLayout>
      <CreateAPIResourceDialog
        open={createDialogOpen}
        onDismiss={onDismissCreateDialog}
        onCreated={() => {
          void refetch();
        }}
      />
      <DeleteResourceDialog
        data={deleteDialogData}
        isLoading={isDeleting}
        // eslint-disable-next-line @typescript-eslint/strict-void-return
        onConfirm={onConfirmDelete}
        onDismiss={onDismissDeleteDialog}
      />
    </>
  );
};

interface ResourceListItem {
  id: string;
  name?: string | null;
  resourceURI: string;
}

export default APIResourcesScreen;
