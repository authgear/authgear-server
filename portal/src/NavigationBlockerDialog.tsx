import React, { useCallback, useContext, useMemo, useState } from "react";
import { useBlocker, useNavigate } from "react-router-dom";
import { Location } from "history";
import {
  IDialogContentProps,
  Dialog,
  DialogType,
  DialogFooter,
  PrimaryButton,
  DefaultButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

interface NavigationBlockerDialogProps {
  blockNavigation: boolean;
}

const NavigationBlockerDialog: React.FC<NavigationBlockerDialogProps> = function NavigationBlockerDialog(
  props: NavigationBlockerDialogProps
) {
  const { blockNavigation } = props;

  const { renderToString } = useContext(Context);
  const navigate = useNavigate();

  // disable block navigation when dialog visible
  const [disableBlockNavigation, setDisableBlockNavigation] = useState(false);
  const [navigationBlockerDialog, setNavigationBlockerDialog] = useState<{
    visible: boolean;
    destination?: Location;
  }>({ visible: false });

  const _blockNavigation = useMemo(() => {
    return !disableBlockNavigation && blockNavigation;
  }, [blockNavigation, disableBlockNavigation]);

  useBlocker((tx) => {
    setNavigationBlockerDialog({
      visible: true,
      destination: tx.location,
    });
    setDisableBlockNavigation(true);
  }, _blockNavigation);

  const dialogContentProps: IDialogContentProps = useMemo(
    () => ({
      type: DialogType.normal,
      title: <FormattedMessage id="NavigationBlockerDialog.title" />,
      subText: renderToString("NavigationBlockerDialog.content"),
    }),
    [renderToString]
  );

  const onDialogDismiss = useCallback(() => {
    setDisableBlockNavigation(false);
    setNavigationBlockerDialog({ visible: false });
  }, []);

  const onDialogConfirm = useCallback(() => {
    const { destination } = navigationBlockerDialog;
    if (destination != null) {
      navigate(destination);
    } else {
      onDialogDismiss();
    }
  }, [navigate, navigationBlockerDialog, onDialogDismiss]);

  return (
    <Dialog
      hidden={!navigationBlockerDialog.visible}
      onDismiss={onDialogDismiss}
      dialogContentProps={dialogContentProps}
    >
      <DialogFooter>
        <PrimaryButton onClick={onDialogConfirm}>
          <FormattedMessage id="confirm" />
        </PrimaryButton>
        <DefaultButton onClick={onDialogDismiss}>
          <FormattedMessage id="cancel" />
        </DefaultButton>
      </DialogFooter>
    </Dialog>
  );
};

export default NavigationBlockerDialog;
