import React, { useCallback, useContext, useMemo, useState } from "react";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";
import { FormErrorMessageBar } from "./FormErrorMessageBar";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import DefaultLayout from "./DefaultLayout";
import {
  FormContainerBase,
  FormContainerBaseProps,
  useFormContainerBaseContext,
} from "./FormContainerBase";
import ActionButton from "./ActionButton";

export interface SaveButtonProps {
  labelId: string;
  iconProps?: {
    iconName?: string;
  };
}

interface FormContainerHeaderComponentProps {
  children?: React.ReactNode;
}

export interface FormContainerProps extends FormContainerBaseProps {
  className?: string;
  saveButtonProps?: SaveButtonProps;
  stickyFooterComponent?: boolean;
  fallbackErrorMessageID?: string;
  messageBar?: React.ReactNode;
  showDiscardButton?: boolean;
  hideFooterComponent?: boolean;
  HeaderComponent?: React.VFC<FormContainerHeaderComponentProps>;
}

const FormContainer_: React.VFC<FormContainerProps> = function FormContainer_(
  props
) {
  const {
    saveButtonProps = { labelId: "save" },
    messageBar,
    hideFooterComponent,
    showDiscardButton = false,
    stickyFooterComponent = false,
    HeaderComponent,
  } = props;

  const { canSave, isUpdating, canReset, onReset, onSave, onSubmit } =
    useFormContainerBaseContext();
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

  const resetDialogContentProps = useMemo(() => {
    return {
      title: <FormattedMessage id="FormContainer.reset-dialog.title" />,
      subText: renderToString("FormContainer.reset-dialog.message"),
    };
  }, [renderToString]);

  return (
    <>
      <DefaultLayout
        position={stickyFooterComponent ? "sticky" : "end"}
        HeaderComponent={HeaderComponent}
        footer={
          hideFooterComponent ? null : (
            <>
              <PrimaryButton
                text={renderToString(saveButtonProps.labelId)}
                iconProps={saveButtonProps.iconProps}
                disabled={!canSave}
                onClick={onSave}
              />
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
