import React, { useContext } from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import { useFormTopErrors } from "./form";
import { renderError } from "./error/parse";

export const FormErrorMessageBar: React.FC = (props) => {
  const { renderToString } = useContext(Context);

  const errors = useFormTopErrors();
  if (errors.length === 0) {
    return <>{props.children}</>;
  }

  return (
    <MessageBar messageBarType={MessageBarType.error}>
      {errors.map((err, i) => (
        <Text key={i}>{renderError(err, renderToString)}</Text>
      ))}
    </MessageBar>
  );
};
