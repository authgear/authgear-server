import React from "react";
import cn from "classnames";
import { Text as RadixText } from "@radix-ui/themes";
import styles from "./Text.module.css";

interface TextProps {
  className?: string;
  children?: React.ReactNode;
}

function Heading({ className, children }: TextProps): React.ReactElement {
  return (
    <RadixText
      className={cn(styles.heading, className)}
      size="6"
      weight="medium"
    >
      {children}
    </RadixText>
  );
}

function HeadingHint({ className, children }: TextProps): React.ReactElement {
  return (
    <RadixText
      className={cn(styles.headingHint, className)}
      size="3"
      weight="medium"
    >
      {children}
    </RadixText>
  );
}

function Body({ className, children }: TextProps): React.ReactElement {
  return (
    <RadixText className={cn(styles.body, className)} size="4" weight="regular">
      {children}
    </RadixText>
  );
}

export const Text = {
  Heading,
  HeadingHint,
  Body,
};
