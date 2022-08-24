import React from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { useFormTopErrors } from "./form";

export const FormErrorMessageBar: React.VFC = (props) => {
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
