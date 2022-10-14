import React, { useContext, useMemo } from "react";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import PrimaryButton from "../../PrimaryButton";
import ExternalLink from "../../ExternalLink";

const DIALOG_MAX_WIDTH = "33%";

interface Web3ConfigurationLargeCollectionDialogProps {
  isVisible: boolean;
  onDismiss: () => void;
}
const Web3ConfigurationLargeCollectionDialog: React.VFC<
  Web3ConfigurationLargeCollectionDialogProps
> = (props) => {
  const { isVisible, onDismiss } = props;
  const { renderToString } = useContext(Context);
  const dialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="Web3ConfigurationScreen.collection-list.add-collection.large-collection-dialog.title" />
      ),
      subText: renderToString(
        "Web3ConfigurationScreen.collection-list.add-collection.large-collection-dialog.description"
      ),
    };
  }, [renderToString]);
  return (
    <Dialog
      hidden={!isVisible}
      dialogContentProps={dialogContentProps}
      onDismiss={onDismiss}
      maxWidth={DIALOG_MAX_WIDTH}
    >
      <DialogFooter>
        <ExternalLink href="https://www.authgear.com/talk-with-us?utm_source=portal&utm_medium=link">
          <PrimaryButton
            text={
              <FormattedMessage id="Web3ConfigurationScreen.collection-list.add-collection.large-collection-dialog.contact-sales" />
            }
          />
        </ExternalLink>
        <DefaultButton
          onClick={onDismiss}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
    </Dialog>
  );
};

export default Web3ConfigurationLargeCollectionDialog;
