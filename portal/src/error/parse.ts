import { ApolloError, ServerError } from "@apollo/client";
import { APIError, isAPIError } from "./error";
import { ValidationFailedErrorInfoCause } from "./validation";
import { Values } from "@oursky/react-messageformat";
import { APIPasswordPolicyViolatedError } from "./password";
import { APIResourceTooLargeError } from "./resources";
import {
  parentChildToJSONPointer,
  matchParentChild,
} from "../util/jsonpointer";

export interface FormField {
  parentJSONPointer: string | RegExp;
  fieldName: string;
  rules?: ErrorParseRule[];
}

export interface ParsedAPIError {
  message?: string;
  messageID?: string;
  arguments?: Values;
}

export function parseRawError(error: unknown): APIError[] {
  const errors: APIError[] = [];
  if (error instanceof ApolloError) {
    if (error.networkError) {
      if ((error.networkError as ServerError).statusCode === 413) {
        // the request may be blocked by the load balancer if the payload is too large
        // e.g. upload a large image as the app logo
        errors.push({
          reason: "RequestEntityTooLarge",
          errorName: "RequestEntityTooLarge",
        });
      } else {
        errors.push({ reason: "NetworkFailed", errorName: "NetworkFailed" });
      }
    }
    for (const e of error.graphQLErrors) {
      if (isAPIError(e.extensions)) {
        errors.push({
          ...e.extensions,
          message: e.message,
        });
      } else {
        errors.push({
          reason: "Unknown",
          errorName: "Unknown",
          message: e.message,
        });
      }
    }
  } else if (error instanceof Error) {
    errors.push({
      reason: "Unknown",
      errorName: "Unknown",
      message: error.message,
    });
  } else if (typeof error === "object" && isAPIError(error)) {
    errors.push(error);
  } else if (error) {
    errors.push({
      reason: "Unknown",
      errorName: "Unknown",
      message: String(error),
    });
  }
  return errors;
}

const errorCauseMessageIDs = {
  required: "errors.validation.required",
  general: "errors.validation.general",
  format: "errors.validation.format",
  pattern: "errors.validation.pattern",
  minItems: "errors.validation.minItems",
  uniqueItems: "errors.validation.uniqueItems",
  minimum: "errors.validation.minimum",
  maximum: "errors.validation.maximum",
  minLength: "errors.validation.minLength",
  maxLength: "errors.validation.maxLength",
  blocked: "errors.validation.blocked",
  noPrimaryAuthenticator: "errors.validation.noPrimaryAuthenticator",
  noSecondaryAuthenticator: "errors.validation.noSecondaryAuthenticator",
};

function isKnownCauseKind(kind: string): boolean {
  if (kind === "__local") {
    return true;
  }
  return kind in errorCauseMessageIDs;
}

function parseCause(cause: ValidationFailedErrorInfoCause): ParsedAPIError {
  if (cause.kind === "general") {
    return { message: cause.details.msg };
  } else if (cause.kind === "__local") {
    return cause.details.error;
  }

  const messageID = errorCauseMessageIDs[cause.kind];
  if (!messageID) {
    return {
      messageID: "errors.validation.unknown",
      arguments: cause as unknown as Values,
    };
  }
  const args: Values = cause.details;
  return { messageID, arguments: args };
}

function parsePasswordPolicyViolatedError(
  error: APIPasswordPolicyViolatedError
): ParsedAPIError[] {
  const errors: ParsedAPIError[] = [];
  let hasUnmatched = false;

  for (const cause of error.info.causes) {
    switch (cause.Name) {
      case "PasswordReused":
        errors.push({
          messageID: "errors.password-policy.password-reused",
        });
        break;
      case "PasswordContainingExcludedKeywords":
        errors.push({
          messageID: "errors.password-policy.containing-excluded-keywords",
        });
        break;
      default:
        hasUnmatched = true;
        break;
    }
  }

  if (hasUnmatched && errors.length === 0) {
    errors.push({ messageID: "errors.password-policy.unknown" });
  }

  return errors;
}

function parseResourceTooLargeError(
  error: APIResourceTooLargeError
): ParsedAPIError {
  const dir = error.info.path.split("/");
  const fileName = dir[dir.length - 1];
  const resourceType = fileName.slice(0, fileName.lastIndexOf("."));

  return {
    messageID: "errors.resource-too-large",
    arguments: {
      maxSize: error.info.max_size / 1024,
      resourceType,
    },
  };
}

