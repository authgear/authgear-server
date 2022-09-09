import React, { useMemo } from "react";
import {
  // eslint-disable-next-line no-restricted-imports
  TextField as FluentUITextField,
  ITextFieldProps,
  useTheme,
} from "@fluentui/react";

export interface TextFieldProps extends ITextFieldProps {}

const TextField: React.VFC<TextFieldProps> = function TextField(
  props: TextFieldProps
) {
  const { ...rest } = props;
  const theme = useTheme();
  const styles = useMemo(() => {
    let styles: TextFieldProps["styles"] = {
      field: {
        "::placeholder": {
          color: theme.palette.neutralTertiary,
        },
      },
    };
    if (props.description) {
      // only apply margin bottom to wrapper when there is description
      styles = {
        wrapper: {
          marginBottom: "8px",
        },
        description: {
          fontSize: "14px",
          color: theme.semanticColors.bodyText,
          lineHeight: "20px",
        },
        ...styles,
      };
    }
    if (props.readOnly) {
      styles = {
        field: {
          backgroundColor: theme.palette.neutralLight,
        },
        fieldGroup: {
          border: "none",
        },
        ...styles,
      };
    }
    styles = { ...styles, ...props.styles };
    return styles;
  }, [
    props.description,
    props.readOnly,
    props.styles,
    theme.semanticColors.bodyText,
    theme.palette.neutralLight,
    theme.palette.neutralTertiary,
  ]);

  return <FluentUITextField styles={styles} {...rest} />;
};

export default TextField;
