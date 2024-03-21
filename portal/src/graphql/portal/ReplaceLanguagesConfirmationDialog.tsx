import React, { useMemo, useContext, useCallback } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Dialog, DialogFooter, IDialogContentProps } from "@fluentui/react";

import DefaultButton from "../../DefaultButton";
import PrimaryButton from "../../PrimaryButton";

export interface ReplaceLanguagesConfirmationDialogProps {
  visible: boolean;
  onConfirm: () => void;
  onDismiss: () => void;
}

const ReplaceLanguagesConfirmationDialog: React.VFC<ReplaceLanguagesConfirmationDialogProps> =
  function ReplaceLanguagesConfirmationDialog(props) {
    const { visible, onConfirm: saveLanguages, onDismiss } = props;

    const { renderToString } = useContext(Context);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      return {
        title: (
          <FormattedMessage id="ReplaceLanguagesConfirmationDialog.title" />
        ),
        subText: renderToString("ReplaceLanguagesConfirmationDialog.message"),
      };
    }, [renderToString]);

    const onConfirmClicked = useCallback(() => {
      saveLanguages();
    }, [saveLanguages]);

    return (
      <Dialog
        hidden={!visible}
        dialogContentProps={dialogContentProps}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={onConfirmClicked}
            text={<FormattedMessage id="confirm" />}
            disabled={!visible}
          />
          <DefaultButton
            disabled={!visible}
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };

export default ReplaceLanguagesConfirmationDialog;
