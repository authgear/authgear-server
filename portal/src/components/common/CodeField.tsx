import React from "react";
import { Text } from "@fluentui/react";
import styles from "./CodeField.module.css";

export interface CodeFieldProps {
  label?: string;
  description?: string;
  children?: React.ReactNode;
}

export function CodeField({
  label,
  description,
  children,
}: CodeFieldProps): React.ReactElement {
  return (
    <div className="">
      {label != null ? (
        <Text block={true} variant="medium" className="font-semibold leading-5">
          {label}
        </Text>
      ) : null}
      <code className={styles.code}>
        <Text>{children}</Text>
      </code>
      {description != null ? (
        <Text block={true} variant="medium" className="mt-2 leading-5">
          {description}
        </Text>
      ) : null}
    </div>
  );
}
