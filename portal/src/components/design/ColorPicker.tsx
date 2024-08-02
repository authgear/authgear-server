import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  Callout,
  ColorPicker as FluentUIColorPicker,
  getColorFromString,
} from "@fluentui/react";
import cn from "classnames";
import { CSSColor } from "../../model/themeAuthFlowV2";

import styles from "./ColorPicker.module.css";

interface ColorPickerProps {
  className?: string;
  color: CSSColor | undefined;
  placeholderColor: CSSColor;
  onChange: (CSSColor: string | undefined) => void;
}
export const ColorPicker: React.VFC<ColorPickerProps> = function ColorPicker(
  props
) {
  const { color, placeholderColor, onChange } = props;

  const colorboxRef = useRef<HTMLDivElement | null>(null);

  const [inputValue, setInputValue] = useState<string>(color ?? "");
  const [isColorPickerVisible, setIsColorPickerVisible] = useState(false);
  const [isFocusingInput, setIsFocusingInput] = useState(false);

  useEffect(() => {
    if (color != null) {
      setInputValue(color);
    }
  }, [color]);

  const onInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setInputValue(e.currentTarget.value);
      const colorObject = getColorFromString(e.currentTarget.value);
      if (colorObject == null) {
        onChange(undefined);
        return;
      }
      onChange(colorObject.str);
    },
    [onChange]
  );

  const onFocusInput = useCallback(() => {
    setIsFocusingInput(true);
  }, []);
  const onBlurInput = useCallback(() => {
    setIsFocusingInput(false);
  }, []);

  const showColorPicker = useCallback(() => {
    setIsFocusingInput(true);
    setIsColorPickerVisible(true);
  }, []);
  const hideColorPicker = useCallback(() => {
    setIsFocusingInput(false);
    setIsColorPickerVisible(false);
  }, []);

  const onColorPickerChange = useCallback(
    (_e, newColor) => {
      setInputValue(newColor.str);
      onChange(newColor.str);
    },
    [onChange]
  );

  const colorObject = getColorFromString(color ?? "");
  const placeholderColorObject = getColorFromString(placeholderColor);
  return (
    <div className={cn(styles.colorPicker, isFocusingInput && styles.active)}>
      <div
        ref={colorboxRef}
        className={cn(
          "inline-block",
          "h-5",
          "w-5",
          "rounded",
          "overflow-hidden",
          "border",
          "border-solid",
          "border-neutral-tertiaryAlt"
        )}
        style={{
          backgroundColor: colorObject?.str ?? placeholderColorObject?.str,
        }}
        onClick={showColorPicker}
      ></div>
      <input
        className={cn(
          "ml-2",
          "flex-1",
          "h-full",
          "border-none",
          "outline-none"
        )}
        type="text"
        value={inputValue}
        placeholder={placeholderColor}
        onChange={onInputChange}
        onBlur={onBlurInput}
        onFocus={onFocusInput}
      />
      {isColorPickerVisible ? (
        <Callout
          target={colorboxRef.current}
          gapSpace={10}
          onDismiss={hideColorPicker}
        >
          <FluentUIColorPicker
            color={colorObject ?? placeholderColorObject ?? ""}
            onChange={onColorPickerChange}
            alphaType="none"
          />
        </Callout>
      ) : null}
    </div>
  );
};
