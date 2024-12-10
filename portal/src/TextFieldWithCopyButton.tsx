import { IButtonProps, IconButton } from "@fluentui/react";
import React from "react";
import cn from "classnames";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useCopyFeedback } from "./hook/useCopyFeedback";
import TextField, { TextFieldProps } from "./TextField";
import styles from "./TextFieldWithCopyButton.module.css";

export interface TextFieldWithCopyButtonProps extends TextFieldProps {
  additionalIconButtons?: IButtonProps[];
}

const TextFieldWithCopyButton: React.VFC<TextFieldWithCopyButtonProps> =
  function TextFieldWithCopyButton(props: TextFieldWithCopyButtonProps) {
    const { disabled, additionalIconButtons, ...rest } = props;
    const { themes } = useSystemConfig();
    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: props.value ?? "",
    });

    return (
      <div className={styles.container}>
        <TextField className={styles.textField} disabled={disabled} {...rest} />
        <IconButton
          {...copyButtonProps}
          className={cn(
            styles.actionButton,
            disabled ? styles["actionButton--hide"] : null
          )}
          theme={themes.actionButton}
        />
        <Feedback />
        {additionalIconButtons?.map((props, idx) => {
          const { className, ...restProps } = props;
          return (
            <IconButton
              key={idx}
              theme={themes.actionButton}
              className={cn(styles.actionButton, className)}
              {...restProps}
            />
          );
        })}
      </div>
    );
  };

export default TextFieldWithCopyButton;
