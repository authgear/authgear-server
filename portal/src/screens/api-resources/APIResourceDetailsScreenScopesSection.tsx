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
import { accessPolicyStateFromAccessPolicy } from "../../components/api-resources/accessPolicy";
import {
  BulkAccessChanges,
  useBulkScopeAccess,
} from "../../components/api-resources/useBulkScopeAccess";
import { ScopeBulkAccessBar } from "../../components/api-resources/ScopeBulkAccessBar";
import { ScopeBulkAccessDialog } from "../../components/api-resources/ScopeBulkAccessDialog";
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
  // A new scope starts with the resource's own access policy, so the common
  // case needs no change; the resource must allow a category anyway for the
  // scope to be requestable.
  const [initialState] = useState<CreateScopeFormState>({
    scope: "",
    description: "",
    accessPolicy: accessPolicyStateFromAccessPolicy(resource.accessPolicy),
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
            accessPolicy: sanitized.accessPolicy,
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

  const bulk = useBulkScopeAccess({
    resourceURI: resource.resourceURI,
    scopes,
    refetch,
  });
  const selectedIDs = useMemo(
    () => new Set(bulk.selected.keys()),
    [bulk.selected]
  );

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

  const onToggleAllVisible = useCallback(
    (checked: boolean) => {
      bulk.toggleAll(scopes, checked);
    },
    [bulk, scopes]
  );
  const [isBulkDialogOpen, setIsBulkDialogOpen] = useState(false);
  const selectedScopes = useMemo(
    () => Array.from(bulk.selected.values()),
    [bulk.selected]
  );
  const onOpenBulkDialog = useCallback(() => {
    setIsBulkDialogOpen(true);
  }, []);
  const onDismissBulkDialog = useCallback(() => {
    if (!bulk.isApplying) {
      setIsBulkDialogOpen(false);
    }
  }, [bulk.isApplying]);
  const onApplyBulk = useCallback(
    (changes: BulkAccessChanges) => {
      bulk.apply(changes).finally(() => {
        setIsBulkDialogOpen(false);
      });
    },
    [bulk]
  );

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
              resourceAccessPolicy={resource.accessPolicy}
              state={form.state}
              setState={form.setState}
            />
          </div>
          <div className={styles.listSection}>
            {hasListContent ? (
              <>
                <div className={styles.toolbar}>
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
                  <ScopeBulkAccessBar
                    count={bulk.selected.size}
                    disabled={bulk.isApplying}
                    onChangeAccess={onOpenBulkDialog}
                    onClear={bulk.clear}
                  />
                </div>
                {scopes.length > 0 ? (
                  <ScopeList
                    className={styles.list}
                    scopes={scopes}
                    resourceAccessPolicy={resource.accessPolicy}
                    selectedIDs={selectedIDs}
                    onToggleSelected={bulk.toggle}
                    onToggleAllVisible={onToggleAllVisible}
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
        resourceAccessPolicy={resource.accessPolicy}
        scope={editingScope}
        onDismiss={onDismissEditDialog}
        onSaved={() => {
          refetch().catch(() => {});
        }}
      />
      <ScopeBulkAccessDialog
        open={isBulkDialogOpen}
        scopes={selectedScopes}
        isApplying={bulk.isApplying}
        onApply={onApplyBulk}
        onDismiss={onDismissBulkDialog}
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
