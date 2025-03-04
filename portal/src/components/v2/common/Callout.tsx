import { InfoCircledIcon, Cross2Icon } from "@radix-ui/react-icons";
import { Callout as RadixCallout } from "@radix-ui/themes";
import React, { ComponentProps, useCallback } from "react";
import styles from "./Callout.module.css";
import { useToastContext, useToastProviderContext } from "./Toast";

export enum CalloutColor {
  error = "error",
  success = "success",
}

export interface CalloutProps {
  color: CalloutColor;
  text?: React.ReactChild;
  showCloseButton?: boolean;
}

function colorToRadixColor(
  color: CalloutColor
): ComponentProps<typeof RadixCallout.Root>["color"] {
  switch (color) {
    case CalloutColor.error:
      return "red";
    case CalloutColor.success:
      return "green";
  }
}

export function Callout({
  color,
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
      color={colorToRadixColor(color)}
      size="2"
      variant="surface"
    >
      <RadixCallout.Icon className={styles.calloutIcon}>
        <InfoCircledIcon />
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
