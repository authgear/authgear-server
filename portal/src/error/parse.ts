import { ApolloError } from "@apollo/client";
import { APIError, isAPIError } from "./error";
import { ValidationFailedErrorInfoCause } from "./validation";
import { Values } from "@oursky/react-messageformat";
import { APIPasswordPolicyViolatedError } from "./password";
import { APIResourceTooLargeError } from "./resources";

export function parseRawError(error: unknown): APIError[] {
  const errors: APIError[] = [];
  if (error instanceof ApolloError) {
    if (error.networkError) {
      errors.push({ reason: "NetworkFailed", errorName: "NetworkFailed" });
    }
    for (const e of error.graphQLErrors) {
      if (isAPIError(e.extensions)) {
        errors.push(e.extensions);
      } else {
        errors.push({
          reason: "Unknown",
          errorName: "Unknown",
          info: { message: e.message },
        });
      }
    }
  } else if (error instanceof Error) {
    errors.push({
      reason: "Unknown",
      errorName: "Unknown",
      info: { message: error.message },
    });
  } else if (typeof error === "object" && isAPIError(error)) {
    errors.push(error);
  } else if (error) {
    errors.push({
      reason: "Unknown",
      errorName: "Unknown",
      info: { message: String(error) },
    });
  }
  return errors;
}

export interface ParsedAPIError {
  message?: string;
  messageID?: string;
  arguments?: Values;
}

const errorCauseMessageIDs = {
  required: "errors.validation.required",
  general: "errors.validation.general",
  format: "errors.validation.format",
  minItems: "errors.validation.minItems",
  minimum: "errors.validation.minimum",
  maximum: "errors.validation.maximum",
  minLength: "errors.validation.minLength",
  maxLength: "errors.validation.maxLength",
  blocked: "errors.validation.blocked",
  noPrimaryAuthenticator: "errors.validation.noPrimaryAuthenticator",
};

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
  let args: Values = cause.details;
  if (cause.kind === "required") {
    args = { fieldName: cause.details.missing.join(", ") };
  }
  return { messageID, arguments: args };
}

function parsePasswordPolicyViolatedError(
  error: APIPasswordPolicyViolatedError,
  errors: ParsedAPIError[]
) {
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
}

function parseResourceTooLargeError(
  error: APIResourceTooLargeError,
  errors: ParsedAPIError[]
) {
  const faviconRe = /^static\/(.+)\/favicon\.(.*)/;
  const appLogoRe = /^static\/(.+)\/app_logo\.(.*)/;

  if (error.info.path) {
    if (faviconRe.test(error.info.path)) {
      errors.push({
        messageID: "errors.resource-too-large",
        arguments: {
          maxSize: error.info.max_size / 1024,
          resourceType: "favicon",
        },
      });
    } else if (appLogoRe.test(error.info.path)) {
      errors.push({
        messageID: "errors.resource-too-large",
        arguments: {
          maxSize: error.info.max_size / 1024,
          resourceType: "app logo",
        },
      });
    } else {
      errors.push({
        messageID: "errors.resource-too-large",
        arguments: {
          maxSize: error.info.max_size / 1024,
          resourceType: "unknown",
        },
      });
    }
  }
}

