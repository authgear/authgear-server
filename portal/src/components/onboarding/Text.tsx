import React from "react";
import cn from "classnames";
import styles from "./Text.module.css";

interface TextProps {
  className?: string;
  children?: React.ReactNode;
}

function Heading({ className, children }: TextProps): React.ReactElement {
  return <span className={cn(styles.heading, className)}>{children}</span>;
}

function Body({ className, children }: TextProps): React.ReactElement {
  return <span className={cn(styles.body, className)}>{children}</span>;
}

export const Text = {
  Heading,
  Body,
};
