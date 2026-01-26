import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
} from "@fluentui/react";
import React, { useCallback, useMemo } from "react";
import { FormattedMessage } from "../../../../intl";
import { useSystemConfig } from "../../../../context/SystemConfigContext";
import PrimaryButton from "../../../../PrimaryButton";
import DefaultButton from "../../../../DefaultButton";
import ErrorDialog from "../../../../error/ErrorDialog";

export interface RolesAndGroupsBaseDeleteDialogProps<T> {
  title: string;
  subText?: string;
  buttonText: string;
  isHidden: boolean;
  loading: boolean;
  data: T | null;
  onDismiss: (isDeleted: boolean) => void;
  onDismissed?: () => void;
  onConfirm: () => void;
  error: unknown;
}

const dialogStyles = { main: { minHeight: 0 } };
function RolesAndGroupsBaseDeleteDialog<T>(
  props: RolesAndGroupsBaseDeleteDialogProps<T>
): React.ReactElement {
  const {
    isHidden,
    loading,
    onDismiss,
    onDismissed,
    onConfirm,
    title,
    subText,
    error,
    buttonText,
  } = props;
  const { themes } = useSystemConfig();
  const dialogContentProps: IDialogContentProps = {
    title,
    subText,
  };

  const onDialogDismiss = useCallback(() => {
    if (loading || isHidden) {
      return;
    }
    onDismiss(false);
  }, [loading, isHidden, onDismiss]);

  const modalProps = useMemo((): IModalProps => {
    return {
      onDismissed,
    };
  }, [onDismissed]);

  return (
    <>
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
            disabled={loading}
            onClick={onConfirm}
            text={buttonText}
          />
          <DefaultButton
            onClick={onDialogDismiss}
            disabled={loading}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
      <ErrorDialog error={error} />
    </>
  );
}

export default RolesAndGroupsBaseDeleteDialog;
