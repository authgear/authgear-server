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
  value: T | undefined;
  keyExtractor?: (option: Option<T>) => string;
  renderOption: (
    option: Option<T>,
    selected: boolean
  ) => React.ReactElement | null;
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

export default ButtonToggleGroup;
