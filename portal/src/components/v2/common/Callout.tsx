import { InfoCircledIcon } from "@radix-ui/react-icons";
import { Callout as RadixCallout } from "@radix-ui/themes";
import React, { ComponentProps, useCallback } from "react";
import styles from "./Callout.module.css";
import { useToastContext } from "./Toast";

export enum CalloutColor {
  error = "error",
  success = "success",
}

export interface CalloutProps {
  color: CalloutColor;
  text?: React.ReactChild;
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

export function Callout({ color, text }: CalloutProps): React.ReactElement {
  return (
    <RadixCallout.Root
      className={styles.calloutRoot}
      color={colorToRadixColor(color)}
      size="2"
      variant="surface"
    >
      <RadixCallout.Icon>
        <InfoCircledIcon />
      </RadixCallout.Icon>
      <RadixCallout.Text>{text}</RadixCallout.Text>
    </RadixCallout.Root>
  );
}

export function useCalloutToast(): {
  showToast: (props: CalloutProps) => void;
} {
  const { registerToast } = useToastContext();

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
