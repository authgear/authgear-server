import React, { useCallback, useRef, useState, useEffect } from "react";
import { TextField as RadixTextField } from "@radix-ui/themes";
import { TextInput } from "../TextField/TextField";

import styles from "./ColorPickerField.module.css";

export type ColorHex = string;

// Note: Only the format of #xxxxxx is accepted by color input, so we do not handle other color format
const COLOR_REGEX = /^#?[0-9a-fA-F]{6}$/;

export interface ColorPickerFieldProps {
  value: ColorHex;
  onValueChange?: (value: ColorHex) => void;
}

export function ColorPickerField({
  value,
  onValueChange,
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
    <TextInput
      size="3"
      value={textInputValue}
      onChange={onTextInputChange}
      onBlur={onTextInputBlur}
    >
      <RadixTextField.Slot side="left">
        <ColorPicker value={value} onValueChange={onValueChange} />
      </RadixTextField.Slot>
    </TextInput>
  );
}

function ColorPicker({
  value,
  onValueChange,
}: {
  value: ColorHex;
  onValueChange?: (value: ColorHex) => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);

  const openPicker = useCallback(() => {
    inputRef.current?.click();
  }, []);

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
        className="h-0 w-0"
        onChange={handleColorInputChange}
      />
    </div>
  );
}
