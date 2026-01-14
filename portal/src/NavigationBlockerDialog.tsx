import React, { useCallback, useState } from "react";
import { Location } from "react-router";
import { useNavigate, useBlocker } from "react-router-dom";
import BlockerDialog from "./BlockerDialog";

interface NavigationBlockerDialogProps {
  blockNavigation: boolean;
  onConfirmNavigation?: () => void;
}

const NavigationBlockerDialog: React.VFC<NavigationBlockerDialogProps> =
  function NavigationBlockerDialog(props: NavigationBlockerDialogProps) {
    const { blockNavigation, onConfirmNavigation } = props;

    const navigate = useNavigate();

    const [navigationBlockerDialog, setNavigationBlockerDialog] = useState<{
      visible: boolean;
      destination?: Location;
    }>({ visible: false });

    useBlocker(
      useCallback(
        ({
          nextLocation,
        }: {
          currentLocation: Location;
          nextLocation: Location;
        }) => {
          if (blockNavigation && !navigationBlockerDialog.visible) {
            setNavigationBlockerDialog({
              visible: true,
              destination: nextLocation,
            });
            return true; // Block navigation
          }
          return false; // Do not block navigation
        },
        [blockNavigation, navigationBlockerDialog.visible]
      )
    );

    const onDialogDismiss = useCallback(() => {
      setNavigationBlockerDialog({ visible: false });
    }, []);

    const onDialogConfirm = useCallback(() => {
      const { destination } = navigationBlockerDialog;
      if (destination != null) {
        navigate(destination, { state: destination.state });
        onConfirmNavigation?.();
      }
      // We must dismiss the dialog because some navigation is merely hash change, e.g. Pivot.
      // If we do not dismiss the dialog, the dialog will block the content.
      setNavigationBlockerDialog({ visible: false });
    }, [navigate, navigationBlockerDialog, onConfirmNavigation]);

    return (
      <BlockerDialog
        hidden={!navigationBlockerDialog.visible}
        contentTitleId="NavigationBlockerDialog.title"
        contentSubTextId="NavigationBlockerDialog.content"
        contentConfirmId="NavigationBlockerDialog.confirm"
        onDialogConfirm={onDialogConfirm}
        onDialogDismiss={onDialogDismiss}
      />
    );
  };

export default NavigationBlockerDialog;
