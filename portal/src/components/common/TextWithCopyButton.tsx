import React from "react";
import { IconButton } from "@fluentui/react";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import styles from "./TextWithCopyButton.module.css";

interface TextWithCopyButtonProps {
  text: string;
  TextComponent?: (props: {
    children: React.ReactNode | undefined;
  }) => React.ReactElement;
}

const TextWithCopyButton: React.VFC<TextWithCopyButtonProps> =
  function TextWithCopyButton({
    text,
    TextComponent,
  }: TextWithCopyButtonProps) {
    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: text,
    });

    const C = TextComponent ?? "span";

    return (
      <div className={styles.textWithCopyButton}>
        <C>{text}</C>
        <IconButton {...copyButtonProps} />
        <Feedback />
      </div>
    );
  };

export { TextWithCopyButton };
