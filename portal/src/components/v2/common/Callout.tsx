import {
  Cross2Icon,
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from "@radix-ui/react-icons";
import { Callout as RadixCallout } from "@radix-ui/themes";
import React, { ComponentProps, useCallback } from "react";
import styles from "./Callout.module.css";
import { useToastContext, useToastProviderContext } from "./Toast";

export enum CalloutType {
  error = "error",
  success = "success",
}

export interface CalloutProps {
  type: CalloutType;
  text?: React.ReactChild;
  showCloseButton?: boolean;
}

function typeToRadixColor(
  type: CalloutType
): ComponentProps<typeof RadixCallout.Root>["color"] {
  switch (type) {
    case CalloutType.error:
      return "red";
    case CalloutType.success:
      return "green";
  }
}

function CalloutIcon({ color }: { color: CalloutType }) {
  switch (color) {
    case CalloutType.error:
      return <ExclamationTriangleIcon />;
    case CalloutType.success:
      return <CheckCircledIcon />;
  }
}

export function Callout({
  type,
  text,
  showCloseButton = true,
}: CalloutProps): React.ReactElement {
  const { setOpen } = useToastContext();

  const onClose = useCallback(() => {
    setOpen(false);
  }, [setOpen]);

  return (
    <RadixCallout.Root
      className={styles.calloutRoot}
      color={typeToRadixColor(type)}
      size="2"
      variant="surface"
    >
      <RadixCallout.Icon className={styles.calloutIcon}>
        <CalloutIcon color={type} />
      </RadixCallout.Icon>
      <RadixCallout.Text className={styles.calloutText}>
        {text}
      </RadixCallout.Text>
      {showCloseButton ? (
        <button
          type="button"
          onClick={onClose}
          className={styles.calloutAction}
        >
          <Cross2Icon />
        </button>
      ) : null}
    </RadixCallout.Root>
  );
}

export function useCalloutToast(): {
  showToast: (props: CalloutProps) => void;
} {
  const { registerToast } = useToastProviderContext();

  const showToast = useCallback(
    (props: CalloutProps) => {
      registerToast(<Callout {...props} />);
    },
    [registerToast]
  );

  return {
    showToast,
  };
}
