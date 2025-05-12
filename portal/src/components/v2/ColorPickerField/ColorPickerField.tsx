import React, { useCallback, useRef, useState, useEffect } from "react";
import { TextField as RadixTextField } from "@radix-ui/themes";
import { TextField } from "../TextField/TextField";

import styles from "./ColorPickerField.module.css";
import { FormField } from "../FormField/FormField";

export type ColorHex = string;

type ColorPickerFieldSize = "2" | "3";

// Note: Only the format of #xxxxxx is accepted by color input, so we do not handle other color format
const COLOR_REGEX = /^#?[0-9a-fA-F]{6}$/;

export interface ColorPickerFieldProps {
  darkMode?: boolean;
  size: ColorPickerFieldSize;
  disabled?: boolean;
  placeholder?: string;
  optional?: boolean;
  label?: React.ReactNode;
  error?: React.ReactNode;
  hint?: React.ReactNode;
  value: ColorHex;
  onValueChange?: (value: ColorHex) => void;
  onFocus?: React.FocusEventHandler<HTMLInputElement>;
  onOpenPicker?: () => void;
}

export function ColorPickerField({
  darkMode,
  size,
  disabled,
  placeholder,
  optional,
  label,
  error,
  hint,
  value,
  onValueChange,
  onFocus,
  onOpenPicker,
}: ColorPickerFieldProps): React.ReactElement {
  const [textInputValue, setTextInputValue] = useState(value);
  const onTextInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      let value = e.currentTarget.value;
      setTextInputValue(value);
      if (COLOR_REGEX.test(value)) {
        if (!value.startsWith("#")) {
          value = "#" + value;
        }
        onValueChange?.(value);
      }
    },
    [onValueChange]
  );

  const onTextInputBlur = useCallback(
    (_: React.FormEvent<HTMLInputElement>) => {
      setTextInputValue(value);
    },
    [value]
  );

  useEffect(() => {
    setTextInputValue(value);
  }, [value]);

  return (
    <FormField
      darkMode={darkMode}
      size={size}
      label={label}
      optional={optional}
      error={error}
      hint={hint}
      labelSpace="1"
    >
      <TextField.Input
        size={size}
        value={textInputValue}
        disabled={disabled}
        placeholder={placeholder}
        onChange={onTextInputChange}
        onBlur={onTextInputBlur}
        onFocus={onFocus}
      >
        <RadixTextField.Slot side="left">
          <ColorPicker
            value={value}
            onValueChange={onValueChange}
            onOpen={onOpenPicker}
          />
        </RadixTextField.Slot>
      </TextField.Input>
    </FormField>
  );
}

function ColorPicker({
  value,
  onValueChange,
  onOpen,
}: {
  value: ColorHex;
  onValueChange?: (value: ColorHex) => void;
  onOpen?: () => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);

  const openPicker = useCallback(() => {
    onOpen?.();
    inputRef.current?.click();
  }, [onOpen]);

  const handleColorInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const el = e.currentTarget;
      onValueChange?.(el.value);
    },
    [onValueChange]
  );

  return (
    <div
      className={styles.colorPickerField__pickerContainer}
      style={{ backgroundColor: value }}
    >
      <button
        type="button"
        className={styles.colorPickerField__pickerButton}
        onClick={openPicker}
      />
      <input
        ref={inputRef}
        type="color"
        className="h-0 w-0 appearance-none border-none"
        onChange={handleColorInputChange}
      />
    </div>
  );
}
