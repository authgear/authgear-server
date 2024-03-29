import { ICommandBarItemProps } from "@fluentui/react";
import React from "react";
import PrimaryButton from "./PrimaryButton";

export function onRenderCommandBarPrimaryButton(
  item?: ICommandBarItemProps
): React.ReactNode | undefined {
  if (item == null) {
    return null;
  }
  return (
    <PrimaryButton
      styles={{
        root: {
          padding: "0 16px",
          margin: "6px 4px",
        },
        // https://github.com/authgear/authgear-server/issues/2348#issuecomment-1226545493
        label: {
          whiteSpace: "nowrap",
        },
      }}
      iconProps={item.iconProps}
      disabled={item.disabled}
      text={item.text}
      className={item.className}
      onClick={(e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();
        item.onClick?.();
      }}
    />
  );
}
