import React, { useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { CSSColor } from "../../model/themeAuthFlowV2";

import styles from "./ColorPicker.module.css";

// Only #rrggbb is accepted by the native color input, so that is the format
// we validate against here.
const HEX_REGEX = /^#[0-9a-fA-F]{6}$/;

function toHexColor(color: string | undefined, fallback: string): string {
  if (color != null && HEX_REGEX.test(color)) {
    return color;
  }
  if (HEX_REGEX.test(fallback)) {
    return fallback;
  }
  return "#000000";
}

interface ColorPickerProps {
  className?: string;
  color: CSSColor | undefined;
  placeholderColor: CSSColor;
  onChange: (CSSColor: string | undefined) => void;
}
export const ColorPicker: React.VFC<ColorPickerProps> = function ColorPicker(
  props
) {
  const { className, color, placeholderColor, onChange } = props;

  const [inputValue, setInputValue] = useState<string>(color ?? "");
  const [isFocusingInput, setIsFocusingInput] = useState(false);

  useEffect(() => {
    if (color != null) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setInputValue(color);
    }
  }, [color]);

  const onInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      setInputValue(value);
      onChange(HEX_REGEX.test(value) ? value : undefined);
    },
    [onChange]
  );

  const onColorInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.currentTarget.value;
      setInputValue(value);
      onChange(value);
    },
    [onChange]
  );

  const onFocusInput = useCallback(() => {
    setIsFocusingInput(true);
  }, []);
  const onBlurInput = useCallback(() => {
    setIsFocusingInput(false);
    if (color == null) {
      // Clear the input value on blur if the value is not a valid color.
      setInputValue("");
    }
  }, [color]);

  return (
    <div
      className={cn(
        styles.colorPicker,
        isFocusingInput && styles.active,
        className
      )}
    >
      <div
        className={styles.swatch}
        style={{ backgroundColor: color ?? placeholderColor }}
      >
        {/* The input itself covers the swatch so the user's click lands
            directly on it. Safari only opens the native color panel for an
            input with a real rendered box; programmatically clicking a
            zero-size input does nothing there. */}
        <input
          type="color"
          className={styles.swatchInput}
          value={toHexColor(color, placeholderColor)}
          onChange={onColorInputChange}
          onFocus={onFocusInput}
          onBlur={onBlurInput}
        />
      </div>
      <input
        className={styles.textInput}
        type="text"
        value={inputValue}
        placeholder={placeholderColor}
        onChange={onInputChange}
        onBlur={onBlurInput}
        onFocus={onFocusInput}
      />
    </div>
  );
};
