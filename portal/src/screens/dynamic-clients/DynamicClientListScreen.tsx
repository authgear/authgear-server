import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "../../intl";
import cn from "classnames";
import { Heading } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import Link from "../../Link";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
  useErrorMessageBarContext,
} from "../../ErrorMessageBar";
import { parseRawError } from "../../error/parse";
import { encodeOffsetToCursor } from "../../util/pagination";
import { PaginationProps } from "../../PaginationWidget";
import { useDynamicClientsQueryQuery } from "../../graphql/adminapi/query/dynamicClientsQuery.generated";
import { useDeleteDynamicClientMutationMutation } from "../../graphql/adminapi/mutations/deleteDynamicClientMutation.generated";
import {
  DynamicClientList,
  DynamicClientListItem,
} from "../../components/dynamic-clients/DynamicClientList";
import { DynamicClientsEmptyView } from "../../components/dynamic-clients/DynamicClientsEmptyView";
import { DynamicClientDetailsDialog } from "../../components/dynamic-clients/DynamicClientDetailsDialog";
import {
  DeleteDynamicClientDialog,
  DeleteDynamicClientDialogData,
} from "../../components/dynamic-clients/DeleteDynamicClientDialog";
import styles from "./DynamicClientListScreen.module.css";

const PAGE_SIZE = 10;

function DynamicClientListScreenContent(): React.ReactElement {
  const { appID } = useParams() as { appID: string };
  const { setErrors } = useErrorMessageBarContext();

  const [offset, setOffset] = useState(0);
  const [detailsClient, setDetailsClient] =
    useState<DynamicClientListItem | null>(null);
  const [deleteDialogData, setDeleteDialogData] =
    useState<DeleteDynamicClientDialogData | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const { data, loading, error, refetch } = useDynamicClientsQueryQuery({
    variables: {
      first: PAGE_SIZE,
      after: encodeOffsetToCursor(offset),
    },
    fetchPolicy: "network-only",
  });

  const [deleteDynamicClient] = useDeleteDynamicClientMutationMutation();

  const clients = useMemo((): DynamicClientListItem[] => {
    return (
      data?.dynamicClients?.edges
        ?.map((edge) => edge?.node)
        .filter((node): node is NonNullable<typeof node> => !!node)
        .map((node) => ({
          id: node.id,
          clientID: node.clientID,
          clientName: node.clientName ?? null,
          name: node.name,
          kind: node.kind,
          source: node.source,
          registeredAt: node.registeredAt ?? null,
          applicationType: node.applicationType ?? null,
          redirectURIs: [...node.redirectURIs],
          grantTypes: [...node.grantTypes],
          responseTypes: [...node.responseTypes],
          logoURI: node.logoURI ?? null,
          clientURI: node.clientURI ?? null,
          tosURI: node.tosURI ?? null,
          policyURI: node.policyURI ?? null,
        })) ?? []
    );
  }, [data]);

  const totalCount = data?.dynamicClients?.totalCount ?? undefined;

  const onChangeOffset = useCallback((offset: number) => {
    setOffset(offset);
  }, []);

  const pagination: PaginationProps = {
    offset,
    pageSize: PAGE_SIZE,
    totalCount,
    onChangeOffset,
  };

  const onItemClicked = useCallback((item: DynamicClientListItem) => {
    setDetailsClient(item);
  }, []);

  const onDismissDetails = useCallback(() => {
    setDetailsClient(null);
  }, []);

  const onRequestDelete = useCallback((client: DynamicClientListItem) => {
    setDeleteDialogData({
      clientID: client.clientID,
      clientName: client.name,
    });
  }, []);

  const onDismissDeleteDialog = useCallback(() => {
    setDeleteDialogData(null);
  }, []);

  const onConfirmDelete = useCallback(
    (data: DeleteDynamicClientDialogData) => {
      setIsDeleting(true);
      deleteDynamicClient({
        variables: { input: { clientID: data.clientID } },
      })
        .then(async () => {
          // If the last item of a later page was deleted, step back a page so
          // the list does not show an empty page.
          if (clients.length === 1 && offset > 0) {
            setOffset(Math.max(0, offset - PAGE_SIZE));
          }
          return refetch();
        })
        .catch((e: unknown) => {
          setErrors(parseRawError(e));
        })
        .finally(() => {
          setIsDeleting(false);
          // Close the dialogs on failure too — the error message bar renders
          // behind the modal overlay and would otherwise be invisible.
          setDeleteDialogData(null);
          setDetailsClient(null);
        });
    },
    [deleteDynamicClient, clients.length, offset, refetch, setErrors]
  );

  const isEmpty = !loading && clients.length === 0 && offset === 0;

  return (
    <ScreenContent>
      <div className={cn(styles.widget, styles.pageHeader)}>
        <Link
          to={`/project/${appID}/configuration/apps#dynamic-clients`}
          className={styles.backLink}
        >
          <ChevronLeftIcon className={styles.backLinkIcon} />
          <span>
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          </span>
        </Link>
        <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="DynamicClientListScreen.title" />
        </Heading>
      </div>
      <div className={cn(styles.widget, styles.content)}>
        <ErrorMessageBar />
        {error != null ? (
          // eslint-disable-next-line @typescript-eslint/strict-void-return
          <ShowError error={error} onRetry={refetch} />
        ) : isEmpty ? (
          <DynamicClientsEmptyView />
        ) : (
          <DynamicClientList
            clients={clients}
            loading={loading}
            pagination={pagination}
            onDelete={onRequestDelete}
            onItemClicked={onItemClicked}
          />
        )}
      </div>
      <DynamicClientDetailsDialog
        client={detailsClient}
        onDelete={onRequestDelete}
        onDismiss={onDismissDetails}
      />
      <DeleteDynamicClientDialog
        data={deleteDialogData}
        isLoading={isDeleting}
        onConfirm={onConfirmDelete}
        onDismiss={onDismissDeleteDialog}
      />
    </ScreenContent>
  );
}

const DynamicClientListScreen: React.VFC = function DynamicClientListScreen() {
  return (
    <ErrorMessageBarContextProvider>
      <DynamicClientListScreenContent />
    </ErrorMessageBarContextProvider>
  );
};

export default DynamicClientListScreen;
