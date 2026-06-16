import React, { useCallback, useContext, useState } from "react";
import { Spinner, SpinnerSize } from "@fluentui/react";
import { Context, FormattedMessage } from "./intl";
import { useSystemConfig } from "./context/SystemConfigContext";
import { FormErrorMessageBar } from "./FormErrorMessageBar";
import PrimaryButton from "./PrimaryButton";
import DefaultLayout from "./DefaultLayout";
import {
  FormContainerBase,
  FormContainerBaseProps,
  useFormContainerBaseContext,
} from "./FormContainerBase";
import ActionButton from "./ActionButton";
import styles from "./FormContainer.module.css";
import { ConfirmationDialog } from "./components/v2/ConfirmationDialog/ConfirmationDialog";

export interface SaveButtonProps {
  labelId: string;
  iconProps?: {
    iconName?: string;
  };
}

export interface FormContainerProps extends FormContainerBaseProps {
  className?: string;
  saveButtonProps?: SaveButtonProps;
  stickyFooterComponent?: boolean;
  fallbackErrorMessageID?: string;
  messageBar?: React.ReactNode;
  showDiscardButton?: boolean;
  hideFooterComponent?: boolean;
}

const FormContainer_: React.VFC<FormContainerProps> = function FormContainer_(
  props
) {
  const {
    saveButtonProps,
    messageBar,
    hideFooterComponent,
    showDiscardButton = false,
    stickyFooterComponent = false,
  } = props;

  const { canReset, onReset, onSubmit } = useFormContainerBaseContext();
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const [isResetDialogVisible, setIsResetDialogVisible] = useState(false);
  const onDisplayResetDialog = useCallback(() => {
    setIsResetDialogVisible(true);
  }, []);
  const onDismissResetDialog = useCallback(() => {
    setIsResetDialogVisible(false);
  }, []);
  const doReset = useCallback(() => {
    onReset();
    // If the form contains a CodeEditor, dialog dismiss animation does not play.
    // Defer the dismissal to ensure dismiss animation.
    setTimeout(() => setIsResetDialogVisible(false), 0);
  }, [onReset]);

  return (
    <>
      <DefaultLayout
        footerPosition={stickyFooterComponent ? "sticky" : "end"}
        footer={
          hideFooterComponent ? null : (
            <>
              <FormSaveButton saveButtonProps={saveButtonProps} />
              {showDiscardButton ? (
                <ActionButton
                  text={renderToString("discard-changes")}
                  iconProps={{ iconName: "Refresh" }}
                  disabled={!canReset}
                  theme={!canReset ? themes.main : themes.destructive}
                  onClick={onDisplayResetDialog}
                />
              ) : null}
            </>
          )
        }
        messageBar={<FormErrorMessageBar>{messageBar}</FormErrorMessageBar>}
      >
        <form className={props.className} onSubmit={onSubmit}>
          {props.children}
        </form>
      </DefaultLayout>
      <ConfirmationDialog
        open={isResetDialogVisible}
        onOpenChange={(open) => {
          if (!open) {
            onDismissResetDialog();
          }
        }}
        title={<FormattedMessage id="FormContainer.reset-dialog.title" />}
        description={
          <FormattedMessage id="FormContainer.reset-dialog.message" />
        }
        confirmText={
          <FormattedMessage id="FormContainer.reset-dialog.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        confirmColor="red"
        onConfirm={doReset}
        onCancel={onDismissResetDialog}
      />
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

const DEFAULT_SAVE_BUTTON_PROPS = { labelId: "save" } satisfies SaveButtonProps;

export function FormSaveButton({
  saveButtonProps = DEFAULT_SAVE_BUTTON_PROPS,
}: {
  saveButtonProps?: SaveButtonProps;
}): React.ReactElement {
  const { canSave, isUpdating, onSave } = useFormContainerBaseContext();

  return (
    <PrimaryButton
      text={
        <div className={styles.saveButton}>
          {isUpdating ? (
            <Spinner size={SpinnerSize.xSmall} ariaLive="assertive" />
          ) : null}
          <span>
            <FormattedMessage id={saveButtonProps.labelId} />
          </span>
        </div>
      }
      iconProps={saveButtonProps.iconProps}
      disabled={!canSave}
      onClick={onSave}
    />
  );
}
