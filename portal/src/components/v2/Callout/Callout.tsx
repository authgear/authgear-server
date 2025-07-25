import cn from "classnames";
import {
  Cross2Icon,
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from "@radix-ui/react-icons";
import { Callout as RadixCallout } from "@radix-ui/themes";
import React, { useCallback } from "react";
import styles from "./Callout.module.css";
import { useMaybeToastContext, useToastProviderContext } from "../Toast/Toast";
import { semanticToRadixColor } from "../../../util/radix";

export type CalloutType = "error" | "success" | "warning";

export interface CalloutProps {
  className?: string;
  type: CalloutType;
  text?: React.ReactNode;
  showCloseButton?: boolean;
}

function typeToSemantic(type: CalloutType) {
  switch (type) {
    case "error":
      return "error";
    case "success":
      return "success";
    case "warning":
      return "warning";
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
  }
}

export function Callout({
  className,
  type,
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
      color={semanticToRadixColor(typeToSemantic(type))}
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
