import React from "react";
import cn from "classnames";
import { FormErrorMessageBar } from "./FormErrorMessageBar";
import DefaultLayout from "./DefaultLayout";
import {
  FormContainerBase,
  FormContainerBaseProps,
  useFormContainerBaseContext,
} from "./FormContainerBase";
import styles from "./FormContainer.module.css";

export interface FormContainerProps extends FormContainerBaseProps {
  className?: string;
  messageBar?: React.ReactNode;
}

// FormContainer wires a form model into the FormContainerBase context (which
// SaveFunctionBar and the error bar consume) and renders the children inside
// a <form>. Saving/discarding UI is the SaveFunctionBar's job; the old
// FluentUI footer is gone.
const FormContainer_: React.VFC<FormContainerProps> = function FormContainer_(
  props
) {
  const { messageBar } = props;
  const { onSubmit } = useFormContainerBaseContext();

  return (
    <DefaultLayout
      messageBar={<FormErrorMessageBar>{messageBar}</FormErrorMessageBar>}
    >
      <form className={cn(styles.form, props.className)} onSubmit={onSubmit}>
        {props.children}
      </form>
    </DefaultLayout>
  );
};

const FormContainer: React.VFC<FormContainerProps> = function FormContainer(
  props
) {
  return (
    <FormContainerBase {...props}>
      <FormContainer_ {...props} />
    </FormContainerBase>
  );
};

export default FormContainer;
