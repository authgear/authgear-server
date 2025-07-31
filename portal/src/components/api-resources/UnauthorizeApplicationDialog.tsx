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

export interface UnauthorizeApplicationDialogData {
  applicationName: string | null;
}

interface UnauthorizeApplicationDialogProps {
  data: UnauthorizeApplicationDialogData | null;
  onDismiss: () => void;
  onConfirm: (data: UnauthorizeApplicationDialogData) => void;
  isLoading: boolean;
  onDismissed?: () => void;
}

export const UnauthorizeApplicationDialog: React.VFC<UnauthorizeApplicationDialogProps> =
  function UnauthorizeApplicationDialog(props) {
    const { onDismiss, onConfirm, isLoading, onDismissed, data } = props;
    const isHidden = data === null;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    // Keep the latest non-null data, because the dialog has transition animation before dismiss.
    // During the transition, we still need the data. However, the parent may already changed the props.
    const snapshot = useSnapshotData(data);

    const onPressConfirm = useCallback(() => {
      if (isLoading || isHidden) {
        return;
      }
      onConfirm(data);
    }, [isLoading, isHidden, onConfirm, data]);

    const dialogStyles = { main: { minHeight: 0 } };
    const dialogContentProps: IDialogContentProps = {
      title: renderToString("UnauthorizeApplicationDialog.title"),
      subText: (
        <FormattedMessage
          id="UnauthorizeApplicationDialog.description"
          values={{
            applicationName: snapshot?.applicationName ?? "Unknown Application",
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
              disabled={isLoading}
              onClick={onPressConfirm}
              text={
                <FormattedMessage id="UnauthorizeApplicationDialog.unauthorize" />
              }
            />
            <DefaultButton
              onClick={onDialogDismiss}
              disabled={isLoading}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
      </>
    );
  };
