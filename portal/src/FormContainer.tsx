import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  CommandBar,
  DefaultButton,
  Dialog,
  DialogFooter,
  ICommandBarItemProps,
  PrimaryButton,
  ProgressIndicator,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useValidationError } from "./error/useValidationError";
import { FormContext } from "./error/FormContext";
import ShowUnhandledValidationErrorCause from "./error/ShowUnhandledValidationErrorCauses";
import { useSystemConfig } from "./context/SystemConfigContext";
import ShowError from "./ShowError";
import styles from "./FormContainer.module.scss";
import NavigationBlockerDialog from "./NavigationBlockerDialog";

const progressIndicatorStyles = {
  itemProgress: {
    padding: 0,
  },
};

export interface FormModel {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  reset: () => void;
  save: () => void;
}

export interface FormContainerProps {
  form: FormModel;
}

const FormContainer: React.FC<FormContainerProps> = function FormContainer(
  props
) {
  const { updateError, isDirty, isUpdating, reset, save } = props.form;
  const {
    otherError,
    unhandledCauses,
    value: formContextValue,
  } = useValidationError(updateError);

  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const onFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      save();
    },
    [save]
  );

  const [isResetDialogVisible, setIsResetDialogVisible] = useState(false);
  const onDismissResetDialog = useCallback(() => {
    setIsResetDialogVisible(false);
  }, []);
  const doReset = useCallback(() => {
    reset();
    setIsResetDialogVisible(false);
  }, [reset]);

  const disabled = isUpdating || !isDirty;
  const commandBarItems: ICommandBarItemProps[] = useMemo(
    () => [
      {
        key: "save",
        text: renderToString("save"),
        iconProps: { iconName: "Save" },
        disabled,
        onClick: () => save(),
      },
      {
        key: "reset",
        text: renderToString("reset"),
        iconProps: { iconName: "Delete" },
        disabled,
        theme: disabled ? themes.main : themes.destructive,
        onClick: () => setIsResetDialogVisible(true),
      },
    ],
    [disabled, save, renderToString, themes]
  );

  const resetDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="FormContainer.reset-dialog.title" />,
      subText: renderToString("FormContainer.reset-dialog.message"),
    };
  }, [renderToString]);

  return (
    <FormContext.Provider value={formContextValue}>
      <form onSubmit={onFormSubmit}>
        <div className={styles.header}>
          <CommandBar className={styles.commandBar} items={commandBarItems} />
          {isUpdating && (
            <ProgressIndicator
              className={styles.progressBar}
              styles={progressIndicatorStyles}
            />
          )}
          <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
          {(unhandledCauses ?? []).length === 0 && otherError && (
            <ShowError error={otherError} />
          )}
        </div>
        {props.children}
      </form>
      <Dialog
        hidden={!isResetDialogVisible}
        dialogContentProps={resetDialogContentProps}
        onDismiss={onDismissResetDialog}
      >
        <DialogFooter>
          <PrimaryButton onClick={doReset} theme={themes.destructive}>
            <FormattedMessage id="reset" />
          </PrimaryButton>
          <DefaultButton onClick={onDismissResetDialog}>
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
      <NavigationBlockerDialog blockNavigation={isDirty} />
    </FormContext.Provider>
  );
};

export default FormContainer;
