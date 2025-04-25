import React from "react";
import { IconButton as RadixIconButton } from "@radix-ui/themes";
import { TrashIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";
import { semanticToRadixColor } from "../../../util/radix";
import styles from "./IconButton.module.css";

export enum IconButtonIcon {
  Trash = "Trash",
  MagnifyingGlass = "MagnifyingGlass",
}

export type IconButtonVariant = "default" | "destroy";

type IconButtonSize = "1" | "2" | "3";

export interface IconButtonProps {
  variant: IconButtonVariant;
  size: IconButtonSize;
  icon: IconButtonIcon;

  type?: "button" | "reset" | "submit";
  onClick?: React.MouseEventHandler<HTMLButtonElement>;
}

function toIconSizeClassName(size: IconButtonSize) {
  switch (size) {
    case "1":
      return styles["iconButton__icon--size1"];
    case "2":
      return styles["iconButton__icon--size2"];
    case "3":
      return styles["iconButton__icon--size3"];
  }
}

function toAccentColor(variant: IconButtonVariant) {
  switch (variant) {
    case "default":
      return undefined;
    case "destroy":
      return semanticToRadixColor("error");
  }
}

function Icon({ icon, size }: { icon: IconButtonIcon; size: IconButtonSize }) {
  const iconClassName = toIconSizeClassName(size);

  switch (icon) {
    case IconButtonIcon.Trash:
      return <TrashIcon className={iconClassName} />;
    case IconButtonIcon.MagnifyingGlass:
      return <MagnifyingGlassIcon className={iconClassName} />;
  }
}

export function IconButton({
  variant,
  size,
  icon,
  type = "button",
  onClick,
}: IconButtonProps): React.ReactElement {
  return (
    <RadixIconButton
      type={type}
      size={size}
      color={toAccentColor(variant)}
      onClick={onClick}
    >
      <Icon icon={icon} size={size} />
    </RadixIconButton>
  );
}
