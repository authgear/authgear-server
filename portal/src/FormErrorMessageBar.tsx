import React from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useFormTopErrors } from "./form";

export interface FormErrorMessageBarProps {
  children?: React.ReactNode;
}

export const FormErrorMessageBar: React.VFC<FormErrorMessageBarProps> = (
  props: FormErrorMessageBarProps
) => {
  const errors = useFormTopErrors();
  if (errors.length === 0) {
    return <>{props.children}</>;
  }

  return (
    <MessageBar messageBarType={MessageBarType.error}>
      {errors.map((err, i) => (
        <Text key={i}>
          <FormattedMessage id={err.messageID ?? ""} values={err.arguments} />
        </Text>
      ))}
    </MessageBar>
  );
};
