import React from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import styles from "./CodeField.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";

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
  const { themes } = useSystemConfig();
  return (
    <div className={className}>
      {label != null ? (
        <Text block={true} variant="medium" className="font-semibold leading-5">
          {label}
        </Text>
      ) : null}
      <code className={cn(styles.code, codeClassName)}>
        {children ? (
          <Text styles={{ root: { fontFamily: "inherit" } }}>{children}</Text>
        ) : (
          <Text
            styles={{
              root: {
                fontFamily: "inherit",
                color: themes.main.palette.neutralTertiary,
              },
            }}
          >
            {placeholder}
          </Text>
        )}
      </code>
      {description != null ? (
        <Text block={true} variant="medium" className="mt-2 leading-5">
          {description}
        </Text>
      ) : null}
    </div>
  );
}
