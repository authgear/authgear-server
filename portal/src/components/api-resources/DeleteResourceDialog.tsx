import React, { useCallback } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { DeleteConfirmationDialog } from "../common/DeleteConfirmationDialog";

export interface DeleteResourceDialogData {
  resourceURI: string;
  resourceName: string | null;
}

interface DeleteResourceDialogProps {
  data: DeleteResourceDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: DeleteResourceDialogData) => void;
  isLoading: boolean;
  onDismissed?: () => void;
}

export const DeleteResourceDialog: React.VFC<DeleteResourceDialogProps> =
  function DeleteResourceDialog(props) {
    const { onDismiss, onConfirm, isLoading, onDismissed, data } = props;
    const { renderToString } = React.useContext(Context);

    const renderTitle = useCallback(() => {
      return renderToString("DeleteResourceDialog.title");
    }, [renderToString]);

    const renderSubText = useCallback((data: DeleteResourceDialogData) => {
      return (
        <FormattedMessage
          id="DeleteResourceDialog.description"
          values={{
            name: data.resourceName ?? data.resourceURI,
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
