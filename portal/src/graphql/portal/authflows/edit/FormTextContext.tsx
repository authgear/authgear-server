import React, { useContext } from "react";
import { FormText } from "./useFormText";

export interface FormTextContextValue {
  text: FormText;
  // The project's languages, the default first.
  languages: string[];
  defaultLanguage: string;
  // The language the form panel, field panel and canvas show text in.
  language: string;
  setLanguage: (language: string) => void;
  readOnly: boolean;
  // Form names in authgear.yaml with the edits applied.
  usedNames: Set<string>;
  // A name for a new form.
  newFormName: () => string;
}

export const FormTextContext = React.createContext<FormTextContextValue | null>(
  null
);

// Absent outside the auth flows editor.
export function useFormTextContext(): FormTextContextValue | null {
  return useContext(FormTextContext);
}
