import React, { useMemo } from "react";
import {
  TextField as FluentUITextField,
  ITextFieldProps,
} from "@fluentui/react";
import { useSystemConfig } from "./context/SystemConfigContext";

export interface TextFieldProps extends ITextFieldProps {}

const TextField: React.FC<TextFieldProps> = function TextField(
  props: TextFieldProps
) {
  const { ...rest } = props;
  const { themes } = useSystemConfig();
  const styles = useMemo(() => {
    let styles = {};
    if (props.description) {
      // only apply margin bottom to wrapper when there is description
      styles = {
        wrapper: {
          marginBottom: "8px",
        },
        description: {
          fontSize: "14px",
          color: themes.main.semanticColors.bodyText,
          lineHeight: "20px",
        },
        ...styles,
      };
    }
    if (props.readOnly) {
      styles = {
        field: {
          backgroundColor: themes.main.palette.neutralLight,
        },
        fieldGroup: {
          borderColor: themes.main.palette.neutralLight,
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
    themes.main.semanticColors.bodyText,
    themes.main.palette.neutralLight,
  ]);

  return <FluentUITextField styles={styles} {...rest} />;
};

export default TextField;
