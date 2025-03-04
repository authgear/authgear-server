import React from "react";
import { Toast } from "radix-ui";
import styles from "./Toast.module.css";

export interface ToastProviderProps {
  children?: React.ReactChild;
}

export function ToastProvider({
  children,
}: ToastProviderProps): React.ReactElement {
  return (
    <Toast.Provider swipeDirection="right">
      {children}
      <Toast.Viewport className={styles.ToastViewport} />
    </Toast.Provider>
  );
}

export function ToastRoot({
  children,
  open,
  onOpenChange,
}: {
  children?: React.ReactChild | null;
  open: boolean;
  onOpenChange: (value: boolean) => void;
}): React.ReactElement {
  return (
    <Toast.Root
      className={styles.ToastRoot}
      open={open}
      onOpenChange={onOpenChange}
    >
      {children}
    </Toast.Root>
  );
}
