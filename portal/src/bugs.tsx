import { IBasePickerStyles } from "@fluentui/react";

// The TagPicker renders a absolute div without specifying its position.
// This causes the screen containing it to expend beyond its intrinsic size.
// The workaround is to position that div on the top left corner.
export const fixTagPickerStyles: Partial<IBasePickerStyles> = {
  screenReaderText: {
    left: "0px",
    top: "0px",
  },
};