// eslint-disable-next-line complexity
function parseError(error: APIError): ParsedAPIError[] {
  const errors: ParsedAPIError[] = [];
  switch (error.reason) {
    case "ValidationFailed":
      errors.push(...error.info.causes.map((c) => parseCause(c)));
      break;
    case "NetworkFailed":
      errors.push({ messageID: "errors.network" });
      break;
    case "RequestEntityTooLarge":
      errors.push({ messageID: "errors.request-entity-too-large" });
      break;
    case "Unknown":
      errors.push({
        messageID: "errors.unknown",
        arguments: { message: error.message ?? "" },
      });
      break;
    case "ResourceTooLarge": {
      errors.push(parseResourceTooLargeError(error));
      break;
    }
    case "PasswordPolicyViolated": {
      errors.push(...parsePasswordPolicyViolatedError(error));
      break;
    }
    case "WebHookDisallowed": {
      errors.push({
        messageID: "errors.webhook.disallowed",
        arguments: {
          info: JSON.stringify(error.info ?? {}),
        },
      });
      break;
    }
    case "WebHookDeliveryTimeout": {
      errors.push({
        messageID: "errors.webhook.timeout",
      });
      break;
    }
    case "WebHookInvalidResponse": {
      errors.push({
        messageID: "errors.webhook.invalid-response",
      });
      break;
    }
    case "UnsupportedImageFile": {
      errors.push({
        messageID: "errors.unsupported-image-file",
        arguments: {
          type: error.info.type,
        },
      });
      break;
    }
    case "AlchemyProtocol": {
      errors.push({
        messageID: "errors.alchemy-protocol",
      });
      break;
    }
    case "DenoCheckError": {
      errors.push({
        messageID: "errors.deno-hook.typecheck",
        arguments: {
          message: error.message ?? "",
        },
      });
      break;
    }
    default:
      errors.push({
        messageID: "errors.unknown",
        arguments: { message: error.reason },
      });
  }
  return errors;
}

export interface ErrorParseRule {
  (apiError: APIError): ParsedAPIError[];
}

export function makeLocalErrorParseRule(
  sentinel: APIError,
  error: ParsedAPIError
): ErrorParseRule {
  return (apiError: APIError): ParsedAPIError[] => {
    if (apiError === sentinel) {
      return [error];
    }
    return [];
  };
}

export function makeReasonErrorParseRule(
  reason: APIError["reason"],
  errorMessageID: string
): ErrorParseRule {
  return (apiError: APIError): ParsedAPIError[] => {
    if (apiError.reason === reason) {
      return [{ messageID: errorMessageID }];
    }
    return [];
  };
}

export function makeValidationErrorMatchUnknownKindParseRule(
  kind: string,
  locationRegExp: RegExp,
  errorMessageID: string,
  values?: Values
): ErrorParseRule {
  return (apiError: APIError): ParsedAPIError[] => {
    if (apiError.reason === "ValidationFailed") {
      for (const cause of apiError.info.causes) {
        if (kind === cause.kind && locationRegExp.test(cause.location)) {
          return [{ messageID: errorMessageID, arguments: values }];
        }
      }
    }
    return [];
  };
}

export function makeInvariantViolatedErrorParseRule(
  kind: string,
  errorMessageID: string
): ErrorParseRule {
  return (apiError: APIError): ParsedAPIError[] => {
    if (apiError.reason === "InvariantViolated") {
      if (apiError.info.cause.kind === kind) {
        return [{ messageID: errorMessageID }];
      }
    }
    return [];
  };
}

function matchRule(rule: ErrorParseRule, error: APIError): ParsedAPIError[] {
  return rule(error);
}

function matchField(
  cause: ValidationFailedErrorInfoCause,
  field: FormField
): boolean {
  if (!isKnownCauseKind(cause.kind)) {
    return false;
  }

  if (cause.kind === "required") {
    for (const child of cause.details.missing) {
      const condidate = parentChildToJSONPointer(cause.location, child);
      const matched = matchParentChild(
        condidate,
        field.parentJSONPointer,
        field.fieldName
      );
      if (matched) {
        return true;
      }
    }
  }

  return matchParentChild(
    cause.location,
    field.parentJSONPointer,
    field.fieldName
  );
}

