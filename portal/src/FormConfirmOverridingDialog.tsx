import React, { useCallback, useState } from "react";
import { useFormConflictErrors } from "./form";
import { FormattedMessage } from "./intl";
import { ConfirmationDialog } from "./components/v2/ConfirmationDialog/ConfirmationDialog";

interface FormConfirmOverridingDialogProps {
  save: (ignoreConflict: boolean) => void;
}

const FormConfirmOverridingDialog: React.VFC<FormConfirmOverridingDialogProps> =
  function FormConfirmOverridingDialog(props) {
    const { save } = props;

    const errors = useFormConflictErrors();
    const isConflicted = errors.length !== 0;
    const saveWithoutChecksum = useCallback(() => {
      save(true);
    }, [save]);

    // Open the dialog when the form becomes conflicted and close it when
    // the conflict resolves (e.g. the forced save succeeded).
    const [visible, setVisible] = useState(isConflicted);
    const [prevIsConflicted, setPrevIsConflicted] = useState(isConflicted);
    if (prevIsConflicted !== isConflicted) {
      setPrevIsConflicted(isConflicted);
      setVisible(isConflicted);
    }

    const onCancel = useCallback(() => {
      setVisible(false);
    }, []);

    const onOpenChange = useCallback((open: boolean) => {
      if (!open) {
        setVisible(false);
      }
    }, []);

    return (
      <ConfirmationDialog
        open={visible}
        onOpenChange={onOpenChange}
        title={<FormattedMessage id="FormConfirmOverridingDialog.title" />}
        description={
          <FormattedMessage id="FormConfirmOverridingDialog.subtext" />
        }
        confirmText={
          <FormattedMessage id="FormConfirmOverridingDialog.button.confirm" />
        }
        cancelText={
          <FormattedMessage id="FormConfirmOverridingDialog.button.cancel" />
        }
        confirmColor="red"
        onConfirm={saveWithoutChecksum}
        onCancel={onCancel}
      />
    );
  };

export default FormConfirmOverridingDialog;
