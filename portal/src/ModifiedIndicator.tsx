import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  Dialog,
  DialogFooter,
  IDialogProps,
  MessageBar,
  MessageBarType,
} from "@fluentui/react";
import { Context, FormattedMessage } from "./intl";

import styles from "./ModifiedIndicator.module.css";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import MessageBarButton from "./MessageBarButton";

export interface ModifiedIndicatorProps {
  className?: string;
  isModified: boolean;
  resetForm: () => void;
}

const MESSAGE_BAR_STYLES = {
  actions: { margin: "4px 12px" },
};

export const ModifiedIndicator: React.VFC<ModifiedIndicatorProps> =
  function ModifiedIndicator(props: ModifiedIndicatorProps) {
    const { isModified, resetForm, className } = props;
    const { renderToString } = useContext(Context);

    const [confirmDialogVisible, setConfirmDialogVisible] = useState(false);

    const onResetClicked = useCallback(() => {
      setConfirmDialogVisible(true);
    }, []);

    const dismissConfirmDialog = useCallback(() => {
      setConfirmDialogVisible(false);
    }, []);

    const onConfirmClicked = useCallback(() => {
      resetForm();
      // workaround, the animation is not smooth
      // due to the handling reset state
      window.setTimeout(() => {
        setConfirmDialogVisible(false);
      }, 0);
    }, [resetForm]);

    const confirmDialogContentProps = useMemo<
      IDialogProps["dialogContentProps"]
    >(() => {
      return {
        title: <FormattedMessage id="ModifiedIndicator.confirm-dialog.title" />,
        subText: renderToString("ModifiedIndicator.confirm-dialog.message"),
      };
    }, [renderToString]);

    const actions = useMemo(() => {
      return (
        <MessageBarButton
          className={styles.messageBarButton}
          onClick={onResetClicked}
          iconProps={{
            iconName: "Refresh",
            className: styles.messageBarButtonIcon,
          }}
          text={<FormattedMessage id="reset" />}
        />
      );
    }, [onResetClicked]);

    return (
      <>
        <Dialog
          hidden={!confirmDialogVisible}
          dialogContentProps={confirmDialogContentProps}
          onDismiss={dismissConfirmDialog}
        >
          <DialogFooter>
            <PrimaryButton
              onClick={onConfirmClicked}
              text={<FormattedMessage id="confirm" />}
            />
            <DefaultButton
              onClick={dismissConfirmDialog}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>

        {isModified ? (
          <MessageBar
            className={cn(styles.messageBar, className)}
            styles={MESSAGE_BAR_STYLES}
            messageBarType={MessageBarType.warning}
            actions={actions}
          >
            <FormattedMessage id="ModifiedIndicator.message" />
          </MessageBar>
        ) : null}
      </>
    );
  };
