import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  FormField,
  ErrorParseRule,
  parseAPIErrors,
  ParsedAPIError,
  parseRawError,
} from "./error/parse";
import { APIResourceUpdateConflictError } from "./error/resourceUpdateConflict";

export interface FormContext {
  loading: boolean;
  readonly fieldErrors: ReadonlyMap<FormField, ParsedAPIError[]>;
  readonly topErrors: readonly ParsedAPIError[];
  readonly conflictErrors: readonly APIResourceUpdateConflictError[];
  registerField(field: FormField): void;
  unregisterField(field: FormField): void;
}

const context = createContext<FormContext | null>(null);

export interface FormProviderProps {
  loading: boolean;
  error?: unknown;
  fallbackErrorMessageID?: string;
  rules?: ErrorParseRule[];
  children?: React.ReactNode;
}

interface ErrorContext {
  fields: FormField[];
  topRules: ErrorParseRule[];
  fallbackErrorMessageID?: string;
}

export const FormProvider: React.VFC<FormProviderProps> = (props) => {
  const {
    loading,
    error,
    fallbackErrorMessageID,
    rules = [],
    children,
  } = props;

  const [fields, setFields] = useState<FormField[]>([]);
  const registerField = useCallback((field: FormField) => {
    setFields((fields) => [...fields, field]);
  }, []);
  const unregisterField = useCallback((field: FormField) => {
    setFields((fields) => fields.filter((f) => f !== field));
  }, []);

  const errorContextRef = useRef<ErrorContext>({ fields: [], topRules: [] });
  useEffect(() => {
    errorContextRef.current = {
      fields,
      topRules: rules,
      fallbackErrorMessageID,
    };
  }, [fields, rules, fallbackErrorMessageID]);

  const { fieldErrors, topErrors, conflictErrors } = useMemo(() => {
    const apiErrors = parseRawError(error);
    const { fields, topRules, fallbackErrorMessageID } =
      errorContextRef.current;
    const { fieldErrors, topErrors, conflictErrors } = parseAPIErrors(
      apiErrors,
      fields,
      topRules,
      fallbackErrorMessageID
    );
    return {
      fieldErrors,
      topErrors,
      conflictErrors,
    };
  }, [error]);

  const value = useMemo(
    () => ({
      loading,
      fieldErrors,
      topErrors,
      conflictErrors,
      registerField,
      unregisterField,
    }),
    [
      loading,
      fieldErrors,
      topErrors,
      conflictErrors,
      registerField,
      unregisterField,
    ]
  );

  return <context.Provider value={value}>{children}</context.Provider>;
};

function equal(a: FormField, b: FormField): boolean {
  if (a.fieldName !== b.fieldName) {
    return false;
  }

  if (
    typeof a.parentJSONPointer === "string" &&
    typeof b.parentJSONPointer === "string"
  ) {
    return a.parentJSONPointer === b.parentJSONPointer;
  }
  if (
    a.parentJSONPointer instanceof RegExp &&
    b.parentJSONPointer instanceof RegExp
  ) {
    return a.parentJSONPointer.source === b.parentJSONPointer.source;
  }

  return false;
}

function getFieldErrors(
  fieldErrors: ReadonlyMap<FormField, ParsedAPIError[]>,
  field: FormField
): ParsedAPIError[] {
  const errors = [];
  for (const [key, value] of fieldErrors.entries()) {
    if (equal(key, field)) {
      errors.push(...value);
    }
  }
  return errors;
}

export interface FormFieldProps {
  loading: boolean;
  errors: readonly ParsedAPIError[];
}

export function useFormField(field: FormField): FormFieldProps {
  const ctx = useContext(context);
  if (!ctx) {
    throw new Error("Attempted to use useFormField outside FormProvider");
  }
  // eslint-disable-next-line @typescript-eslint/unbound-method
  const { loading, registerField, unregisterField, fieldErrors } = ctx;

  useEffect(() => {
    registerField(field);
    return () => unregisterField(field);
  }, [registerField, unregisterField, field]);

  // We cannot simply use get to retrieve errors because FormField is not a value type.
  const errors = useMemo(
    () => getFieldErrors(fieldErrors, field),
    [field, fieldErrors]
  );

  return {
    loading,
    errors,
  };
}

export function useFormTopErrors(): readonly ParsedAPIError[] {
  const ctx = useContext(context);
  if (!ctx) {
    throw new Error("Attempted to use useFormField outside FormProvider");
  }

  return ctx.topErrors;
}

export function useFormConflictErrors(): readonly APIResourceUpdateConflictError[] {
  const ctx = useContext(context);
  if (!ctx) {
    throw new Error("Attempted to use useFormField outside FormProvider");
  }

  return ctx.conflictErrors;
}
