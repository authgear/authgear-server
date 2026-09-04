/* global IntlTelInputInitOptions IntlTelInputInstance JSX */
import React, { createRef } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import intlTelInput from "intl-tel-input";
import {
  cleanRawInputValue,
  trimCountryCallingCode,
  makePartialValue,
} from "./util/phone";
import styles from "./PhoneTextField.module.css";

export interface PhoneTextFieldValues {
  // Suppose the input now looks like
  // +852 23
  //
  // then
  //
  // rawInputValue === "23"
  // e164 === undefined
  // partialValue === "+85223"
  // alpha2 === "HK"
  // countryCallingCode === "852"
  rawInputValue: string;
  e164?: string;
  partialValue?: string;
  alpha2?: string;
  countryCallingCode?: string;
}

export interface PhoneTextFieldProps {
  className?: string;
  inputContainerClassName?: string;
  inputClassNameOverride?: string;
  label?: React.ReactNode;
  disabled?: boolean;
  pinnedList?: string[];
  allowlist?: string[];
  initialCountry?: string;
  initialInputValue: string;
  onChange: (values: PhoneTextFieldValues) => void;
  errorMessage?: React.ReactNode;
}

export default class PhoneTextField extends React.Component<PhoneTextFieldProps> {
  inputRef: React.RefObject<HTMLInputElement>;
  instance: IntlTelInputInstance | null;

  constructor(props: PhoneTextFieldProps) {
    super(props);
    this.inputRef = createRef();
    this.instance = null;
  }

  componentDidMount(): void {
    const options: IntlTelInputInitOptions = {
      autoPlaceholder: "aggressive",
      customContainer: styles.container,
      formatOnDisplay: false,
    };
    if (this.props.inputContainerClassName) {
      options.customContainer = `${options.customContainer} ${this.props.inputContainerClassName}`;
    }
    if (this.props.initialCountry != null) {
      // seems IntlTelInputInitOptions.initialCountry must be lowercase
      // https://github.com/jackocnr/intl-tel-input/blob/c53a32b4f39996d50cde4ffa0df37726e8435ec2/src/spec/tests/options/initialCountry.js#L15
      options.initialCountry = this.props.initialCountry.toLowerCase();
    }
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
      instance.setNumber(this.props.initialInputValue);
      this.instance = instance;
      this._formatInputValue(this.inputRef.current);

      this.inputRef.current.addEventListener("input", this.onInputChange);
      this.inputRef.current.addEventListener("blur", this.onInputBlur);
      this.inputRef.current.addEventListener(
        "countrychange",
        this.onCountryChange
      );
    }
  }

  onInputChange = (e: Event): void => {
    // Accept only + and digits.
    if (e.target instanceof HTMLInputElement) {
      const value = e.target.value;
      const cleaned = cleanRawInputValue(value);
      e.target.value = cleaned;
    }

    this.emitOnChange();
  };

  onInputBlur = (e: Event): void => {
    if (e.target instanceof HTMLInputElement) {
      this._formatInputValue(e.target);
    }
  };

  onCountryChange = (): void => {
    this.emitOnChange();
  };

  _formatInputValue(el: HTMLInputElement): void {
    const values = this.prepareValues();
    if (values != null) {
      const { countryCallingCode, rawInputValue } = values;
      if (countryCallingCode != null) {
        const value = trimCountryCallingCode(rawInputValue, countryCallingCode);
        el.value = value;
      }
    }
  }

  emitOnChange(): void {
    const values = this.prepareValues();
    if (values != null) {
      this.props.onChange(values);
    }
  }

  prepareValues(): PhoneTextFieldValues | undefined {
    if (this.instance != null && this.inputRef.current != null) {
      const rawInputValue = this.inputRef.current.value;
      const countryData = this.instance.getSelectedCountryData();
      const alpha2 = countryData.iso2;
      const countryCallingCode = countryData.dialCode;
      // The output of getNumber() is very unstable.
      // If isPossibleNumber(), then it has +countryCallingCode,
      // otherwise it is rawInputValue with some spaces in it.
      const maybeInvalid = this.instance.getNumber();
      let e164;
      if (this.instance.isPossibleNumber()) {
        if (maybeInvalid != null) {
          e164 = maybeInvalid;
        }
      }
      let partialValue;
      if (e164 != null) {
        partialValue = e164;
      } else if (countryCallingCode != null) {
        partialValue = makePartialValue(rawInputValue, countryCallingCode);
      }

      const values = {
        e164,
        rawInputValue,
        alpha2,
        countryCallingCode,
        partialValue,
      };
      return values;
    }

    return undefined;
  }

  render(): JSX.Element {
    const { className, label, errorMessage, disabled } = this.props;
    return (
      <div className={className}>
        {label != null ? (
          <Text as="label" size="2" weight="medium" className={styles.label}>
            {label}
          </Text>
        ) : null}
        <input
          className={cn(
            this.props.inputClassNameOverride ?? styles.input,
            errorMessage != null && styles.inputError
          )}
          type="text"
          ref={this.inputRef}
          disabled={disabled}
          pattern="^[\+0-9]*$"
        />
        {errorMessage ? (
          <Text as="p" size="1" color="red" className={styles.errorMessage}>
            {errorMessage}
          </Text>
        ) : null}
      </div>
    );
  }
}
