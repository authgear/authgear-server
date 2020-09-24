import React, { useCallback, useMemo, useState } from "react";
import { useBlocker, useNavigate } from "react-router-dom";
import { Location } from "history";
import BlockerDialog from "./BlockerDialog";

interface NavigationBlockerDialogProps {
  blockNavigation: boolean;
}

const NavigationBlockerDialog: React.FC<NavigationBlockerDialogProps> = function NavigationBlockerDialog(
  props: NavigationBlockerDialogProps
) {
  const { blockNavigation } = props;

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
    <BlockerDialog
      hidden={!navigationBlockerDialog.visible}
      contentTitleId="NavigationBlockerDialog.title"
      contentSubTextId="NavigationBlockerDialog.content"
      onDialogConfirm={onDialogConfirm}
      onDialogDismiss={onDialogDismiss}
    />
  );
};

export default NavigationBlockerDialog;
