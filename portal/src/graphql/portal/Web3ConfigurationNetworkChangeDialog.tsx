import React, { useContext, useMemo } from "react";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

interface Web3ConfigurationNetworkChangeDialogProps {
  isVisible: boolean;
  onConfirm: () => void;
  onDismiss: () => void;
}

const Web3ConfigurationNetworkChangeDialog: React.VFC<Web3ConfigurationNetworkChangeDialogProps> =
  function Web3ConfigurationNetworkChangeDialog(props) {
    const { isVisible, onConfirm, onDismiss } = props;

    const { renderToString } = useContext(Context);

    const { themes } = useSystemConfig();

    const dialogContentProps = useMemo(() => {
      return {
        title: (
          <FormattedMessage id="Web3ConfigurationScreen.network-change-dialog.title" />
        ),
        subText: renderToString(
          "Web3ConfigurationScreen.network-change-dialog.description"
        ),
      };
    }, [renderToString]);

    return (
      <Dialog
        hidden={!isVisible}
        dialogContentProps={dialogContentProps}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={onConfirm}
            theme={themes.destructive}
            text={<FormattedMessage id="confirm" />}
          />
          <DefaultButton
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };

export default Web3ConfigurationNetworkChangeDialog;
