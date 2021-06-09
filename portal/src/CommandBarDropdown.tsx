import React, { useMemo } from "react";
import {
  Icon,
  IDropdownProps,
  IDropdownOption,
  IDropdownStyleProps,
  Dropdown,
  IIconProps,
  IIconStyleProps,
} from "@fluentui/react";

import styles from "./CommandBarDropdown.module.scss";

export interface CommandBarDropdownProps extends IDropdownProps {
  iconProps?: IIconProps;
}

function getIconStyles(props: IIconStyleProps) {
  return {
    root: {
      color: props.theme?.palette.themePrimary,
      fontSize: "16px",
      margin: "0 4px",
    },
  };
}

function getDropdownStyles(props: IDropdownStyleProps) {
  return {
    dropdown: {
      height: "100%",
      // FIXME(style): figure out how to merge styles.
      width: "300px",
    },
    title: {
      height: "100%",
      border: "0",
      padding: "0 8px 0 0",
      color: props.theme?.semanticColors.buttonText,
    },
    caretDownWrapper: {
      top: "8px",
    },
  };
}

function Placeholder(props: CommandBarDropdownProps) {
  const { iconProps, placeholder } = props;
  return (
    <div className={styles.placeholder}>
      <Icon {...iconProps} styles={getIconStyles} />
      <span>{placeholder}</span>
    </div>
  );
}

function Title(props: CommandBarDropdownProps, options: IDropdownOption[]) {
  // FIXME: multiSelect is not supported :(
  const { iconProps } = props;
  const text = options[0].text;
  return (
    <div className={styles.placeholder}>
      <Icon {...iconProps} styles={getIconStyles} />
      <span>{text}</span>
    </div>
  );
}

// CommandBarDropdown is Dropdown component that looks like CommandBarButton.
// The primary usage is for placing a Dropdown in the CommandBar.
const CommandBarDropdown: React.FC<CommandBarDropdownProps> =
  function CommandBarDropdown(props: CommandBarDropdownProps) {
    const BoundTitle = useMemo(() => Title.bind(null, props), [props]);
    return (
      <Dropdown
        {...props}
        onRenderPlaceholder={Placeholder as any}
        onRenderTitle={BoundTitle as any}
        styles={getDropdownStyles}
      />
    );
  };

export default CommandBarDropdown;
