import cn from "classnames";
import {
  Cross2Icon,
  CheckCircledIcon,
  ExclamationTriangleIcon,
  InfoCircledIcon,
} from "@radix-ui/react-icons";
import { Callout as RadixCallout } from "@radix-ui/themes";
import React, { useCallback } from "react";
import styles from "./Callout.module.css";
import { useMaybeToastContext, useToastProviderContext } from "../Toast/Toast";
import { semanticToRadixColor } from "../../../util/radix";

export type CalloutType = "error" | "success" | "warning" | "info";

export interface CalloutProps {
  className?: string;
  type: CalloutType;
  color?: React.ComponentProps<typeof RadixCallout.Root>["color"];
  size?: "1" | "2" | "3";
  text?: React.ReactNode;
  showCloseButton?: boolean;
}

function typeToColor(type: CalloutType) {
  switch (type) {
    case "error":
      return semanticToRadixColor("error");
    case "success":
      return semanticToRadixColor("success");
    case "warning":
      return semanticToRadixColor("warning");
    case "info":
      return "blue" as const;
  }
}

function CalloutIcon({ color }: { color: CalloutType }) {
  switch (color) {
    case "error":
      return <ExclamationTriangleIcon width="1rem" height="1rem" />;
    case "success":
      return <CheckCircledIcon width="1rem" height="1rem" />;
    case "warning":
      return <ExclamationTriangleIcon width="1rem" height="1rem" />;
    case "info":
      return <InfoCircledIcon width="1rem" height="1rem" />;
  }
}

export function Callout({
  className,
  type,
  color,
  size = "2",
  text,
  showCloseButton = true,
}: CalloutProps): React.ReactElement {
  const toastContext = useMaybeToastContext();

  const onClose = useCallback(() => {
    toastContext?.setOpen(false);
  }, [toastContext]);

  return (
    <RadixCallout.Root
      className={cn(styles.calloutRoot, className)}
      color={color ?? typeToColor(type)}
      size={size}
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
          <Cross2Icon width="1rem" height="1rem" />
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
      registerToast(
        <Callout
          {...props}
          className={cn(props.className, styles["calloutRoot--toast"])}
        />
      );
    },
    [registerToast]
  );

  return {
    showToast,
  };
}
