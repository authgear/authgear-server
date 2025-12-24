import React from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { ParsedAPIError } from "./error/parse";

export interface ErrorRendererProps {
  error?: ParsedAPIError;
  errors?: readonly ParsedAPIError[];
}

const ErrorRenderer: React.VFC<ErrorRendererProps> = function ErrorRenderer(
  props: ErrorRendererProps
) {
  const { error, errors } = props;

  let errorArray: ParsedAPIError[] = [];
  if (error != null) {
    errorArray.push(error);
  }
  if (errors != null) {
    errorArray = [...errorArray, ...errors];
  }

  const children = [];
  for (let i = 0; i < errorArray.length; i++) {
    const e = errorArray[i];
    if (children.length > 0) {
      // If not the first error, add a comma
      children.push(", ");
    }
    if (e.messageID) {
      children.push(
        <FormattedMessage key={i} id={e.messageID ?? ""} values={e.arguments} />
      );
    } else {
      children.push(<>{e.message}</>);
    }
  }

  if (children.length === 0) {
    return null;
  }

  return <>{children}</>;
};

export default ErrorRenderer;
