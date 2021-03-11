import React, { useCallback } from "react";
import { DefaultEffects, ICommandBarItemProps, Stack } from "@fluentui/react";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { APIError } from "../../error/error";
import { FormProvider } from "../../form";
import { ErrorParseRule } from "../../error/parse";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import ButtonWithLoading from "../../ButtonWithLoading";
import styles from "./OnboardingFormContainer.module.scss";

export interface FormModel {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  canSave?: boolean;
  reset: () => void;
  save: () => void;
}

export interface SaveButtonProps {
  labelId: string;
}

export interface OnboardingFormContainerProps {
  form: FormModel;
  canSave?: boolean;
  saveButtonProps?: SaveButtonProps;
  localError?: APIError | null;
  errorRules?: ErrorParseRule[];
  fallbackErrorMessageID?: string;
  farItems?: ICommandBarItemProps[];
  messageBar?: React.ReactNode;
}

const OnboardingFormContainer: React.FC<OnboardingFormContainerProps> = function OnboardingFormContainer(
  props
) {
  const {
    updateError,
    isDirty,
    isUpdating,
    save,
    canSave: formCanSave,
  } = props.form;
  const {
    canSave = true,
    saveButtonProps = { labelId: "save" },
    localError,
    errorRules,
    fallbackErrorMessageID,
    messageBar,
  } = props;

  const onFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      save();
    },
    [save]
  );

  const onSaveButtonClick = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      save();
    },
    [save]
  );
  const allowSave = formCanSave !== undefined ? formCanSave : isDirty;
  const disabled = isUpdating || !allowSave;

  return (
    <FormProvider
      error={updateError ?? localError}
      rules={errorRules}
      fallbackErrorMessageID={fallbackErrorMessageID}
    >
      <div className={styles.formContainer}>
        <div className={styles.messageBarWrapper}>
          <FormErrorMessageBar>{messageBar}</FormErrorMessageBar>
        </div>
        <div className={styles.sectionWrapper}>
          <div
            className={styles.section}
            style={{ boxShadow: DefaultEffects.elevation4 }}
          >
            <form onSubmit={onFormSubmit}>{props.children}</form>
            <Stack
              horizontal={true}
              tokens={{ childrenGap: 10 }}
              horizontalAlign="end"
            >
              <ButtonWithLoading
                type="submit"
                disabled={disabled || !canSave}
                loading={isUpdating}
                labelId={saveButtonProps.labelId}
                onClick={onSaveButtonClick}
              />
            </Stack>
            <NavigationBlockerDialog blockNavigation={isDirty} />
          </div>
        </div>
      </div>
    </FormProvider>
  );
};

export default OnboardingFormContainer;