interface FieldRule {
  field: FormField | null;
  rule: ErrorParseRule;
}

function aggregateRules(
  fields: FormField[],
  topRules: ErrorParseRule[]
): FieldRule[] {
  const rules: FieldRule[] = [];
  for (const field of fields) {
    for (const rule of field.rules ?? []) {
      rules.push({ field, rule });
    }
  }
  for (const rule of topRules) {
    rules.push({ field: null, rule });
  }
  return rules;
}

function parseValidationErrors(
  errors: APIError[],
  fields: FormField[]
): {
  rawFieldCauses: Map<FormField, ValidationFailedErrorInfoCause[]>;
  unhandledErrors: APIError[];
} {
  const unhandledErrors: APIError[] = [];
  const rawFieldCauses = new Map<FormField, ValidationFailedErrorInfoCause[]>();

  for (const error of errors) {
    const unhandledCauses: ValidationFailedErrorInfoCause[] = [];

    if (error.reason === "ValidationFailed") {
      const causes = error.info.causes;
      for (const cause of causes) {
        const matchedFields = fields.filter((field) =>
          matchField(cause, field)
        );

        if (matchedFields.length === 0) {
          unhandledCauses.push(cause);
        } else {
          for (const field of matchedFields) {
            const value = rawFieldCauses.get(field);
            if (value == null) {
              rawFieldCauses.set(field, [cause]);
            } else {
              value.push(cause);
            }
          }
        }
      }
    } else {
      unhandledErrors.push(error);
    }

    if (unhandledCauses.length !== 0) {
      unhandledErrors.push({
        reason: "ValidationFailed",
        errorName: error.errorName,
        info: { causes: unhandledCauses },
      });
    }
  }

  return { rawFieldCauses, unhandledErrors };
}

function parseErrorWithRules(
  errors: APIError[],
  rules: FieldRule[]
): {
  fieldErrors: Map<FormField, ParsedAPIError[]>;
  topErrors: ParsedAPIError[];
  unhandledErrors: APIError[];
} {
  const fieldErrors: Map<FormField, ParsedAPIError[]> = new Map();
  const topErrors: ParsedAPIError[] = [];
  const unhandledErrors: APIError[] = [];

  for (const error of errors) {
    let handled = false;

    for (const { field, rule } of rules) {
      const matchedErrors = matchRule(rule, error);
      if (matchedErrors.length > 0) {
        handled = true;
        if (field != null) {
          const value = fieldErrors.get(field);
          if (value == null) {
            fieldErrors.set(field, matchedErrors);
          } else {
            value.push(...matchedErrors);
          }
        } else {
          topErrors.push(...matchedErrors);
        }
      }
    }

    if (!handled) {
      unhandledErrors.push(error);
    }
  }

  return {
    fieldErrors,
    topErrors,
    unhandledErrors,
  };
}

export interface ErrorParseResult {
  fieldErrors: Map<FormField, ParsedAPIError[]>;
  topErrors: ParsedAPIError[];
}

export function parseAPIErrors(
  errors: APIError[],
  fields: FormField[],
  topRules: ErrorParseRule[],
  fallbackMessageID?: string
): ErrorParseResult {
  if (errors.length === 0) {
    return { fieldErrors: new Map(), topErrors: [] };
  }

  const rules = aggregateRules(fields, topRules);

  const { rawFieldCauses, unhandledErrors: unhandledErrorsForRules } =
    parseValidationErrors(errors, fields);

  const { fieldErrors, topErrors, unhandledErrors } = parseErrorWithRules(
    unhandledErrorsForRules,
    rules
  );

  // Add rawFieldCauses to fieldErrors
  for (const [field, causes] of rawFieldCauses.entries()) {
    const errors = fieldErrors.get(field);
    if (errors == null) {
      fieldErrors.set(field, causes.map(parseCause));
    } else {
      errors.push(...causes.map(parseCause));
    }
  }

  // Handle fallbackMessageID
  if (unhandledErrors.length > 0) {
    if (fallbackMessageID != null) {
      topErrors.push({
        messageID: fallbackMessageID,
      });
    } else {
      for (const error of unhandledErrors) {
        topErrors.push(...parseError(error));
      }
    }
  }

  return { fieldErrors, topErrors };
}
