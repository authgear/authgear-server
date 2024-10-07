import { IconButton } from "@fluentui/react";
import React from "react";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useCopyFeedback } from "./hook/useCopyFeedback";
import TextField, { TextFieldProps } from "./TextField";
import styles from "./TextFieldWithCopyButton.module.css";

export interface TextFieldWithCopyButtonProps extends TextFieldProps {}

const TextFieldWithCopyButton: React.VFC<TextFieldWithCopyButtonProps> =
  function TextFieldWithCopyButton(props: TextFieldWithCopyButtonProps) {
    const { ...rest } = props;
    const { themes } = useSystemConfig();
    // eslint-disable-next-line no-useless-assignment
    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: props.value ?? "",
    });

    return (
      <div className={styles.container}>
        <TextField className={styles.textField} {...rest} />
        <IconButton
          {...copyButtonProps}
          className={styles.copyButton}
          theme={themes.actionButton}
        />
        <Feedback />
      </div>
    );
  };

export default TextFieldWithCopyButton;
