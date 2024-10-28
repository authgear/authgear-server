import React, { FormEvent, useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { Label, Slider } from "@fluentui/react";
import TextField from "../../TextField";
import Configuration from "./Configuration";

const PIXEL_HEIGHT_REGEX = /^[0-9]+px$/;
const REM_HEIGHT_REGEX = /^[0-9]+rem$/;

const FALLBACK_HEIGHT_PX = 100;

/**
 * parseHeightString handles all css units
 * 1 rem -> 16 px
 * 1 px  ->  1 px
 * unidentified units -> 100 px
 *
 * @param {string} height
 * @param {?string} [defaultValue]
 * @returns {number}
 */
function parseHeightString(height: string, defaultValue?: string): number {
  if (PIXEL_HEIGHT_REGEX.test(height)) {
    return Number(height.replace("px", ""));
  }
  if (REM_HEIGHT_REGEX.test(height)) {
    return Number(height.replace("rem", "")) * 16;
  }

  if (defaultValue != null && PIXEL_HEIGHT_REGEX.test(defaultValue)) {
    return Number(defaultValue.replace("px", ""));
  }

  return FALLBACK_HEIGHT_PX;
}

interface AppLogoHeightSetterProps {
  /**
   * @type {string}
   * @example "40px"
   */
  value: string;
  defaultValue?: string;
  onChange: (value: string) => void;
  labelKey: string;
  minHeight?: number;
  maxHeight?: number;
  sliderAriaLabel?: string;
}

const APP_LOGO_MIN_HEIGHT = 24;
const APP_LOGO_MAX_HEIGHT = 120;

const APP_LOGO_HEIGHT_INPUT_REGEX = /^[0-9]{0,3}$/;

const AppLogoHeightSetter: React.VFC<AppLogoHeightSetterProps> =
  function AppLogoHeightSetter(props) {
    const {
      value,
      defaultValue,
      onChange,
      sliderAriaLabel,
      minHeight,
      maxHeight,
      labelKey,
    } = props;

    const [heightPX, setHeightPX] = useState(
      parseHeightString(value, defaultValue)
    );

    useEffect(() => {
      onChange(`${heightPX}px`);
    }, [heightPX, onChange]);

    const onChangeInput = useCallback(
      (
        _e: FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        newValue?: string
      ) => {
        if (newValue == null) {
          return;
        }
        if (APP_LOGO_HEIGHT_INPUT_REGEX.test(newValue) === false) {
          return;
        }

        const newPX = Number(newValue);
        setHeightPX(newPX);
      },
      []
    );

    return (
      <Configuration labelKey={labelKey}>
        <div className={cn("flex", "items-center", "gap-x-2")}>
          <Slider
            className={cn("flex-1")}
            aria-label={sliderAriaLabel}
            showValue={false}
            value={heightPX}
            onChange={setHeightPX}
            min={minHeight ?? APP_LOGO_MIN_HEIGHT}
            max={maxHeight ?? APP_LOGO_MAX_HEIGHT}
          />
          <TextField
            type="number"
            min={minHeight ?? APP_LOGO_MIN_HEIGHT}
            max={maxHeight ?? APP_LOGO_MAX_HEIGHT}
            step={1}
            onChange={onChangeInput}
            value={heightPX.toString()}
          />
          <Label>px</Label>
        </div>
      </Configuration>
    );
  };

export default AppLogoHeightSetter;
