import React, { useCallback, useEffect } from "react";
import { useBlocker } from "react-router-dom";
import BlockerDialog from "./BlockerDialog";

interface NavigationBlockerDialogProps {
  // Always-fresh dirty check (see useSyncFormStates), called at the
  // exact moment a navigation is attempted. A plain boolean prop would
  // only be as fresh as the last render -- a form's save() can clear
  // dirtiness and then immediately navigate away (e.g. via
  // FormContainerBase's afterSave) before React has re-rendered this
  // component with the updated value, causing this dialog to show
  // spuriously right after a successful save.
  getIsDirty: () => boolean;
  onConfirmNavigation?: () => void;
}

const NavigationBlockerDialog: React.VFC<NavigationBlockerDialogProps> =
  function NavigationBlockerDialog(props: NavigationBlockerDialogProps) {
    const { getIsDirty, onConfirmNavigation } = props;

    const shouldBlock = useCallback(
      ({
        currentLocation,
        nextLocation,
      }: {
        currentLocation: { pathname: string };
        nextLocation: { pathname: string };
      }) => {
        // A navigation that stays on the same path (e.g. a hash-only
        // change from Pivot, a search-param update, or an internal
        // replace navigation like useLocationEffect popping location
        // state) does not navigate the user away from this page, so it
        // must never trigger the confirmation dialog.
        if (currentLocation.pathname === nextLocation.pathname) {
          return false;
        }
        return getIsDirty();
      },
      [getIsDirty]
    );

    const blocker = useBlocker(shouldBlock);

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
    }, [blocker]);

    const onDialogConfirm = useCallback(() => {
      onConfirmNavigation?.();
      if (blocker.state === "blocked") {
        blocker.proceed();
      }
    }, [blocker, onConfirmNavigation]);

    return (
      <BlockerDialog
        open={blocker.state === "blocked"}
        contentTitleId="NavigationBlockerDialog.title"
        contentSubTextId="NavigationBlockerDialog.content"
        contentConfirmId="NavigationBlockerDialog.confirm"
        onDialogConfirm={onDialogConfirm}
        onDialogDismiss={onDialogDismiss}
      />
    );
  };

export default NavigationBlockerDialog;
