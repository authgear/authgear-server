import { useContext } from "react";
import { GraphQLError } from "graphql";
import { Context, Values } from "@oursky/react-messageformat";

import { APIError, isAPIError, isApolloError } from "./error";

export interface GenericErrorHandlingRule {
  errorMessageID: string;
  reason: APIError["reason"];
  kind?: string;
  cause?: string;
  field?: string;
}

function constructErrorMessageFromGenericGraphQLError(
  renderToString: (messageID: string, values?: Values) => string,
  error: GraphQLError,
  rules: GenericErrorHandlingRule[]
): { errorMessage: string; violatedRule: GenericErrorHandlingRule } | null {
  if (!isAPIError(error.extensions)) {
    return null;
  }

  const { extensions } = error;
  for (const rule of rules) {
    if (extensions.reason !== rule.reason) {
      continue;
    }
    const matchedResult = {
      errorMessage: renderToString(rule.errorMessageID),
      violatedRule: rule,
    };
    // some error reason need special handling
    // depends on error info
    if (extensions.reason === "InvariantViolated") {
      const cause = extensions.info.cause;
      if (cause.kind === rule.kind) {
        return matchedResult;
      }
      continue;
    }
    if (extensions.reason === "PasswordPolicyViolated") {
      const causes = extensions.info.causes;
      const causeNames = causes.map((cause) => cause.Name);
      if (rule.cause != null && causeNames.includes(rule.cause)) {
        return matchedResult;
      }
      continue;
    }
    // for other error reason, only need to match reason
    return matchedResult;
  }

  // no matching rules
  return null;
}

export function useGenericError(
  error: unknown,
  rules: GenericErrorHandlingRule[],
  fallbackErrorMessageID: string = "generic-error.unknown-error"
): {
  errorMessage: string | undefined;
  errorMessageMap: Partial<Record<string, string>>;
  unrecognizedError?: unknown;
} {
  const { renderToString } = useContext(Context);

  if (error == null) {
    return { errorMessage: undefined, errorMessageMap: {} };
  }

  const fallbackErrorMessage = renderToString(fallbackErrorMessageID);
  if (!isApolloError(error)) {
    console.warn("[Handle generic error]: Unhandled error\n", error);
    return {
      errorMessage: fallbackErrorMessage,
      errorMessageMap: {},
      unrecognizedError: error,
    };
  }

  const errorMessageList: string[] = [];
  const errorMessageMap: Partial<Record<string, string>> = {};
  let containUnrecognizedError = false;
  for (const graphQLError of error.graphQLErrors) {
    const violation = constructErrorMessageFromGenericGraphQLError(
      renderToString,
      graphQLError,
      rules
    );
    if (violation != null) {
      const { errorMessage, violatedRule } = violation;
      errorMessageList.push(errorMessage);
      if (violatedRule.field != null) {
        errorMessageMap[violatedRule.field] =
          errorMessageMap[violatedRule.field] == null
            ? errorMessage
            : `${errorMessageMap[violatedRule.field]}\n${errorMessage}`;
      }
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

  return {
    errorMessage: errorMessageList.join("\n"),
    errorMessageMap,
    unrecognizedError: containUnrecognizedError ? error : undefined,
  };
}
