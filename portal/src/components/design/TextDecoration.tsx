import React, { useCallback, useMemo } from "react";
import cn from "classnames";

import ButtonToggleGroup, { Option } from "../common/ButtonToggleGroup";

import {
  AllTextDecorations,
  TextDecorationType,
} from "../../model/themeAuthFlowV2";

import styles from "./TextDecoration.module.css";

interface TextDecorationProps {
  value: TextDecorationType;
  onChange: (value: TextDecorationType) => void;
}

const TextDecoration: React.VFC<TextDecorationProps> = function TextDecoration(
  props
) {
  const { value, onChange } = props;
  const options = useMemo(
    () => AllTextDecorations.map((value) => ({ value })),
    []
  );

  const onSelectOption = useCallback(
    (option: Option<TextDecorationType>) => {
      onChange(option.value);
    },
    [onChange]
  );

  const renderOption = useCallback(
    (option: Option<TextDecorationType>, selected: boolean) => {
      return (
        <span
          className={cn(
            styles.icTextDecoration,
            (() => {
              switch (option.value) {
                case "none":
                  return styles.icTextDecorationNone;
                case "underline":
                  return styles.icTextDecorationUnderline;
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
    <ButtonToggleGroup
      value={value}
      options={options}
      onSelectOption={onSelectOption}
      renderOption={renderOption}
    ></ButtonToggleGroup>
  );
};

export default TextDecoration;
