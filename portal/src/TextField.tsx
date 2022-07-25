import React, { useMemo } from "react";
import {
  TextField as FluentUITextField,
  ITextFieldProps,
} from "@fluentui/react";
import { useSystemConfig } from "./context/SystemConfigContext";

export interface TextProps extends ITextFieldProps {}

const TextField: React.FC<ITextFieldProps> = function TextField(
  props: ITextFieldProps
) {
  const { ...rest } = props;
  const { themes } = useSystemConfig();
  const styles = useMemo(() => {
    return props.description
      ? {
          // only apply margin bottom to wrapper when there is description
          wrapper: {
            marginBottom: "8px",
          },
          description: {
            fontSize: "14px",
            color: themes.main.semanticColors.bodyText,
            lineHeight: "20px",
          },
          ...props.styles,
        }
      : props.styles;
  }, [props.description, props.styles, themes.main.semanticColors.bodyText]);

  return <FluentUITextField {...rest} styles={styles} />;
};

export default TextField;
