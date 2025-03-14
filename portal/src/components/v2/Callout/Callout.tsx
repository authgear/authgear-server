import {
  Cross2Icon,
  CheckCircledIcon,
  ExclamationTriangleIcon,
} from "@radix-ui/react-icons";
import { Callout as RadixCallout } from "@radix-ui/themes";
import React, { useCallback } from "react";
import styles from "./Callout.module.css";
import { useMaybeToastContext, useToastProviderContext } from "../Toast/Toast";

export type CalloutType = "error" | "success";

export interface CalloutProps {
  type: CalloutType;
  text?: React.ReactNode;
  showCloseButton?: boolean;
}

function typeToColor(type: CalloutType): string {
  switch (type) {
    case "error":
      return "semantics-error";
    case "success":
      return "semantics-success";
  }
}

function CalloutIcon({ color }: { color: CalloutType }) {
  switch (color) {
    case "error":
      return <ExclamationTriangleIcon width="1rem" height="1rem" />;
    case "success":
      return <CheckCircledIcon width="1rem" height="1rem" />;
  }
}

export function Callout({
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
      className={styles.calloutRoot}
      data-accent-color={typeToColor(type)}
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
      registerToast(<Callout {...props} />);
    },
    [registerToast]
  );

  return {
    showToast,
  };
}
