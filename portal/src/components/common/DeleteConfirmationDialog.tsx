import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";
import { useSnapshotData } from "../../hook/useSnapshotData";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

interface DeleteConfirmationDialogProps<T> {
  data: T | null;
  renderTitle: (data: T) => React.ReactNode;
  renderSubText: (data: T) => React.ReactNode;
  onDismiss: () => void;
  onConfirm: (data: T) => void;
  isLoading: boolean;
  onDismissed?: () => void;
}

export function DeleteConfirmationDialog<T>(
  props: DeleteConfirmationDialogProps<T>
): React.ReactElement {
  const {
    onDismiss,
    onConfirm,
    isLoading,
    onDismissed,
    data,
    renderTitle,
    renderSubText,
  } = props;
  const isHidden = data === null || data === undefined;

  // Keep rendering the last data while the close transition plays.
  const snapshot = useSnapshotData(data);

  const onPressConfirm = useCallback(() => {
    if (isLoading || isHidden) {
      return;
    }
    onConfirm(data);
  }, [isLoading, isHidden, onConfirm, data]);

  const onDialogDismiss = useCallback(() => {
    if (isHidden) {
      return;
    }
    onDismiss();
  }, [isHidden, onDismiss]);

  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        if (!isHidden && !isLoading) {
          onDismiss();
        }
        onDismissed?.();
      }
    },
    [isHidden, isLoading, onDismiss, onDismissed]
  );

  return (
    <ConfirmationDialog
      open={!isHidden}
      onOpenChange={onOpenChange}
      title={snapshot != null ? renderTitle(snapshot) : ""}
      description={snapshot != null ? renderSubText(snapshot) : ""}
      confirmText={<FormattedMessage id="delete" />}
      cancelText={<FormattedMessage id="cancel" />}
      onConfirm={onPressConfirm}
      onCancel={onDialogDismiss}
      loading={isLoading}
      confirmColor="red"
    />
  );
}
