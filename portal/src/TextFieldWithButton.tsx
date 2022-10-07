import React, { useCallback } from "react";
import PrimaryButton from "./PrimaryButton";
import TextField, { TextFieldProps } from "./TextField";
import styles from "./TextFieldWithButton.module.css";

export interface TextFieldWithButtonProps extends TextFieldProps {
  buttonText?: React.ReactNode;
  onButtonClick?: () => void;
}

const TextFieldWithButton: React.VFC<TextFieldWithButtonProps> =
  function TextFieldWithButton(props: TextFieldWithButtonProps) {
    const { buttonText, onButtonClick, ...rest } = props;

    const onClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        onButtonClick?.();
      },
      [onButtonClick]
    );

    return (
      <div className={styles.container}>
        <TextField className={styles.textField} {...rest} />
        <PrimaryButton
          className={styles.button}
          onClick={onClick}
          text={buttonText}
        />
      </div>
    );
  };

export default TextFieldWithButton;
