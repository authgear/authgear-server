import React, { useCallback, useContext, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Text } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import DefaultButton from "../../DefaultButton";
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
import { DynamicClientList, DynamicClientListItem } from "./DynamicClientList";
import { DynamicClientsEmptyView } from "./DynamicClientsEmptyView";
import { DynamicClientDetailsDialog } from "./DynamicClientDetailsDialog";
import {
  DeleteDynamicClientDialog,
  DeleteDynamicClientDialogData,
} from "./DeleteDynamicClientDialog";
import { TextWithCopyButton } from "../common/TextWithCopyButton";
import styles from "./DynamicClientsTab.module.css";

const PAGE_SIZE = 10;

export interface DynamicClientsTabProps {
  registrationEnabled: boolean;
  publicOrigin: string;
  // The smallest block-action quota configured for oauth_client_dcr, or null
  // when the plan does not limit dynamic clients.
  dcrClientQuota: number | null;
}

function DynamicClientsTabContent({
  registrationEnabled,
  publicOrigin,
  dcrClientQuota,
}: DynamicClientsTabProps): React.ReactElement {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
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
    skip: !registrationEnabled,
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

  const onSettingsClick = useCallback(() => {
    navigate("./dcr");
  }, [navigate]);

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
          setDeleteDialogData(null);
          setDetailsClient(null);
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
        });
    },
    [deleteDynamicClient, clients.length, offset, refetch, setErrors]
  );

  if (!registrationEnabled) {
    return <DynamicClientsEmptyView registrationEnabled={false} />;
  }

  if (error != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={error} onRetry={refetch} />;
  }

  const isEmpty = !loading && clients.length === 0 && offset === 0;

  return (
    <div className={styles.root}>
      <ErrorMessageBar />
      <div className={styles.header}>
        <div className={styles.headerText}>
          <Text
            variant="medium"
            block={true}
            styles={{ root: { color: "var(--gray-11)" } }}
          >
            <FormattedMessage id="DynamicClientsTab.registration-endpoint.label" />
          </Text>
          <TextWithCopyButton text={`${publicOrigin}/oauth2/register`} />
          {dcrClientQuota != null && totalCount != null ? (
            <Text
              variant="medium"
              block={true}
              styles={{ root: { color: "var(--gray-11)" } }}
            >
              <FormattedMessage
                id="DynamicClientsTab.quota"
                values={{ count: totalCount, quota: dcrClientQuota }}
              />
            </Text>
          ) : null}
        </div>
        <DefaultButton
          text={renderToString("DynamicClientsTab.settings-button")}
          onClick={onSettingsClick}
        />
      </div>
      {isEmpty ? (
        <DynamicClientsEmptyView registrationEnabled={true} />
      ) : (
        <DynamicClientList
          clients={clients}
          loading={loading}
          pagination={pagination}
          onDelete={onRequestDelete}
          onItemClicked={onItemClicked}
        />
      )}
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
    </div>
  );
}

export const DynamicClientsTab: React.VFC<DynamicClientsTabProps> =
  function DynamicClientsTab(props) {
    return (
      <ErrorMessageBarContextProvider>
        <DynamicClientsTabContent {...props} />
      </ErrorMessageBarContextProvider>
    );
  };
