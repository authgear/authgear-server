import React, { useState, useCallback, useRef, useMemo } from "react";
import {
  Callout,
  ColorPicker,
  TextField,
  getColorFromString,
} from "@fluentui/react";
import styles from "./PortalColorPicker.module.scss";

export interface PortalColorPickerProps {
  color: string;
  onChange: (color: string) => void;
}

const PortalColorPicker: React.FC<PortalColorPickerProps> = function PortalColorPicker(
  props: PortalColorPickerProps
) {
  const { color, onChange } = props;

  const [colorStr, setColorStr] = useState(color);
  const [isColorPickerVisible, setIsColorPickerVisible] = useState(false);
  const colorboxRef = useRef(null);

  const onTextFieldChange = useCallback(
    (_e, newValue) => {
      if (newValue == null) {
        return;
      }
      setColorStr(newValue);
      const newColor = getColorFromString(newValue);
      if (newColor != null) {
        onChange(newColor.str);
      }
    },
    [onChange]
  );

  const onColorPickerChange = useCallback(
    (_e, newColor) => {
      setColorStr(newColor.str);
      onChange(newColor.str);
    },
    [onChange]
  );

  const onColorboxClick = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setIsColorPickerVisible(true);
  }, []);

  const onCalloutDismiss = useCallback(() => {
    setIsColorPickerVisible(false);
  }, []);

  const iColor = useMemo(() => {
    return getColorFromString(color)!;
  }, [color]);

  return (
    <>
      <div className={styles.root}>
        <div
          className={styles.colorbox}
          style={{
            backgroundColor: color,
          }}
          onClick={onColorboxClick}
        />
        <TextField
          className={styles.textField}
          value={colorStr}
          onChange={onTextFieldChange}
        />
      </div>
      {isColorPickerVisible && (
        <Callout
          gapSpace={10}
          target={colorboxRef.current}
          onDismiss={onCalloutDismiss}
        >
          <ColorPicker
            color={iColor}
            onChange={onColorPickerChange}
            alphaSliderHidden={true}
          />
        </Callout>
      )}
    </>
  );
};

export default PortalColorPicker;
