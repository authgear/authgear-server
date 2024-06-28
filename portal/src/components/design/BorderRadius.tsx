import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { Context as MFContext } from "@oursky/react-messageformat";
import cn from "classnames";

import ButtonToggleGroup, { Option } from "../common/ButtonToggleGroup";
import TextField from "../../TextField";

import {
  AllBorderRadiusStyleTypes,
  BorderRadiusStyle,
  BorderRadiusStyleType,
} from "../../model/themeAuthFlowV2";

import styles from "./BorderRadius.module.css";

interface BorderRadiusProps {
  value: BorderRadiusStyle;
  onChange: (value: BorderRadiusStyle) => void;
}
const BorderRadius: React.VFC<BorderRadiusProps> = function BorderRadius(
  props
) {
  const { value, onChange } = props;
  const { renderToString } = useContext(MFContext);
  const options = useMemo(
    () => AllBorderRadiusStyleTypes.map((value) => ({ value })),
    []
  );

  const [radiusValue, setRadiusValue] = useState(() => {
    if (value.type !== "rounded") {
      return "0";
    }
    return value.radius;
  });

  useEffect(() => {
    if (value.type !== "rounded") {
      return;
    }
    setRadiusValue(value.radius);
  }, [value, radiusValue]);

  const onSelectOption = useCallback(
    (option: Option<BorderRadiusStyleType>) => {
      if (option.value === "rounded") {
        onChange({
          type: option.value,
          radius: radiusValue,
        });
      } else {
        onChange({
          type: option.value,
        });
      }
    },
    [radiusValue, onChange]
  );

  const onBorderRadiusChange = useCallback(
    (_: any, value?: string) => {
      if (value == null) {
        return;
      }
      onChange({
        type: "rounded",
        radius: value,
      });
    },
    [onChange]
  );

  const renderOption = useCallback(
    (option: Option<BorderRadiusStyleType>, selected: boolean) => {
      return (
        <span
          className={cn(
            styles.icBorderRadius,
            (() => {
              switch (option.value) {
                case "none":
                  return styles.icBorderRadiusSquare;
                case "rounded":
                  return styles.icBorderRadiusRounded;
                case "rounded-full":
                  return styles.icBorderRadiusFullRounded;
                default:
                  return undefined;
              }
            })(),
            selected && styles.selected
          )}
        ></span>
      );
    },
    []
  );

  return (
    <div>
      <ButtonToggleGroup
        value={value.type}
        options={options}
        onSelectOption={onSelectOption}
        renderOption={renderOption}
      ></ButtonToggleGroup>
      {value.type === "rounded" ? (
        <TextField
          className={cn("mt-3")}
          label={renderToString(
            "DesignScreen.configuration.borderRadius.label"
          )}
          value={radiusValue}
          onChange={onBorderRadiusChange}
        />
      ) : null}
    </div>
  );
};

export default BorderRadius;