function parseError(error: APIError): ParsedAPIError[] {
  const errors: ParsedAPIError[] = [];
  switch (error.reason) {
    case "ValidationFailed":
      errors.push(...error.info.causes.map((c) => parseCause(c)));
      break;
    case "NetworkFailed":
      errors.push({ messageID: "errors.network" });
      break;
    case "Unknown":
      errors.push({
        messageID: "errors.unknown",
        arguments: { message: error.info.message },
      });
      break;
    case "ResourceTooLarge": {
      parseResourceTooLargeError(error, errors);
      break;
    }
    case "PasswordPolicyViolated": {
      parsePasswordPolicyViolatedError(error, errors);
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

export function renderErrors(
  field: FormField | null,
  errors: readonly ParsedAPIError[],
  renderMessage: (id: string, args: Values) => string
): string | undefined {
  if (errors.length === 0) {
    return undefined;
  }
  return errors.map((err) => renderError(field, err, renderMessage)).join("\n");
}

export function renderError(
  field: FormField | null,
  error: ParsedAPIError,
  renderMessage: (id: string, args: Values) => string
): string {
  if (error.messageID) {
    const args: Values = { ...error.arguments };
    if (error.messageID === errorCauseMessageIDs["required"]) {
      if (field?.fieldNameMessageID) {
        args.fieldName = renderMessage(field.fieldNameMessageID, {});
      } else if (field?.fieldName) {
        args.fieldName = field.fieldName;
      }
    }
    return renderMessage(error.messageID, args);
  }
  return error.message ?? "";
}

interface BaseErrorParseRule {
  reason: string;
  errorMessageID: string;
}

interface ValidationErrorParseRule extends BaseErrorParseRule {
  reason: "ValidationFailed";
  kind: ValidationFailedErrorInfoCause["kind"];
  location: string;
  errorMessageID: string;
}

interface InvariantViolationErrorParseRule extends BaseErrorParseRule {
  reason: "InvariantViolated";
  kind: string;
  errorMessageID: string;
}

type TypedErrorParseRules =
  | ValidationErrorParseRule
  | InvariantViolationErrorParseRule;

interface GenericErrorParseRule extends BaseErrorParseRule {
  reason: Exclude<APIError["reason"], TypedErrorParseRules["reason"]>;
  errorMessageID: string;
}

export type ErrorParseRule = TypedErrorParseRules | GenericErrorParseRule;

function matchRule(rule: ErrorParseRule, error: APIError): ParsedAPIError[] {
  if (rule.reason !== error.reason) {
    return [];
  }

  const parsedErrors: ParsedAPIError[] = [];
  switch (error.reason) {
    case "ValidationFailed": {
      const { kind, location } = rule as ValidationErrorParseRule;
      for (const cause of error.info.causes) {
        if (matchPattern(location, cause.location) && kind === cause.kind) {
          parsedErrors.push({ messageID: rule.errorMessageID });
        }
      }
      break;
    }
    case "InvariantViolated": {
      const { kind } = rule as InvariantViolationErrorParseRule;
      if (kind === error.info.cause.kind) {
        parsedErrors.push({ messageID: rule.errorMessageID });
      }
      break;
    }
    default:
      parsedErrors.push({ messageID: rule.errorMessageID });
      break;
  }
  return parsedErrors;
}

export interface FormField {
  parentJSONPointer: string;
  fieldName: string;
  fieldNameMessageID?: string;
  rules?: ErrorParseRule[];
}

export function parseJSONPointer(pointer: string): [string, string] | null {
  const locMatch = /^(.+)?\/([^/]+)?$/.exec(pointer);
  if (!locMatch) {
    return ["", ""];
  }
  const [, parentJSONPointer = "/", fieldName = ""] = locMatch;
  return [parentJSONPointer, fieldName];
}

function matchPattern(pattern: string, value: string) {
  return new RegExp(`^${pattern}$`).test(value);
}

function matchField(
  cause: ValidationFailedErrorInfoCause,
  parentJSONPointer: string,
  fieldName: string,
  field: FormField
) {
  if (cause.kind === "required") {
    parentJSONPointer = `${parentJSONPointer}/${fieldName}`;
    return (
      matchPattern(field.parentJSONPointer, parentJSONPointer) &&
      cause.details.missing.includes(field.fieldName)
    );
  }
  return (
    matchPattern(field.parentJSONPointer, parentJSONPointer) &&
    field.fieldName === fieldName
  );
}

function matchFields(
  cause: ValidationFailedErrorInfoCause,
  fields: FormField[],
  result: Map<string, ValidationFailedErrorInfoCause[]>
) {
  const locMatch = parseJSONPointer(cause.location);
  if (!locMatch) {
    return false;
  }
  const [parentJSONPointer, fieldName] = locMatch;

  let matched = false;
  for (const field of fields) {
    if (!matchField(cause, parentJSONPointer, fieldName, field)) {
      continue;
    }
    const jsonPointer = `${field.parentJSONPointer}/${field.fieldName}`;
    const matches = result.get(jsonPointer) ?? [];
    matches.push(cause);
    result.set(jsonPointer, matches);
    matched = true;
  }

  return matched;
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

function parseValidationErrors(errors: APIError[], fields: FormField[]) {
  const rawFieldCauses = new Map<string, ValidationFailedErrorInfoCause[]>();
  const unhandledErrors: APIError[] = [];
  for (const error of errors) {
    if (error.reason === "ValidationFailed") {
      const causes = error.info.causes;
      const unhandledCauses: ValidationFailedErrorInfoCause[] = [];
      for (const cause of causes) {
        if (matchFields(cause, fields, rawFieldCauses)) {
          continue;
        }
        unhandledCauses.push(cause);
      }
      if (unhandledCauses.length !== 0) {
        unhandledErrors.push({
          reason: "ValidationFailed",
          errorName: error.errorName,
          info: { causes: unhandledCauses },
        });
      }
    } else {
      unhandledErrors.push(error);
    }
  }

  return { rawFieldCauses, unhandledErrors };
}

function parseErrorWithRules(
  rawFieldCauses: Map<string, ValidationFailedErrorInfoCause[]>,
  errors: APIError[],
  rules: FieldRule[],
  fallbackMessageID?: string
) {
  const fieldErrors = new Map(
    Array.from(rawFieldCauses.entries()).map(([field, causes]) => [
      field,
      causes.map((cause) => parseCause(cause)),
    ])
  );
  const topErrors: ParsedAPIError[] = [];

  let hasFallback = false;
  for (const error of errors) {
    let matched = false;
    for (const { field, rule } of rules) {
      const matchedErrors = matchRule(rule, error);
      if (matchedErrors.length === 0) {
        continue;
      }

      if (field) {
        const jsonPointer = `${field.parentJSONPointer}/${field.fieldName}`;
        const errors = fieldErrors.get(jsonPointer) ?? [];
        errors.push(...matchedErrors);
        fieldErrors.set(jsonPointer, errors);
      } else {
        topErrors.push(...matchedErrors);
      }
      matched = true;
      break;
    }
    if (!matched) {
      if (fallbackMessageID) {
        if (!hasFallback) {
          topErrors.push({ messageID: fallbackMessageID });
          hasFallback = true;
        }
      } else {
        topErrors.push(...parseError(error));
      }
    }
  }

  return { fieldErrors, topErrors };
}

export interface ErrorParseResult {
  fieldErrors: Map<string, ParsedAPIError[]>;
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

  const { rawFieldCauses, unhandledErrors } = parseValidationErrors(
    errors,
    fields
  );

  const { fieldErrors, topErrors } = parseErrorWithRules(
    rawFieldCauses,
    unhandledErrors,
    rules,
    fallbackMessageID
  );

  return { fieldErrors, topErrors };
}
