import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  ErrorParseRule,
  parseAPIErrors,
  ParsedAPIError,
  parseRawError,
} from "./error/parse";

export interface FormField {
  parentJSONPointer: string;
  fieldName: string;
  rules?: ErrorParseRule[];
}

export interface FormContext {
  readonly fieldErrors: ReadonlyMap<string, ParsedAPIError[]>;
  readonly topErrors: readonly ParsedAPIError[];

  registerField(field: FormField): void;

  unregisterField(field: FormField): void;
}

const context = React.createContext<FormContext | null>(null);

export interface FormProviderProps {
  error?: unknown;
  fallbackErrorMessageID?: string;
  rules?: ErrorParseRule[];
}

export const FormProvider: React.FC<FormProviderProps> = (props) => {
  const { error, fallbackErrorMessageID, rules = [], children } = props;

  const [fields, setFields] = useState<FormField[]>([]);
  const registerField = useCallback((field: FormField) => {
    setFields((fields) => [...fields, field]);
  }, []);
  const unregisterField = useCallback((field: FormField) => {
    setFields((fields) => fields.filter((f) => f !== field));
  }, []);

  interface ErrorContext {
    fields: FormField[];
    topRules: ErrorParseRule[];
    fallbackErrorMessageID?: string;
  }

  const errorContextRef = useRef<ErrorContext>({ fields: [], topRules: [] });
  useEffect(() => {
    errorContextRef.current = {
      fields,
      topRules: rules,
      fallbackErrorMessageID,
    };
  }, [fields, rules, fallbackErrorMessageID]);

  const { fieldErrors, topErrors } = useMemo(() => {
    const apiErrors = parseRawError(error);
    const {
      fields,
      topRules,
      fallbackErrorMessageID,
    } = errorContextRef.current;
    const { fieldErrors, topErrors } = parseAPIErrors(
      apiErrors,
      fields,
      topRules,
      fallbackErrorMessageID
    );
    return {
      fieldErrors,
      topErrors,
    };
  }, [error]);

  const value = useMemo(
    () => ({
      fieldErrors,
      topErrors,
      registerField,
      unregisterField,
    }),
    [fieldErrors, topErrors, registerField, unregisterField]
  );

  return <context.Provider value={value}>{children}</context.Provider>;
};

export function useFormField(
  field: FormField
): { errors: readonly ParsedAPIError[] } {
  const ctx = useContext(context);
  if (!ctx) {
    throw new Error("Attempted to use useFormField outside FormProvider");
  }
  // eslint-disable-next-line @typescript-eslint/unbound-method
  const { registerField, unregisterField, fieldErrors } = ctx;

  useEffect(() => {
    registerField(field);
    return () => unregisterField(field);
  }, [registerField, unregisterField, field]);

  const jsonPointer = `${field.parentJSONPointer}/${field.fieldName}`;
  return {
    errors: fieldErrors.get(jsonPointer) ?? [],
  };
}

export function useFormTopErrors(): readonly ParsedAPIError[] {
  const ctx = useContext(context);
  if (!ctx) {
    throw new Error("Attempted to use useFormField outside FormProvider");
  }

  return ctx.topErrors;
}
