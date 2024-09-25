import React, { useEffect, useState } from "react";
import cn from "classnames";
import { Label, Slider } from "@fluentui/react";

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
        <Label>px</Label>
      </div>
    );
  };

export default AppLogoHeightSetter;
