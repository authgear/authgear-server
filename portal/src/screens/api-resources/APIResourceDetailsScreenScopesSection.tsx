import React, { useState, useMemo, useCallback, useContext } from "react";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Text } from "@radix-ui/themes";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { FormContainerBase } from "../../FormContainerBase";
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
import { EditScopeDialog } from "../../components/api-resources/EditScopeDialog";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { usePaginatedSearchParams } from "../../hook/usePaginatedSearchParams";
import { useDebounced } from "../../hook/useDebounced";
import styles from "./APIResourceDetailsScopesSection.module.css";

export function APIResourceDetailsScreenScopesSection({
  resource,
}: {
  resource: Resource;
}): JSX.Element {
  const [createScope] = useCreateScopeMutationMutation();
  const [deleteScope] = useDeleteScopeMutationMutation();
  const [initialState] = useState<CreateScopeFormState>({
    scope: "",
    description: "",
    allowDynamicThirdPartyClientAccess: false,
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
            accessPolicy: {
              allowDynamicThirdPartyClientAccess:
                sanitized.allowDynamicThirdPartyClientAccess,
            },
          },
        },
        refetchQueries: [ResourceScopesQueryDocument],
        awaitRefetchQueries: true,
      });
    },
  });

  const { offset, setOffset, searchKeyword, setSearchKeyword } =
    usePaginatedSearchParams();
  const [deleteDialogData, setDeleteDialogData] =
    useState<DeleteScopeDialogData | null>(null);
  const [editingScope, setEditingScope] = useState<Scope | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const pageSize = 10;

  const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);

  const { renderToString } = useContext(MessageContext);

  const onSearchKeywordChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setOffset(0);
      setSearchKeyword(e.target.value);
    },
    [setOffset, setSearchKeyword]
  );

  const onClearSearchKeyword = useCallback(() => {
    setOffset(0);
    setSearchKeyword("");
  }, [setOffset, setSearchKeyword]);

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

  const onEdit = useCallback((scope: Scope) => {
    setEditingScope(scope);
  }, []);

  const onDismissEditDialog = useCallback(() => {
    setEditingScope(null);
  }, []);

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={error} onRetry={refetch} />;
  }

  const hasListContent = scopes.length > 0 || searchKeyword !== "";

  return (
    <FormContainerBase form={form}>
      <div className={styles.root}>
        <SettingsSectionCard
          title={
            <FormattedMessage id="APIResourceDetailsScreen.scopes.list.title" />
          }
          description={
            <FormattedMessage id="APIResourceDetailsScreen.scopes.description" />
          }
          contentClassName={styles.cardContent}
        >
          <div className={styles.addSection}>
            <Text as="p" size="3" weight="medium" className={styles.addHeading}>
              <FormattedMessage id="APIResourceDetailsScreen.scopes.add.title" />
            </Text>
            <CreateScopeForm
              className={styles.createForm}
              state={form.state}
              setState={form.setState}
            />
          </div>
          <hr className={styles.divider} />
          <div className={styles.listSection}>
            {hasListContent ? (
              <>
                <div className={styles.searchField}>
                  <TextField
                    size="2"
                    type="search"
                    placeholder={renderToString("search")}
                    value={searchKeyword}
                    iconStart={TextFieldIcon.MagnifyingGlass}
                    onChange={onSearchKeywordChange}
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
                {scopes.length > 0 ? (
                  <ScopeList
                    className={styles.list}
                    scopes={scopes}
                    loading={loading}
                    pagination={pagination}
                    onEdit={onEdit}
                    onDelete={onDelete}
                  />
                ) : (
                  <Text as="p" size="2" color="gray" className={styles.empty}>
                    <FormattedMessage id="APIResourceDetailsScreen.scopes.list.empty" />
                  </Text>
                )}
              </>
            ) : (
              <Text as="p" size="2" color="gray" className={styles.empty}>
                <FormattedMessage id="APIResourceDetailsScreen.scopes.list.empty" />
              </Text>
            )}
          </div>
        </SettingsSectionCard>
      </div>
      <EditScopeDialog
        resourceURI={resource.resourceURI}
        scope={editingScope}
        onDismiss={onDismissEditDialog}
        onSaved={() => {
          refetch().catch(() => {});
        }}
      />
      <DeleteScopeDialog
        data={deleteDialogData}
        isLoading={isDeleting}
        // eslint-disable-next-line @typescript-eslint/strict-void-return
        onConfirm={onConfirmDelete}
        onDismiss={onDismissDeleteDialog}
      />
    </FormContainerBase>
  );
}
