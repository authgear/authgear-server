import React, {
  useState,
  useCallback,
  useRef,
  useMemo,
  useEffect,
} from "react";
import cn from "classnames";
import {
  Callout,
  ColorPicker,
  IColorPickerProps,
  TextField,
  getColorFromString,
} from "@fluentui/react";
import styles from "./PortalColorPicker.module.scss";

export interface PortalColorPickerProps {
  className?: string;
  disabled?: boolean;
  color: string;
  onChange: (color: string) => void;
  alphaType?: IColorPickerProps["alphaType"];
}

const PortalColorPicker: React.FC<PortalColorPickerProps> = function PortalColorPicker(
  props: PortalColorPickerProps
) {
  const {
    className,
    disabled,
    color,
    onChange,
    alphaType: alphaTypeProp,
  } = props;
  const [colorStr, setColorStr] = useState<string | undefined>();
  const [isColorPickerVisible, setIsColorPickerVisible] = useState(false);
  const colorboxRef = useRef(null);
  const alphaType = alphaTypeProp ?? "none";

  // Set text field value when color changes.
  useEffect(() => {
    setColorStr(color);
  }, [color]);

  const onTextFieldChange = useCallback(
    (_e, newValue) => {
      if (newValue == null) {
        return;
      }

      const newColor = getColorFromString(newValue);
      if (newColor != null) {
        onChange(newColor.str);
        setColorStr(undefined);
      } else {
        setColorStr(newValue);
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

  const onColorboxClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (disabled) {
        return;
      }
      setIsColorPickerVisible(true);
    },
    [disabled]
  );

  const onCalloutDismiss = useCallback(() => {
    setIsColorPickerVisible(false);
  }, []);

  const iColor = useMemo(() => {
    return getColorFromString(color)!;
  }, [color]);

  return (
    <div className={cn(className, styles.root)}>
      <div
        ref={colorboxRef}
        className={styles.colorbox}
        style={{
          backgroundColor: color,
        }}
        onClick={onColorboxClick}
      />
      <TextField
        disabled={disabled}
        className={styles.textField}
        value={colorStr != null ? colorStr : color}
        onChange={onTextFieldChange}
      />
      {isColorPickerVisible && (
        <Callout
          gapSpace={10}
          target={colorboxRef.current}
          onDismiss={onCalloutDismiss}
        >
          <ColorPicker
            color={iColor}
            onChange={onColorPickerChange}
            alphaType={alphaType}
          />
        </Callout>
      )}
    </div>
  );
};

export default PortalColorPicker;
