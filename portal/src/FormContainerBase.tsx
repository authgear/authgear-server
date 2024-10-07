import React, { createContext, useCallback, useContext, useMemo } from "react";
import { ErrorParseRule } from "./error/parse";
import { useConsumeError } from "./hook/error";
import { FormProvider } from "./form";
import NavigationBlockerDialog from "./NavigationBlockerDialog";
import FormConfirmOverridingDialog from "./FormConfirmOverridingDialog";

export interface FormModel {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  reset: () => void;
  save: (ignoreConflict?: boolean) => Promise<void>;
}

export interface FormContainerBaseProps<Form = FormModel> {
  form: Form;
  canSave?: boolean;
  localError?: unknown;
  errorRules?: ErrorParseRule[];
  fallbackErrorMessageID?: string;
  beforeSave?: () => Promise<void>;
  afterSave?: () => void;
  children?: React.ReactNode;
}

export interface FormContainerBaseValues<Form = FormModel> {
  form: Form;
  canReset: boolean;
  canSave: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  onReset: () => void;
  onSave: () => void;
  onSubmit: (e: React.FormEvent) => void;
}

const FormContainerBaseContext = createContext<
  FormContainerBaseValues | undefined
>(undefined);

export const FormContainerBase: React.VFC<FormContainerBaseProps> =
  function FormContainerBase({
    form,
    children,
    canSave: propCanSave = true,
    localError,
    errorRules,
    fallbackErrorMessageID,
    beforeSave,
    afterSave,
  }) {
    const { updateError, isDirty, isUpdating, reset, save } = form;

    const contextError = useConsumeError();

    const callSave = useCallback(
      (ignoreConflict: boolean = false) => {
        (beforeSave ?? (async () => Promise.resolve()))().then(
          () => {
            save(ignoreConflict).then(
              () => afterSave?.(),
              () => {}
            );
          },
          () => {}
        );
      },
      [beforeSave, save, afterSave]
    );

    const onFormSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        callSave();
      },
      [callSave]
    );

    const isResetDisabled = isUpdating || !isDirty;
    const isSaveDisabled = isUpdating || !isDirty || !propCanSave;

    const onConfirmNavigation = useCallback(() => {
      reset();
    }, [reset]);

    const value = useMemo<FormContainerBaseValues>(() => {
      return {
        form,
        canReset: !isResetDisabled,
        canSave: !isSaveDisabled,
        isUpdating,
        isDirty,
        onReset: reset,
        onSave: callSave,
        onSubmit: onFormSubmit,
      };
    }, [
      form,
      callSave,
      isDirty,
      isResetDisabled,
      isSaveDisabled,
      isUpdating,
      onFormSubmit,
      reset,
    ]);

    return (
      <FormContainerBaseContext.Provider value={value}>
        <FormProvider
          loading={isUpdating}
          error={contextError ?? updateError ?? localError}
          rules={errorRules}
          fallbackErrorMessageID={fallbackErrorMessageID}
        >
          {children}
          <NavigationBlockerDialog
            blockNavigation={isDirty}
            onConfirmNavigation={onConfirmNavigation}
          />
          <FormConfirmOverridingDialog save={callSave} />
        </FormProvider>
      </FormContainerBaseContext.Provider>
    );
  };

export function useFormContainerBaseContext<
  Form = FormModel
>(): FormContainerBaseValues<Form> {
  const ctx = useContext(FormContainerBaseContext);
  if (ctx === undefined) {
    throw new Error("FormContainerBaseContext is not provided");
  }
  return ctx as FormContainerBaseValues<Form>;
}
