import { useContext } from "react";
import { GraphQLError } from "graphql";
import { Context, Values } from "@oursky/react-messageformat";

import { APIError, isAPIError, isApolloError } from "./error";

export interface GenericErrorHandlingRule {
  errorMessageID: string;
  reason: APIError["reason"];
  kind?: string;
  cause?: string;
}

function constructErrorMessageFromGenericGraphQLError(
  renderToString: (messageID: string, values?: Values) => string,
  error: GraphQLError,
  rules: GenericErrorHandlingRule[]
): string | null {
  if (!isAPIError(error.extensions)) {
    return null;
  }

  const { extensions } = error;
  for (const rule of rules) {
    if (extensions.reason !== rule.reason) {
      continue;
    }
    // some error reason need special handling
    // depends on error info
    if (extensions.reason === "InvariantViolated") {
      const cause = extensions.info.cause;
      if (cause.kind === rule.cause) {
        return renderToString(rule.errorMessageID);
      }
      continue;
    }
    if (extensions.reason === "PasswordPolicyViolated") {
      const causes = extensions.info.causes;
      const causeNames = causes.map((cause) => cause.Name);
      if (rule.cause != null && causeNames.includes(rule.cause)) {
        return renderToString(rule.errorMessageID);
      }
      continue;
    }
    // for other error reason, only need to match reason
    return renderToString(rule.errorMessageID);
  }

  // no matching rules
  return null;
}

export function useGenericError(
  error: unknown,
  rules: GenericErrorHandlingRule[],
  fallbackErrorMessageID: string = "generic-error.unknown-error"
): string | undefined {
  const { renderToString } = useContext(Context);

  if (error == null) {
    return undefined;
  }

  const fallbackErrorMessage = renderToString(fallbackErrorMessageID);
  if (!isApolloError(error)) {
    console.warn("[Handle generic error]: Unhandled error\n", error);
    return fallbackErrorMessage;
  }

  const errorMessageList: string[] = [];
  let containUnrecognizedError = false;
  for (const graphQLError of error.graphQLErrors) {
    const errorMessage = constructErrorMessageFromGenericGraphQLError(
      renderToString,
      graphQLError,
      rules
    );
    if (errorMessage != null) {
      errorMessageList.push(errorMessage);
    } else {
      console.warn(
        "[Handle generic error]: Contains unrecognized graphQL error \n",
        graphQLError
      );
      containUnrecognizedError = true;
    }
  }
  if (containUnrecognizedError) {
    errorMessageList.push(fallbackErrorMessage);
  }

  return errorMessageList.join("\n");
}
