import React, {
  ChangeEvent,
  PropsWithChildren,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import {
  Callout,
  ColorPicker as FluentUIColorPicker,
  getColorFromString,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import WidgetTitle from "../../../WidgetTitle";
import WidgetSubtitle from "../../../WidgetSubtitle";
import WidgetDescription from "../../../WidgetDescription";

import styles from "./DesignScreen.module.css";

export const Separator: React.VFC = function Separator() {
  return <div className={cn("h-px", "my-12", "bg-separator")}></div>;
};

interface ConfigurationGroupProps {
  labelKey: string;
}
export const ConfigurationGroup: React.VFC<
  PropsWithChildren<ConfigurationGroupProps>
> = function ConfigurationGroup(props) {
  const { labelKey } = props;
  return (
    <div className={cn("space-y-4")}>
      <WidgetTitle>
        <FormattedMessage id={labelKey} />
      </WidgetTitle>
      {props.children}
    </div>
  );
};

interface ConfigurationProps {
  labelKey: string;
}
export const Configuration: React.VFC<PropsWithChildren<ConfigurationProps>> =
  function Configuration(props) {
    const { labelKey } = props;
    return (
      <div>
        <WidgetSubtitle>
          <FormattedMessage id={labelKey} />
        </WidgetSubtitle>
        <div className={cn("mt-[0.3125rem]")}>{props.children}</div>
      </div>
    );
  };

interface ConfigurationDescriptionProps {
  labelKey: string;
}
export const ConfigurationDescription: React.VFC<ConfigurationDescriptionProps> =
  function ConfigurationDescription(props) {
    const { labelKey } = props;
    return (
      <WidgetDescription>
        <FormattedMessage id={labelKey} />
      </WidgetDescription>
    );
  };

export interface Option<T> {
  value: T;
}

interface ButtonToggleProps<T> {
  option: Option<T>;
  selected: boolean;
  renderOption: (
    option: Option<T>,
    selected: boolean
  ) => React.ReactElement | null;
  onClick: (o: Option<T>) => void;
}
export function ButtonToggle<T>(
  props: ButtonToggleProps<T>
): React.ReactElement {
  const { option, selected, renderOption, onClick } = props;
  const _onClick = (e: React.MouseEvent) => {
    e.preventDefault();
    onClick(option);
  };
  return (
    <button
      type="button"
      className={cn(
        "inline-flex",
        "items-center",
        "justify-center",
        "p-1.5",
        "hover:bg-neutral-lighter",
        selected && ["bg-neutral-light"]
      )}
      onClick={_onClick}
    >
      {renderOption(option, selected)}
    </button>
  );
}

function defaultKeyExtractor(o: Option<any>): string {
  return String(o.value);
}
interface ButtonToggleGroupProps<T> {
  className?: string;
  options: Option<T>[];
  onSelectOption: (option: Option<T>) => void;
  value: T;
  keyExtractor?: (option: Option<T>) => string;
  renderOption: (
    option: Option<T>,
    selected: boolean
  ) => React.ReactElement | null;
}
export function ButtonToggleGroup<T>(
  props: ButtonToggleGroupProps<T>
): React.ReactElement {
  const {
    options,
    onSelectOption,
    value,
    keyExtractor = defaultKeyExtractor,
    renderOption,
  } = props;
  return (
    <div
      className={cn(
        "inline-block",
        "rounded",
        "overflow-hidden",
        "border",
        "border-solid",
        "border-grey-grey110",
        props.className
      )}
    >
      {options.map((o) => (
        <ButtonToggle<T>
          key={keyExtractor(o)}
          option={o}
          selected={o.value === value}
          onClick={onSelectOption}
          renderOption={renderOption}
        />
      ))}
    </div>
  );
}

interface ColorPickerProps {
  className?: string;
  color: string;
  onChange: (color: string) => void;
}
export const ColorPicker: React.VFC<ColorPickerProps> = function ColorPicker(
  props
) {
  const { color, onChange } = props;

  const colorboxRef = useRef<HTMLDivElement | null>(null);

  const [inputValue, setInputValue] = useState(color);
  const [isColorPickerVisible, setIsColorPickerVisible] = useState(false);
  const [isFocusingInput, setIsFocusingInput] = useState(false);

  useEffect(() => {
    setInputValue(color);
  }, [color]);

  const onInputChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      setInputValue(e.currentTarget.value);
      const colorObject = getColorFromString(e.currentTarget.value);
      if (colorObject == null) {
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

  const colorObject = getColorFromString(color);
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
        style={{ backgroundColor: colorObject?.str }}
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
        onChange={onInputChange}
        onBlur={onBlurInput}
        onFocus={onFocusInput}
      />
      {isColorPickerVisible && colorObject != null ? (
        <Callout
          target={colorboxRef.current}
          gapSpace={10}
          onDismiss={hideColorPicker}
        >
          <FluentUIColorPicker
            color={colorObject}
            onChange={onColorPickerChange}
            alphaType="none"
          />
        </Callout>
      ) : null}
    </div>
  );
};
