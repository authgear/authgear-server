import React, { useCallback, useContext } from "react";
import { Context, FormattedMessage } from "../../intl";
import { DeleteConfirmationDialog } from "../common/DeleteConfirmationDialog";
import { OAuthClientSecretKey } from "../../types";

export interface DeleteClientSecretConfirmationDialogData {
  clientSecret: OAuthClientSecretKey;
}

interface DeleteClientSecretConfirmationDialogProps {
  data: DeleteClientSecretConfirmationDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: DeleteClientSecretConfirmationDialogData) => void;
  isLoading: boolean;
  onDismissed?: () => void;
}

export const DeleteClientSecretConfirmationDialog: React.VFC<DeleteClientSecretConfirmationDialogProps> =
  function DeleteClientSecretConfirmationDialog(props) {
    const { onDismiss, onConfirm, isLoading, onDismissed, data } = props;
    const { renderToString } = useContext(Context);

    const renderTitle = useCallback(() => {
      return renderToString("DeleteClientSecretConfirmationDialog.title");
    }, [renderToString]);

    const renderSubText = useCallback(
      (_data: DeleteClientSecretConfirmationDialogData) => {
        return (
          <FormattedMessage id="DeleteClientSecretConfirmationDialog.description" />
        );
      },
      []
    );

    return (
      <DeleteConfirmationDialog
        data={data}
        renderTitle={renderTitle}
        renderSubText={renderSubText}
        onDismiss={onDismiss}
        onConfirm={onConfirm}
        isLoading={isLoading}
        onDismissed={onDismissed}
      />
    );
  };
