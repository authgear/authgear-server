import React from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import styles from "./CodeField.module.css";

export interface CodeFieldProps {
  className?: string;
  codeClassName?: string;
  label?: string;
  description?: string;
  placeholder?: React.ReactNode;
  children?: React.ReactNode;
}

export function CodeField({
  className,
  codeClassName,
  label,
  description,
  children,
  placeholder,
}: CodeFieldProps): React.ReactElement {
  return (
    <div className={className}>
      {label != null ? (
        <Text as="p" size="2" className="font-semibold leading-5">
          {label}
        </Text>
      ) : null}
      <code className={cn(styles.code, codeClassName)}>
        {children ? (
          children
        ) : (
          <span className={styles.placeholder}>{placeholder}</span>
        )}
      </code>
      {description != null ? (
        <Text as="p" size="2" className="mt-2 leading-5">
          {description}
        </Text>
      ) : null}
    </div>
  );
}
