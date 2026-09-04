import React, { useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { Slider, Text } from "@radix-ui/themes";
import { TextField } from "../v2/TextField/TextField";
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
      // This causes infinite rerendering, because
      // invoking onChange causes the form state to change, and
      // therefore the onChange callback will change, and
      // finally this effect run again.
      //
      // The effect here should only take heightPX as deps because
      // what it want is when heightPX change, call onChange, but not call onChange when onChange change.
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [heightPX]);

    const onChangeInput = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const newValue = e.target.value;
        if (APP_LOGO_HEIGHT_INPUT_REGEX.test(newValue) === false) {
          return;
        }

        const newPX = Number(newValue);
        setHeightPX(newPX);
      },
      []
    );

    const onSliderChange = useCallback((values: number[]) => {
      if (values.length > 0) {
        setHeightPX(values[0]);
      }
    }, []);

    return (
      <Configuration labelKey={labelKey}>
        <div className={cn("flex", "items-center", "gap-4")}>
          <Slider
            className={cn("flex-1")}
            aria-label={sliderAriaLabel}
            value={[heightPX]}
            onValueChange={onSliderChange}
            min={minHeight ?? APP_LOGO_MIN_HEIGHT}
            max={maxHeight ?? APP_LOGO_MAX_HEIGHT}
            size="2"
            variant="classic"
          />
          <div className={cn("flex", "items-center", "gap-1", "shrink-0")}>
            <TextField
              size="2"
              type="number"
              value={heightPX.toString()}
              onChange={onChangeInput}
              inputClassName={cn("w-16")}
            />
            <Text as="span" size="2">
              px
            </Text>
          </div>
        </div>
      </Configuration>
    );
  };

export default AppLogoHeightSetter;
