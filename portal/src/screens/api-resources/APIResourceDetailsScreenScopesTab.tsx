import React, { useState, useMemo, useCallback, useContext } from "react";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
import WidgetTitle from "../../WidgetTitle";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import { Resource, Scope } from "../../graphql/adminapi/globalTypes.generated";
import {
  CreateScopeForm,
  CreateScopeFormState,
  sanitizeCreateScopeFormState,
} from "../../components/api-resources/CreateScopeForm";
import { useCreateScopeMutationMutation } from "../../graphql/adminapi/mutations/createScopeMutation.generated";
import { useDeleteScopeMutationMutation } from "../../graphql/adminapi/mutations/deleteScopeMutation.generated";
import {
  ResourceScopesQueryDocument,
  useResourceScopesQueryQuery,
} from "../../graphql/adminapi/query/resourceScopesQuery.generated";
import { ScopeList } from "../../components/api-resources/ScopeList";
import { encodeOffsetToCursor } from "../../util/pagination";
import ShowError from "../../ShowError";
import {
  DeleteScopeDialog,
  DeleteScopeDialogData,
} from "../../components/api-resources/DeleteScopeDialog";
import { SearchBox, Text } from "@fluentui/react";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { usePaginatedSearchParams } from "../../hook/usePaginatedSearchParams";
import { useDebounced } from "../../hook/useDebounced";

export function APIResourceDetailsScreenScopesTab({
  resource,
}: {
  resource: Resource;
}): JSX.Element {
  const [createScope] = useCreateScopeMutationMutation();
  const [deleteScope] = useDeleteScopeMutationMutation();
  const [initialState] = useState<CreateScopeFormState>({
    scope: "",
    description: "",
  });
  const form = useSimpleForm<CreateScopeFormState, any>({
    defaultState: initialState,
    submit: async (state) => {
      const sanitized = sanitizeCreateScopeFormState(state);
      await createScope({
        variables: {
          input: {
            resourceURI: resource.resourceURI,
            scope: sanitized.scope,
            description: sanitized.description,
          },
        },
        refetchQueries: [ResourceScopesQueryDocument],
        awaitRefetchQueries: true,
      });
    },
    stateMode:
      "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
  });

  const { offset, setOffset, searchKeyword, setSearchKeyword } =
    usePaginatedSearchParams();
  const [deleteDialogData, setDeleteDialogData] =
    useState<DeleteScopeDialogData | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const pageSize = 10;

  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);

  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();
  const location = useLocation();
  const { appID } = useParams<{ appID: string }>();

  const onSearchKeywordChange = useMemo(
    () => (_: any, newValue?: string) => {
      setOffset(0);
      setSearchKeyword(newValue ?? "");
    },
    [setOffset, setSearchKeyword]
  );

  const { data, loading, error, refetch } = useResourceScopesQueryQuery({
    variables: {
      resourceID: resource.id,
      first: pageSize,
      after: encodeOffsetToCursor(offset),
      searchKeyword:
        debouncedSearchKeyword === "" ? undefined : debouncedSearchKeyword,
    },
    fetchPolicy: "cache-and-network",
  });

  const scopes = useMemo(() => {
    return data?.node?.__typename === "Resource"
      ? data.node.scopes?.edges
          ?.map((edge) => edge?.node)
          .filter((n): n is Scope => !!n) ?? []
      : [];
  }, [data]);

  const totalCount = useMemo(() => {
    return data?.node?.__typename === "Resource"
      ? data.node.scopes?.totalCount ?? 0
      : 0;
  }, [data]);

  const pagination = useMemo(() => {
    return {
      offset,
      pageSize,
      totalCount,
      onChangeOffset: setOffset,
    };
  }, [offset, pageSize, totalCount, setOffset]);

  const onDelete = useCallback((scope: Scope) => {
    setDeleteDialogData({
      scope: scope.scope,
      description: scope.description ?? null,
    });
  }, []);

  const onConfirmDelete = useCallback(
    async (data: DeleteScopeDialogData) => {
      setIsDeleting(true);
      try {
        await deleteScope({
          variables: {
            input: {
              resourceURI: resource.resourceURI,
              scope: data.scope,
            },
          },
        });
        setDeleteDialogData(null);
        await refetch();
      } finally {
        setIsDeleting(false);
      }
    },
    [deleteScope, refetch, resource.resourceURI]
  );

  const onDismissDeleteDialog = useCallback(() => {
    setDeleteDialogData(null);
  }, []);

  const onEdit = useCallback(
    (scope: Scope) => {
      navigate({
        pathname: `/project/${encodeURIComponent(
          appID ?? ""
        )}/api-resources/${encodeURIComponent(
          resource.id
        )}/scopes/${encodeURIComponent(scope.id)}`,
        hash: location.hash,
        search: location.search,
      });
    },
    [navigate, appID, resource.id, location]
  );

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContainerBase form={form}>
      <div className="pt-5 flex-1 flex flex-col space-y-2">
        <header>
          <WidgetTitle className="mb-2">
            <FormattedMessage id="APIResourceDetailsScreen.tab.scopes" />
          </WidgetTitle>
          <Text>
            <FormattedMessage id="APIResourceDetailsScreen.scopes.description" />
          </Text>
        </header>
        <div className="flex-1 flex flex-col space-y-4">
          <div className="flex items-start justify-between gap-x-4 tablet:flex-col tablet:items-stretch">
            <CreateScopeForm
              className="flex-1-0-auto min-w-40"
              state={form.state}
              setState={form.setState}
            />
            <SearchBox
              styles={{
                root: {
                  marginTop: 30,
                  width: 260,
                  "@media (max-width: 1080px)": { marginTop: 0 },
                },
              }}
              placeholder={renderToString("search")}
              value={searchKeyword}
              onChange={onSearchKeywordChange}
            />
          </div>
          {scopes.length > 0 ? (
            <ScopeList
              className="flex-1 min-h-0"
              scopes={scopes}
              loading={loading}
              pagination={pagination}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ) : null}
        </div>
      </div>
      <DeleteScopeDialog
        data={deleteDialogData}
        isLoading={isDeleting}
        onConfirm={onConfirmDelete}
        onDismiss={onDismissDeleteDialog}
      />
    </FormContainerBase>
  );
}
