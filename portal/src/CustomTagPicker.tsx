import React from "react";
import {
  TagPicker as FluentUITagPicker,
  ITagPickerProps,
} from "@fluentui/react";

interface CustomTagPickerProps extends ITagPickerProps {
  onAdd?: (item: string) => void;
}

// CustomTagPicker is workaround of fluentui TagPicker problem that item doesn't
// add when on blur
// CustomTagPicker will call onAdd when user off focus the input
const CustomTagPicker: React.FC<CustomTagPickerProps> = function CustomTagPicker(
  props: CustomTagPickerProps
) {
  const { onAdd, ...rest } = props;

  return (
    <FluentUITagPicker
      {...rest}
      onBlur={(e: React.FocusEvent<HTMLInputElement>) => {
        if (onAdd) {
          // https://stackoverflow.com/questions/23892547/what-is-the-best-way-to-trigger-onchange-event-in-react-js
          /* @ts-expect-error */
          // eslint-disable-next-line @typescript-eslint/unbound-method
          const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
            HTMLInputElement.prototype,
            "value"
          ).set;
          if (nativeInputValueSetter) {
            onAdd(e.target.value);

            // clear the input value
            nativeInputValueSetter.call(e.target, "");
            const ev2 = new Event("input", { bubbles: true });
            e.target.dispatchEvent(ev2);
          }
        }
      }}
    />
  );
};

export default CustomTagPicker;
