import React, { useCallback, useContext, useMemo } from "react";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSnapshotData } from "../../hook/useSnapshotData";
import { useSystemConfig } from "../../context/SystemConfigContext";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

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
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
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
      title: renderToString("DeleteScopeDialog.title"),
      subText: (
        <FormattedMessage
          id="DeleteScopeDialog.description"
          values={{
            scope: snapshot?.scope ?? "Unknown",
          }}
        />
      ) as unknown as string,
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
  };
