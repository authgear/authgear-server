import React, { FormEvent, useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { Label, Slider } from "@fluentui/react";
import TextField from "../../TextField";

interface AppLogoHeightSetterProps {
  /**
   * @type {string}
   * @example "40px"
   */
  value: string;
  onChange: (value: string) => void;
  minHeight?: number;
  maxHeight?: number;
  sliderAriaLabel?: string;
  className?: string;
}

const APP_LOGO_MIN_HEIGHT = 24;
const APP_LOGO_MAX_HEIGHT = 120;

const APP_LOGO_HEIGHT_INPUT_REGEX = /^[0-9]{0,3}$/;

const AppLogoHeightSetter: React.VFC<AppLogoHeightSetterProps> =
  function AppLogoHeightSetter(props) {
    const {
      value,
      onChange,
      sliderAriaLabel,
      minHeight,
      maxHeight,
      className,
    } = props;

    const [heightPX, setHeightPX] = useState(Number(value.replace("px", "")));

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
      <div className={cn(className, "flex items-center gap-x-2")}>
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
          className={cn("w-12.5")}
          onChange={onChangeInput}
          value={heightPX.toString()}
        />
        <Label>px</Label>
      </div>
    );
  };

export default AppLogoHeightSetter;
