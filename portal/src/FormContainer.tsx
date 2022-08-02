import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  DefaultButton,
  Dialog,
  DialogFooter,
  ICommandBarItemProps,
  PrimaryButton,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import NavigationBlockerDialog from "./NavigationBlockerDialog";
import CommandBarContainer from "./CommandBarContainer";
import { FormProvider } from "./form";
import { ErrorParseRule } from "./error/parse";
import { FormErrorMessageBar } from "./FormErrorMessageBar";

export interface FormModel {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  canSave?: boolean;
  reset: () => void;
  save: () => Promise<void>;
}

export interface SaveButtonProps {
  labelId: string;
  iconName: string;
}

export interface FormContainerProps {
  form: FormModel;
  canSave?: boolean;
  saveButtonProps?: SaveButtonProps;
  localError?: unknown;
  errorRules?: ErrorParseRule[];
  fallbackErrorMessageID?: string;
  messageBar?: React.ReactNode;
  primaryItems?: ICommandBarItemProps[];
  secondaryItems?: ICommandBarItemProps[];
  afterSave?: () => void;
}

const FormContainer: React.FC<FormContainerProps> = function FormContainer(
  props
) {
  const {
    updateError,
    isDirty,
    isUpdating,
    reset,
    save,
    canSave: formCanSave,
  } = props.form;
  const {
    canSave = true,
    saveButtonProps = { labelId: "save", iconName: "Save" },
    localError,
    errorRules,
    fallbackErrorMessageID,
    primaryItems,
    secondaryItems,
    messageBar,
    afterSave,
  } = props;

  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const callSave = useCallback(() => {
    save().then(
      () => afterSave?.(),
      () => {}
    );
  }, [save, afterSave]);

  const onFormSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      callSave();
    },
    [callSave]
  );

  const [isResetDialogVisible, setIsResetDialogVisible] = useState(false);
  const onDismissResetDialog = useCallback(() => {
    setIsResetDialogVisible(false);
  }, []);
  const doReset = useCallback(() => {
    reset();
    // If the form contains a CodeEditor, dialog dismiss animation does not play.
    // Defer the dismissal to ensure dismiss animation.
    setTimeout(() => setIsResetDialogVisible(false), 0);
  }, [reset]);

  const allowSave = formCanSave !== undefined ? formCanSave : isDirty;
  const disabled = isUpdating || !allowSave;

  const items: ICommandBarItemProps[] = useMemo(() => {
    let items: ICommandBarItemProps[] = [
      {
        key: "save",
        text: renderToString(saveButtonProps.labelId),
        iconProps: { iconName: saveButtonProps.iconName },
        disabled: disabled || !canSave,
        onClick: () => {
          callSave();
        },
      },
    ];
    if (primaryItems != null) {
      items = [...items, ...primaryItems];
    }
    return items;
  }, [
    canSave,
    disabled,
    callSave,
    saveButtonProps,
    renderToString,
    primaryItems,
  ]);

  const farItems: ICommandBarItemProps[] = useMemo(() => {
    let farItems: ICommandBarItemProps[] = [
      {
        key: "reset",
        text: renderToString("discard-changes"),
        iconProps: { iconName: "Refresh" },
        disabled,
        theme: disabled ? themes.main : themes.destructive,
        onClick: () => setIsResetDialogVisible(true),
      },
    ];
    if (secondaryItems != null) {
      farItems = [...farItems, ...secondaryItems];
    }
    return farItems;
  }, [disabled, renderToString, themes, secondaryItems]);

  const resetDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="FormContainer.reset-dialog.title" />,
      subText: renderToString("FormContainer.reset-dialog.message"),
    };
  }, [renderToString]);

  const onConfirmNavigation = useCallback(() => {
    reset();
  }, [reset]);

  return (
    <FormProvider
      loading={isUpdating}
      error={updateError ?? localError}
      rules={errorRules}
      fallbackErrorMessageID={fallbackErrorMessageID}
    >
      <CommandBarContainer
        isLoading={isUpdating}
        primaryItems={items}
        secondaryItems={farItems}
        messageBar={<FormErrorMessageBar>{messageBar}</FormErrorMessageBar>}
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
            <FormattedMessage id="FormContainer.reset-dialog.confirm" />
          </PrimaryButton>
          <DefaultButton onClick={onDismissResetDialog}>
            <FormattedMessage id="cancel" />
          </DefaultButton>
        </DialogFooter>
      </Dialog>
      <NavigationBlockerDialog
        blockNavigation={isDirty}
        onConfirmNavigation={onConfirmNavigation}
      />
    </FormProvider>
  );
};

export default FormContainer;
