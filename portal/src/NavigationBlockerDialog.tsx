import React, { useCallback, useEffect, useState } from "react";
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

    const blocker = useBlocker(
      useCallback(
        ({
          currentLocation,
          nextLocation,
        }: {
          currentLocation: Location;
          nextLocation: Location;
        }) => {
          // A navigation that stays on the same path (e.g. a hash-only
          // change from Pivot, a search-param update, or an internal
          // replace navigation like useLocationEffect popping location
          // state) does not navigate the user away from this page, so it
          // must never trigger the confirmation dialog.
          const isSamePath = currentLocation.pathname === nextLocation.pathname;
          if (isSamePath) {
            return false;
          }

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

    useEffect(() => {
      // ensure the blocker is reset at unmount
      return () => {
        if (blocker.state === "blocked") blocker.reset();
      };
    }, [blocker]);

    const onDialogDismiss = useCallback(() => {
      // Release the router's blocked transition. Otherwise the navigation
      // stays blocked even after this dialog is hidden, and the very next
      // navigation attempt can find the router still stuck mid-transition.
      if (blocker.state === "blocked") {
        blocker.reset();
      }
      setNavigationBlockerDialog({ visible: false });
    }, [blocker]);

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
