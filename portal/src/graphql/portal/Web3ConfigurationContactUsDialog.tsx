import React, { useContext, useMemo, useCallback } from "react";
import { Dialog, DialogFooter, IDialogProps } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import PrimaryButton from "../../PrimaryButton";
import ExternalLink from "../../ExternalLink";

const DIALOG_MAX_WIDTH = "33%";

interface Web3ConfigurationContactUsDialogProps {
  isVisible: boolean;
  onDismiss: IDialogProps["onDismiss"];
}
const Web3ConfigurationContactUsDialog: React.VFC<
  Web3ConfigurationContactUsDialogProps
> = (props) => {
  const { isVisible, onDismiss } = props;
  const { renderToString } = useContext(Context);
  const dialogContentProps = useMemo(() => {
    return {
      title: renderToString("Web3ConfigurationScreen.contact-us-dialog.title"),
      subText: (
        <FormattedMessage id="Web3ConfigurationScreen.contact-us-dialog.description" />
      ),
    };
  }, [renderToString]);

  const onClickCancel = useCallback(
    (e) => {
      e?.preventDefault();
      e?.stopPropagation();
      onDismiss?.();
    },
    [onDismiss]
  );

  return (
    <Dialog
      hidden={!isVisible}
      // @ts-expect-error
      dialogContentProps={dialogContentProps}
      onDismiss={onDismiss}
      maxWidth={DIALOG_MAX_WIDTH}
    >
      <DialogFooter>
        <ExternalLink href="https://www.authgear.com/talk-with-us?utm_source=portal&utm_medium=link">
          <PrimaryButton
            text={
              <FormattedMessage id="Web3ConfigurationScreen.contact-us-dialog.action" />
            }
          />
        </ExternalLink>
        <DefaultButton
          onClick={onClickCancel}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
};

export default Web3ConfigurationContactUsDialog;
