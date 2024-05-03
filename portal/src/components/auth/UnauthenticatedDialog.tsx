import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  IModalProps,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import React, { useMemo } from "react";
import PrimaryButton from "../../PrimaryButton";

interface UnauthenticatedDialogProps {
  isHidden: boolean;
  onConfirm: () => void;
}

export const UnauthenticatedDialog: React.VFC<UnauthenticatedDialogProps> =
  function UnauthenticatedDialog({ isHidden, onConfirm }) {
    const modalProps = useMemo((): IModalProps => {
      return {
        isBlocking: true,
      };
    }, []);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      return {
        showCloseButton: false,
        title: <FormattedMessage id="UnauthenticatedDialog.title" />,
        subText: (
          <FormattedMessage id="UnauthenticatedDialog.description" />
        ) as unknown as string,
      };
    }, []);

    return (
      <Dialog
        hidden={isHidden}
        modalProps={modalProps}
        dialogContentProps={dialogContentProps}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={onConfirm}
            text={<FormattedMessage id="UnauthenticatedDialog.button" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };
