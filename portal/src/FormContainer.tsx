import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  ICommandBarItemProps,
  PrimaryButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useValidationError } from "./error/useValidationError";
import { FormContext } from "./error/FormContext";
import ShowUnhandledValidationErrorCause from "./error/ShowUnhandledValidationErrorCauses";
import { useSystemConfig } from "./context/SystemConfigContext";
import ShowError from "./ShowError";
import NavigationBlockerDialog from "./NavigationBlockerDialog";
import CommandBarContainer from "./CommandBarContainer";

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

  const messageBar = (
    <>
      <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
      {(unhandledCauses ?? []).length === 0 && otherError && (
        <ShowError error={otherError} />
      )}
    </>
  );

  return (
    <FormContext.Provider value={formContextValue}>
      <CommandBarContainer
        isLoading={isUpdating}
        items={commandBarItems}
        messageBar={messageBar}
      >
        <form onSubmit={onFormSubmit}>{props.children}</form>
      </CommandBarContainer>
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
