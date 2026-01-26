import React, { useCallback, useContext } from "react";
import { Context, FormattedMessage } from "../../intl";
import { DeleteConfirmationDialog } from "../common/DeleteConfirmationDialog";

export interface DeleteScopeDialogData {
  scope: string;
  description: string | null;
}

interface DeleteScopeDialogProps {
  data: DeleteScopeDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: DeleteScopeDialogData) => void;
  isLoading: boolean;
  onDismissed?: () => void;
}

export const DeleteScopeDialog: React.VFC<DeleteScopeDialogProps> =
  function DeleteScopeDialog(props) {
    const { onDismiss, onConfirm, isLoading, onDismissed, data } = props;
    const { renderToString } = useContext(Context);

    const renderTitle = useCallback(() => {
      return renderToString("DeleteScopeDialog.title");
    }, [renderToString]);

    const renderSubText = useCallback((data: DeleteScopeDialogData) => {
      return (
        <FormattedMessage
          id="DeleteScopeDialog.description"
          values={{
            scope: data.scope,
          }}
        />
      );
    }, []);

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
