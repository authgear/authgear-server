import React, { useMemo } from "react";
import { Text, ITextProps, ITheme } from "@fluentui/react";
import { ErrorParseRule } from "./error/parse";
import { useErrorMessage } from "./formbinding";

export interface FormErrorMessageTextProps extends ITextProps {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  errorRules?: ErrorParseRule[];
}

function stylesFunc(_props: ITextProps, theme: ITheme) {
  return {
    root: {
      color: theme.semanticColors.errorText,
    },
  };
}

// FormErrorMessageText is a component to show form field error message.
// It is useful when the control being used does not support errorMessage prop out of the box.
// For TextField, use FormTextField instead.
// If you want to use a standard control with form error, create the FormXXX component yourself.
const FormErrorMessageText: React.FC<FormErrorMessageTextProps> =
  function FormErrorMessageText(props: FormErrorMessageTextProps) {
    const { parentJSONPointer, fieldName, errorRules, ...rest } = props;
    const field = useMemo(
      () => ({
        parentJSONPointer,
        fieldName,
        rules: errorRules,
      }),
      [parentJSONPointer, fieldName, errorRules]
    );
    const { errorMessage } = useErrorMessage(field);
    if (errorMessage == null) {
      return null;
    }
    return (
      <Text variant="small" styles={stylesFunc} {...rest}>
        {errorMessage}
      </Text>
    );
  };

export default FormErrorMessageText;
