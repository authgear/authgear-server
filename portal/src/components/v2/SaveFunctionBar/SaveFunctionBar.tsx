import React, { useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { FormattedMessage } from "../../../intl";
import { useFormContainerBaseContext } from "../../../FormContainerBase";
import { PrimaryButton } from "../Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../Button/SecondaryButton/SecondaryButton";
import { ConfirmationDialog } from "../ConfirmationDialog/ConfirmationDialog";
import styles from "./SaveFunctionBar.module.css";
import { useSaveFunctionBarAlignment } from "./useSaveFunctionBarAlignment";

// Keep in sync with the transition duration in SaveFunctionBar.module.css.
const TRANSITION_MS = 200;

export interface SaveFunctionBarProps {
  className?: string;
  anchorRef?: React.RefObject<HTMLElement | null>;
}

export function SaveFunctionBar({
  className,
  anchorRef,
}: SaveFunctionBarProps): React.ReactElement | null {
  const { canReset, canSave, isDirty, isUpdating, onReset, onSave } =
    useFormContainerBaseContext();
  const alignStyle = useSaveFunctionBarAlignment(anchorRef);

  const [isDiscardDialogOpen, setIsDiscardDialogOpen] = useState(false);
  const onOpenDiscardDialog = useCallback(() => {
    setIsDiscardDialogOpen(true);
  }, []);
  const onDismissDiscardDialog = useCallback(() => {
    setIsDiscardDialogOpen(false);
  }, []);
  const onDiscardDialogOpenChange = useCallback((open: boolean) => {
    if (!open) {
      setIsDiscardDialogOpen(false);
    }
  }, []);
  const onConfirmDiscard = useCallback(() => {
    onReset();
    setTimeout(() => setIsDiscardDialogOpen(false), 0);
  }, [onReset]);

  // Keep the bar mounted across the exit animation: `rendered` controls
  // mounting; the in/out keyframes play via the rootIn/rootOut classes.
  const [rendered, setRendered] = useState(isDirty);
  useEffect(() => {
    if (isDirty) {
      setRendered(true);
      return;
    }
    const timer = setTimeout(() => setRendered(false), TRANSITION_MS);
    return () => clearTimeout(timer);
  }, [isDirty]);

  if (!rendered) {
    return null;
  }

  const fixedStyle =
    anchorRef != null && alignStyle != null ? alignStyle : undefined;

  return (
    <>
      <div
        className={cn(
          styles.root,
          isDirty ? styles.rootIn : styles.rootOut,
          className
        )}
        style={fixedStyle}
        role="region"
        aria-live="polite"
      >
        <div className={styles.message}>
          <InfoCircledIcon className={styles.messageIcon} aria-hidden={true} />
          <Text as="p" size="2" color="gray">
            <FormattedMessage id="SaveFunctionBar.message" />
          </Text>
        </div>
        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            disabled={!canReset}
            text={<FormattedMessage id="SaveFunctionBar.discard" />}
            onClick={onOpenDiscardDialog}
          />
          <PrimaryButton
            size="2"
            disabled={!canSave}
            loading={isUpdating}
            text={<FormattedMessage id="save" />}
            onClick={onSave}
          />
        </div>
      </div>
      <ConfirmationDialog
        open={isDiscardDialogOpen}
        onOpenChange={onDiscardDialogOpenChange}
        title={<FormattedMessage id="FormContainer.reset-dialog.title" />}
        description={
          <FormattedMessage id="FormContainer.reset-dialog.message" />
        }
        confirmText={
          <FormattedMessage id="FormContainer.reset-dialog.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        confirmColor="red"
        onConfirm={onConfirmDiscard}
        onCancel={onDismissDiscardDialog}
      />
    </>
  );
}
