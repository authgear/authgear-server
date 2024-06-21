import React, { PropsWithChildren } from "react";
import cn from "classnames";
import { FormattedMessage } from "@oursky/react-messageformat";
import WidgetTitle from "../../../WidgetTitle";
import WidgetSubtitle from "../../../WidgetSubtitle";

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
        {props.children}
      </div>
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
        "border-[#8A8886]",
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
