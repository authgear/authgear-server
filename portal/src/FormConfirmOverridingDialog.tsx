import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  Dialog,
  DialogFooter,
  DialogType,
  IDialogContentProps,
} from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import { useFormConflictErrors } from "./form";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";

interface FormConfirmOverridingDialogProps {
  save: (ignoreConflict: boolean) => void;
}

const FormConfirmOverridingDialog: React.VFC<FormConfirmOverridingDialogProps> =
  function FormConfirmOverridingDialog(props) {
    const { save } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const errors = useFormConflictErrors();
    const isConfliced = errors.length !== 0;
    const saveWithoutChecksum = useCallback(() => {
      save(true);
    }, [save]);

    const [visible, setVisible] = useState(false);

    const onCancel = useCallback(() => {
      setVisible(false);
    }, []);

    useEffect(() => {
      setVisible(isConfliced);
    }, [isConfliced]);

    const dialogContentProps: IDialogContentProps = useMemo(
      () => ({
        type: DialogType.normal,
        title: <FormattedMessage id="FormConfirmOverridingDialog.title" />,
        subText: renderToString("FormConfirmOverridingDialog.subtext"),
      }),
      [renderToString]
    );
    return (
      <Dialog hidden={!visible} dialogContentProps={dialogContentProps}>
        <DialogFooter>
          <DefaultButton
            onClick={onCancel}
            text={
              <FormattedMessage id="FormConfirmOverridingDialog.button.cancel" />
            }
          />
          <PrimaryButton
            onClick={saveWithoutChecksum}
            theme={themes.destructive}
            text={
              <FormattedMessage id="FormConfirmOverridingDialog.button.confirm" />
            }
          />
        </DialogFooter>
      </Dialog>
    );
  };

export default FormConfirmOverridingDialog;
