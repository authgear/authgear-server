import React from "react";
import { Button, Dialog } from "@radix-ui/themes";
import { SecondaryButton } from "../Button/SecondaryButton/SecondaryButton";
import styles from "./ConfirmationDialog.module.css";

export interface ConfirmationDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: React.ReactNode;
  description: React.ReactNode;
  confirmText: React.ReactNode;
  cancelText: React.ReactNode;
  onConfirm: () => void;
  // Called only when the Cancel button is clicked. Escape and overlay
  // dismissal go through onOpenChange(false) instead, so any cancel
  // side-effect must live in (or be shared with) onOpenChange.
  onCancel: () => void;
  loading?: boolean;
  confirmColor?: "red" | "indigo";
  maxWidth?: string;
}

export function ConfirmationDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmText,
  cancelText,
  onConfirm,
  onCancel,
  loading = false,
  confirmColor = "red",
  maxWidth = "400px",
}: ConfirmationDialogProps): React.ReactElement {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth={maxWidth} size="3">
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.Description size="2">{description}</Dialog.Description>
        <div className={styles.actions}>
          <SecondaryButton
            size="2"
            disabled={loading}
            text={cancelText}
            onClick={onCancel}
          />
          <Button
            size="2"
            variant="solid"
            color={confirmColor}
            loading={loading}
            disabled={loading}
            onClick={onConfirm}
          >
            {confirmText}
          </Button>
        </div>
      </Dialog.Content>
    </Dialog.Root>
  );
}
