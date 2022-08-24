/* global IntlTelInputInitOptions IntlTelInputInstance JSX */
import React, { createRef } from "react";
import { Text, Label } from "@fluentui/react";
import intlTelInput from "intl-tel-input";
import { SystemConfigContext } from "./context/SystemConfigContext";
import styles from "./PhoneTextField.module.css";

export interface PhoneTextFieldProps {
  className?: string;
  label?: string;
  disabled?: boolean;
  pinnedList?: string[];
  allowlist?: string[];
  inputValue: string;
  onChange: (validValue: string, inputValue: string) => void;
  errorMessage?: React.ReactNode;
}

export default class PhoneTextField extends React.Component<PhoneTextFieldProps> {
  inputRef: React.RefObject<HTMLInputElement>;
  instance: IntlTelInputInstance | null;

  static contextType = SystemConfigContext;
  declare context: React.ContextType<typeof SystemConfigContext>;

  constructor(props: PhoneTextFieldProps) {
    super(props);
    this.inputRef = createRef();
    this.instance = null;
  }

  componentDidMount(): void {
    const options: IntlTelInputInitOptions = {
      autoPlaceholder: "aggressive",
      customContainer: styles.container,
    };
    if (this.props.allowlist != null) {
      options.onlyCountries = [...this.props.allowlist];
    }
    if (this.props.pinnedList != null) {
      options.preferredCountries = [...this.props.pinnedList];
    } else {
      options.preferredCountries = [];
    }

    if (this.inputRef.current != null) {
      const instance = intlTelInput(this.inputRef.current, options);
      instance.setNumber(this.props.inputValue);
      this.instance = instance;

      this.inputRef.current.addEventListener("input", this.onInputChange);
      this.inputRef.current.addEventListener(
        "countrychange",
        this.onCountryChange
      );
    }
  }

  componentDidUpdate(prevProps: PhoneTextFieldProps): void {
    if (prevProps.inputValue !== this.props.inputValue) {
      this.instance?.setNumber(this.props.inputValue);
    }
  }

  onInputChange = (): void => {
    this.emitOnChange();
  };

  onCountryChange = (): void => {
    this.emitOnChange();
  };

  emitOnChange(): void {
    if (this.instance != null && this.inputRef.current != null) {
      const isValid = this.instance.isValidNumber();
      if (isValid) {
        const valid = this.instance.getNumber();
        if (valid != null) {
          this.props.onChange(valid, this.inputRef.current.value);
        }
      } else {
        this.props.onChange("", this.inputRef.current.value);
      }
    }
  }

  render(): JSX.Element {
    const { className, label, errorMessage, disabled } = this.props;
    const semanticColors = this.context?.themes.main.semanticColors;
    const inputBorder = semanticColors?.inputBorder ?? "";
    const errorText = semanticColors?.errorText ?? "";
    const inputFocusBorderAlt = semanticColors?.inputFocusBorderAlt ?? "";
    const disabledBackground = semanticColors?.disabledBackground ?? "";
    return (
      <div className={className}>
        {label ? <Label disabled={disabled}>{label}</Label> : null}
        <input
          style={{
            // @ts-expect-error
            "--PhoneTextField-border-color":
              errorMessage != null ? errorText : inputBorder,
            "--PhoneTextField-border-color-focus":
              errorMessage != null ? errorText : inputFocusBorderAlt,
            backgroundColor: disabled ? disabledBackground : undefined,
          }}
          className={styles.input}
          type="text"
          ref={this.inputRef}
          disabled={disabled}
        />
        {errorMessage ? (
          <Text
            block={true}
            styles={{
              root: {
                color: errorText,
              },
            }}
            className={styles.errorMessage}
          >
            {errorMessage}
          </Text>
        ) : null}
      </div>
    );
  }
}
