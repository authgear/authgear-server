import React, { useCallback, useMemo, useState } from "react";
import { useBlocker, useNavigate } from "react-router-dom";
import { Location } from "history";
import BlockerDialog from "./BlockerDialog";

interface NavigationBlockerDialogProps {
  blockNavigation: boolean;
}

const NavigationBlockerDialog: React.FC<NavigationBlockerDialogProps> =
  function NavigationBlockerDialog(props: NavigationBlockerDialogProps) {
    const { blockNavigation } = props;

    const navigate = useNavigate();

    const [navigationBlockerDialog, setNavigationBlockerDialog] = useState<{
      visible: boolean;
      destination?: Location;
    }>({ visible: false });

    // disable block navigation when dialog visible
    const _blockNavigation = useMemo(() => {
      return !navigationBlockerDialog.visible && blockNavigation;
    }, [blockNavigation, navigationBlockerDialog.visible]);

    useBlocker((tx) => {
      setNavigationBlockerDialog({
        visible: true,
        destination: tx.location,
      });
    }, _blockNavigation);

    const onDialogDismiss = useCallback(() => {
      setNavigationBlockerDialog({ visible: false });
    }, []);

    const onDialogConfirm = useCallback(() => {
      const { destination } = navigationBlockerDialog;
      if (destination != null) {
        navigate(destination);
      }
      // We must dismiss the dialog because some navigation is merely hash change, e.g. Pivot.
      // If we do not dismiss the dialog, the dialog will block the content.
      setNavigationBlockerDialog({ visible: false });
    }, [navigate, navigationBlockerDialog]);

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
