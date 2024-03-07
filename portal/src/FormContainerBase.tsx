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

export interface FormContainerBaseProps {
  form: FormModel;
  canSave?: boolean;
  localError?: unknown;
  errorRules?: ErrorParseRule[];
  fallbackErrorMessageID?: string;
  beforeSave?: () => Promise<void>;
  afterSave?: () => void;
  children?: React.ReactNode;
}

export interface FormContainerBaseValues {
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
  function FormContainerBase(props) {
    const { updateError, isDirty, isUpdating, reset, save } = props.form;
    const {
      children,
      canSave: propCanSave = true,
      localError,
      errorRules,
      fallbackErrorMessageID,
      beforeSave = async () => Promise.resolve(),
      afterSave,
    } = props;

    const contextError = useConsumeError();

    const callSave = useCallback(
      (ignoreConflict: boolean = false) => {
        beforeSave().then(
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
        canReset: !isResetDisabled,
        canSave: !isSaveDisabled,
        isUpdating,
        isDirty,
        onReset: reset,
        onSave: callSave,
        onSubmit: onFormSubmit,
      };
    }, [
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

export function useFormContainerBaseContext(): FormContainerBaseValues {
  const ctx = useContext(FormContainerBaseContext);
  if (ctx === undefined) {
    throw new Error("FormContainerBaseContext is not provided");
  }
  return ctx;
}
