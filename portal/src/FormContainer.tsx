import React, { useCallback, useContext, useMemo, useState } from "react";
import { Dialog, DialogFooter, ICommandBarItemProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import NavigationBlockerDialog from "./NavigationBlockerDialog";
import CommandBarContainer from "./CommandBarContainer";
import { FormProvider } from "./form";
import { ErrorParseRule } from "./error/parse";
import { FormErrorMessageBar } from "./FormErrorMessageBar";
import { onRenderCommandBarPrimaryButton } from "./CommandBarPrimaryButton";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import { useConsumeError } from "./hook/error";

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
  beforeSave?: () => Promise<void>;
  afterSave?: () => void;
  children?: React.ReactNode;
}

const FormContainer: React.VFC<FormContainerProps> = function FormContainer(
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
    beforeSave = async () => Promise.resolve(),
    afterSave,
  } = props;

  const contextError = useConsumeError();
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const callSave = useCallback(() => {
    beforeSave().then(
      () => {
        save().then(
          () => afterSave?.(),
          () => {}
        );
      },
      () => {}
    );
  }, [beforeSave, save, afterSave]);

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
        onRender: onRenderCommandBarPrimaryButton,
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
      error={contextError ?? updateError ?? localError}
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
          <PrimaryButton
            onClick={doReset}
            theme={themes.destructive}
            text={<FormattedMessage id="FormContainer.reset-dialog.confirm" />}
          />
          <DefaultButton
            onClick={onDismissResetDialog}
            text={<FormattedMessage id="cancel" />}
          />
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
