import React, { useCallback, useContext, useMemo, useState } from "react";
import { Dialog, DialogFooter, ICommandBarItemProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import CommandBarContainer from "./CommandBarContainer";
import { FormErrorMessageBar } from "./FormErrorMessageBar";
import { onRenderCommandBarPrimaryButton } from "./CommandBarPrimaryButton";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import {
  FormContainerBase,
  FormContainerBaseProps,
  useFormContainerBaseContext,
} from "./FormContainerBase";

export interface SaveButtonProps {
  labelId: string;
  iconName: string;
}

export interface FormContainerProps extends FormContainerBaseProps {
  saveButtonProps?: SaveButtonProps;
  fallbackErrorMessageID?: string;
  messageBar?: React.ReactNode;
  primaryItems?: ICommandBarItemProps[];
  secondaryItems?: ICommandBarItemProps[];
  hideCommandBar?: boolean;
  renderHeaderContent?: (
    defaultHeaderContent: React.ReactNode
  ) => React.ReactNode;
}

const FormContainer_: React.VFC<FormContainerProps> = function FormContainer_(
  props
) {
  const {
    saveButtonProps = { labelId: "save", iconName: "Save" },
    primaryItems,
    secondaryItems,
    messageBar,
    hideCommandBar,
    renderHeaderContent,
  } = props;

  const { canReset, canSave, isUpdating, onReset, onSave, onSubmit } =
    useFormContainerBaseContext();
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const [isResetDialogVisible, setIsResetDialogVisible] = useState(false);
  const onDismissResetDialog = useCallback(() => {
    setIsResetDialogVisible(false);
  }, []);
  const doReset = useCallback(() => {
    onReset();
    // If the form contains a CodeEditor, dialog dismiss animation does not play.
    // Defer the dismissal to ensure dismiss animation.
    setTimeout(() => setIsResetDialogVisible(false), 0);
  }, [onReset]);

  const items: ICommandBarItemProps[] = useMemo(() => {
    let items: ICommandBarItemProps[] = [
      {
        key: "save",
        text: renderToString(saveButtonProps.labelId),
        iconProps: { iconName: saveButtonProps.iconName },
        disabled: !canSave,
        onClick: () => {
          onSave();
        },
        onRender: onRenderCommandBarPrimaryButton,
      },
    ];
    if (primaryItems != null) {
      items = [...items, ...primaryItems];
    }
    return items;
  }, [
    renderToString,
    saveButtonProps.labelId,
    saveButtonProps.iconName,
    canSave,
    primaryItems,
    onSave,
  ]);

  const farItems: ICommandBarItemProps[] = useMemo(() => {
    let farItems: ICommandBarItemProps[] = [
      {
        key: "reset",
        text: renderToString("discard-changes"),
        iconProps: { iconName: "Refresh" },
        disabled: !canReset,
        theme: !canReset ? themes.main : themes.destructive,
        onClick: () => setIsResetDialogVisible(true),
      },
    ];
    if (secondaryItems != null) {
      farItems = [...farItems, ...secondaryItems];
    }
    return farItems;
  }, [
    renderToString,
    canReset,
    themes.main,
    themes.destructive,
    secondaryItems,
  ]);

  const resetDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="FormContainer.reset-dialog.title" />,
      subText: renderToString("FormContainer.reset-dialog.message"),
    };
  }, [renderToString]);

  return (
    <>
      <CommandBarContainer
        renderHeaderContent={renderHeaderContent}
        hideCommandBar={hideCommandBar}
        isLoading={isUpdating}
        primaryItems={items}
        secondaryItems={farItems}
        messageBar={<FormErrorMessageBar>{messageBar}</FormErrorMessageBar>}
      >
        <form onSubmit={onSubmit}>{props.children}</form>
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
    </>
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
