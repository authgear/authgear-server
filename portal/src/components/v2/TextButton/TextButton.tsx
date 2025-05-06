import React from "react";
import cn from "classnames";
import { Spinner, SpinnerProps, Text } from "@radix-ui/themes";
import { ArrowLeftIcon } from "@radix-ui/react-icons";
import styles from "./TextButton.module.css";

export type TextButtonVariant = "default" | "secondary";

export type TextButtonSize = "3" | "4";

export enum TextButtonIcon {
  Back = "Back",
}

function sizeToIconDimension(size: TextButtonSize) {
  switch (size) {
    case "3":
      return 18;
    case "4":
      return 20;
  }
}

function Icon({
  icon,
  size,
}: {
  icon: TextButtonIcon;
  size: TextButtonSize;
}): React.ReactElement {
  const dimension = sizeToIconDimension(size);
  switch (icon) {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    case TextButtonIcon.Back:
      return <ArrowLeftIcon width={dimension} height={dimension} />;
  }
}

export interface TextButtonProps {
  variant: TextButtonVariant;
  size: TextButtonSize;
  darkMode?: boolean;
  disabled?: boolean;
  loading?: boolean;
  text?: React.ReactNode;
  iconStart?: TextButtonIcon;

  type?: "button" | "reset" | "submit";
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

export function TextButton({
  variant,
  size,
  darkMode,
  disabled,
  loading,
  text,
  iconStart,
  type = "button",
  onClick,
}: TextButtonProps): React.ReactElement {
  return (
    <button
      // eslint-disable-next-line react/button-has-type
      type={type}
      className={cn(
        styles.textButton,
        sizeToClassName(size),
        variantToClassName(variant),
        darkMode ? "dark" : null
      )}
      onClick={onClick}
      disabled={loading ? true : disabled}
    >
      <Text
        as="span"
        size={size}
        weight={"medium"}
        className={cn(
          styles.textButton__content,
          loading ? styles["textButton__content--hidden"] : null
        )}
      >
        {iconStart ? <Icon icon={iconStart} size={size} /> : null}
        {text}
      </Text>
      <span
        className={cn(
          styles["textButton__spinnerContainer"],
          !loading ? "invisible" : null
        )}
      >
        <Spinner size={sizeToSpinnerSize(size)} />
      </span>
    </button>
  );
}

function sizeToClassName(size: TextButtonSize): string {
  switch (size) {
    case "3":
      return styles["textButton--size3"];
    case "4":
      return styles["textButton--size4"];
  }
}

function variantToClassName(variant: TextButtonVariant): string {
  switch (variant) {
    case "default":
      return styles["textButton--default"];
    case "secondary":
      return styles["textButton--secondary"];
  }
}

function sizeToSpinnerSize(size: TextButtonSize): SpinnerProps["size"] {
  switch (size) {
    case "3":
      return "2";
    case "4":
      return "3";
  }
}
