import { createContext, useContext } from "react";
import { ValidationFailedErrorInfoCause } from "./validation";

export type FormErrorCauses = Partial<
  Record<string, ValidationFailedErrorInfoCause[]>
>;
export interface FormContextValue {
  registerField: (
    jsonPointer: RegExp | string,
    parentJSONPointer: RegExp | string,
    fieldName: string
  ) => void;
  causes: FormErrorCauses;
}

export const FormContext = createContext<FormContextValue | null>(null);

export function useFormContext(): FormContextValue {
  const value = useContext(FormContext);
  if (value == null) {
    throw new Error("Must be used within FormContext");
  }
  return value;
}
