import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import { NftCollection } from "./globalTypes.generated";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { getNetworkNameId } from "../../util/networkId";
import PrimaryButton from "../../PrimaryButton";
import { truncateAddress } from "../../util/hexAddress";

interface Web3ConfigurationCollectionDeletionDialogProps {
  nftCollection: NftCollection;

  isVisible: boolean;
  onConfirm: (nftCollection: NftCollection) => void;
  onDismiss: () => void;
}

const Web3ConfigurationCollectionDeletionDialog: React.VFC<Web3ConfigurationCollectionDeletionDialogProps> =
  function Web3ConfigurationCollectionDeletionDialog(props) {
    const { nftCollection, isVisible, onConfirm, onDismiss } = props;

    const { themes } = useSystemConfig();
    const { renderToString } = useContext(Context);

    const renderCollectionItem = useCallback(
      (collection: NftCollection) => {
        const networkNameId = getNetworkNameId(collection);

        return `${collection.name} (${renderToString(
          networkNameId
        )}: ${truncateAddress(collection.contractAddress)})`;
      },
      [renderToString]
    );

    const dialogContentProps = useMemo(() => {
      return {
        title: renderToString("Web3ConfigurationScreen.deletion-dialog.title"),
        subText: renderToString(
          "Web3ConfigurationScreen.deletion-dialog.description",
          {
            collection: renderCollectionItem(nftCollection),
          }
        ),
      };
    }, [nftCollection, renderCollectionItem, renderToString]);

    const onConfirmDelete = useCallback(() => {
      onConfirm(nftCollection);
    }, [nftCollection, onConfirm]);

    return (
      <Dialog
        hidden={!isVisible}
        dialogContentProps={dialogContentProps}
        onDismiss={onDismiss}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={onConfirmDelete}
            theme={themes.destructive}
            text={
              <FormattedMessage id="Web3ConfigurationScreen.deletion-dialog.remove" />
            }
          />
          <DefaultButton
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };

export default Web3ConfigurationCollectionDeletionDialog;
