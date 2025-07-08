import React from "react";
import {
  // eslint-disable-next-line no-restricted-imports
  TextField as FluentUITextField,
  ITextFieldProps,
  useTheme,
} from "@fluentui/react";
import { useMergedStyles } from "./util/mergeStyles";

export interface TextFieldProps extends ITextFieldProps {}

const TextField: React.VFC<TextFieldProps> = function TextField(
  props: TextFieldProps
) {
  const { styles: stylesProp, ...rest } = props;
  const theme = useTheme();
  const styles = useMergedStyles(
    {
      field: {
        "::placeholder": {
          color: theme.palette.neutralTertiary,
        },
        backgroundColor: props.readOnly
          ? theme.palette.neutralLight
          : undefined,
      },
      fieldGroup: props.readOnly
        ? {
            border: "none",
          }
        : undefined,
      // only apply margin bottom to wrapper when there is description
      wrapper: props.description
        ? {
            marginBottom: "8px",
          }
        : undefined,
      description: {
        fontSize: "14px",
        color: theme.semanticColors.bodyText,
        lineHeight: "20px",
      },
    },
    stylesProp
  );
  // @ts-expect-error
  return <FluentUITextField styles={styles} {...rest} />;
};

export default TextField;
