import React, { useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { CSSColor } from "../../model/themeAuthFlowV2";
import { parseCSSColor, rgbaOrHexString } from "../../util/shades";

import styles from "./ColorPicker.module.css";

// Accept every CSS color format the previous FluentUI picker accepted
// (#rgb, #rrggbb, rgb(a)(), hsl(a)()) and normalise it the same way, so
// existing themes written in those formats keep working.
function normalizeColor(value: string): string | undefined {
  const rgba = parseCSSColor(value.trim());
  if (rgba == null) {
    return undefined;
  }
  return rgbaOrHexString(rgba.r, rgba.g, rgba.b, rgba.a);
}

// The native color input only understands #rrggbb, so anything with alpha
// falls back to the opaque form of the same color.
function toHexColor(color: string | undefined, fallback: string): string {
  const rgba = parseCSSColor(color ?? "") ?? parseCSSColor(fallback);
  if (rgba == null) {
    return "#000000";
  }
  return rgbaOrHexString(rgba.r, rgba.g, rgba.b, 100);
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
      onChange(normalizeColor(value));
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
