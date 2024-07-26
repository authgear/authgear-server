import React from "react";
import cn from "classnames";

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
  disabled?: boolean;
}
export function ButtonToggle<T>(
  props: ButtonToggleProps<T>
): React.ReactElement {
  const { option, selected, renderOption, onClick, disabled } = props;
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
      disabled={disabled}
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
  disabled?: boolean;
  withBorder?: boolean;
}
function ButtonToggleGroup<T>(
  props: ButtonToggleGroupProps<T>
): React.ReactElement {
  const {
    options,
    onSelectOption,
    value,
    keyExtractor = defaultKeyExtractor,
    renderOption,
    disabled,
    withBorder = true,
  } = props;
  return (
    <div
      className={cn(
        "inline-block",
        "rounded",
        "overflow-hidden",
        {
          ["border"]: withBorder,
          ["border-solid"]: withBorder,
          ["border-grey-grey110"]: withBorder,
        },
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
          disabled={disabled}
        />
      ))}
    </div>
  );
}

export default ButtonToggleGroup;
