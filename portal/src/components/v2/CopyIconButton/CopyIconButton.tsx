import React, {
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { IconButton, Tooltip } from "@radix-ui/themes";
import { CopyIcon } from "@radix-ui/react-icons";
import { Context } from "../../../intl";
import { copyToClipboard } from "../../../util/clipboard";
import styles from "./CopyIconButton.module.css";

export interface CopyIconButtonProps {
  textToCopy: string;
}

export function CopyIconButton({
  textToCopy,
}: CopyIconButtonProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current != null) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  const handleCopy = useCallback(() => {
    copyToClipboard(textToCopy);
    setCopied(true);
    if (timerRef.current != null) {
      clearTimeout(timerRef.current);
    }
    timerRef.current = setTimeout(() => {
      setCopied(false);
    }, 2000);
  }, [textToCopy]);

  return (
    <Tooltip
      content={
        copied
          ? renderToString("copied-to-clipboard")
          : renderToString("copy")
      }
      open={copied ? true : undefined}
    >
      <IconButton
        type="button"
        variant="ghost"
        color="gray"
        size="1"
        aria-label={renderToString("copy")}
        onClick={handleCopy}
        className={styles.copyIconButton}
      >
        <CopyIcon width="1rem" height="1rem" />
      </IconButton>
    </Tooltip>
  );
}
