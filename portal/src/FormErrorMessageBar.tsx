import React from "react";
import { useFormTopErrors } from "./form";
import {
  ErrorMessageBar,
  ErrorMessageBarContextProvider,
} from "./ErrorMessageBar";

export interface FormErrorMessageBarProps {
  children?: React.ReactNode;
}

export const FormErrorMessageBar: React.VFC<FormErrorMessageBarProps> = (
  props: FormErrorMessageBarProps
) => {
  const errors = useFormTopErrors();

  return (
    <ErrorMessageBarContextProvider errors={errors}>
      <ErrorMessageBar>{props.children}</ErrorMessageBar>
    </ErrorMessageBarContextProvider>
  );
};
