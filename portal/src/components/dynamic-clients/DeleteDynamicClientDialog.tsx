import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";
import { OAuthClientSource } from "../../graphql/adminapi/globalTypes.generated";

export interface DeleteDynamicClientDialogData {
  clientID: string;
  clientName: string;
  source: OAuthClientSource;
}

export interface DeleteDynamicClientDialogProps {
  data: DeleteDynamicClientDialogData | null;
  isLoading: boolean;
  onConfirm: (data: DeleteDynamicClientDialogData) => void;
  onDismiss: () => void;
}

export const DeleteDynamicClientDialog: React.VFC<DeleteDynamicClientDialogProps> =
  function DeleteDynamicClientDialog({
    data,
    isLoading,
    onConfirm,
    onDismiss,
  }) {
    const onOpenChange = useCallback(
      (open: boolean) => {
        if (!open && !isLoading) {
          onDismiss();
        }
      },
      [isLoading, onDismiss]
    );

    const onCancel = useCallback(() => {
      if (!isLoading) {
        onDismiss();
      }
    }, [isLoading, onDismiss]);

    const onConfirmClicked = useCallback(() => {
      if (data != null) {
        onConfirm(data);
      }
    }, [data, onConfirm]);

    return (
      <ConfirmationDialog
        open={data != null}
        onOpenChange={onOpenChange}
        title={<FormattedMessage id="DeleteDynamicClientDialog.title" />}
        description={
          // Deleting a CIMD client only evicts the record Authgear fetched:
          // the same client_id resolves again on its next authorization
          // request (docs/specs/cimd.md § Client Limit), so the copy must
          // not describe it as permanent the way the DCR copy does.
          data?.source === OAuthClientSource.Cimd ? (
            <FormattedMessage
              id="DeleteDynamicClientDialog.description.cimd"
              values={{ clientName: data.clientName }}
            />
          ) : (
            <FormattedMessage
              id="DeleteDynamicClientDialog.description.dcr"
              values={{ clientName: data?.clientName ?? "" }}
            />
          )
        }
        confirmText={
          <FormattedMessage id="DeleteDynamicClientDialog.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmClicked}
        onCancel={onCancel}
        loading={isLoading}
        confirmColor="red"
      />
    );
  };
