import React, { useCallback, useMemo } from "react";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
} from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import { useSnapshotData } from "../../hook/useSnapshotData";
import { useSystemConfig } from "../../context/SystemConfigContext";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

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
  const { themes } = useSystemConfig();

  const snapshot = useSnapshotData(data);

  const onPressConfirm = useCallback(() => {
    if (isLoading || isHidden) {
      return;
    }
    onConfirm(data);
  }, [isLoading, isHidden, onConfirm, data]);

  const dialogStyles = { main: { minHeight: 0 } };
  const dialogContentProps: IDialogContentProps = {
    title: snapshot != null ? (renderTitle(snapshot) as unknown as string) : "",
    subText:
      snapshot != null ? (renderSubText(snapshot) as unknown as string) : "",
  };

  const onDialogDismiss = useCallback(() => {
    if (isHidden) {
      return;
    }
    onDismiss();
  }, [isHidden, onDismiss]);

  const modalProps = useMemo((): IModalProps => {
    return {
      onDismissed,
    };
  }, [onDismissed]);

  return (
    <Dialog
      hidden={isHidden}
      onDismiss={onDialogDismiss}
      modalProps={modalProps}
      dialogContentProps={dialogContentProps}
      styles={dialogStyles}
    >
      <DialogFooter>
        <PrimaryButton
          theme={themes.destructive}
          disabled={isLoading}
          onClick={onPressConfirm}
          text={<FormattedMessage id="delete" />}
        />
        <DefaultButton
          onClick={onDialogDismiss}
          disabled={isLoading}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
}
