import React, { useCallback, useContext, useMemo } from "react";
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { getNetworkNameID } from "../../util/networkId";
import PrimaryButton from "../../PrimaryButton";
import { truncateAddress } from "../../util/hex";
import { CollectionItem } from "./Web3ConfigurationScreen";

interface Web3ConfigurationCollectionDeletionDialogProps {
  nftCollection: CollectionItem;

  isVisible: boolean;
  onConfirm: (nftCollection: CollectionItem) => void;
  onDismiss: () => void;
}

const Web3ConfigurationCollectionDeletionDialog: React.VFC<Web3ConfigurationCollectionDeletionDialogProps> =
  function Web3ConfigurationCollectionDeletionDialog(props) {
    const { nftCollection, isVisible, onConfirm, onDismiss } = props;

    const { themes } = useSystemConfig();
    const { renderToString } = useContext(Context);

    const dialogContentProps = useMemo(() => {
      const networkNameId = getNetworkNameID(nftCollection);
      return {
        title: renderToString("Web3ConfigurationScreen.deletion-dialog.title"),
        subText: renderToString(
          "Web3ConfigurationScreen.deletion-dialog.description",
          {
            collection: renderToString("NftCollection.item.identifier", {
              name: nftCollection.name,
              network: renderToString(networkNameId),
              address: truncateAddress(nftCollection.contractAddress),
            }),
          }
        ),
      };
    }, [nftCollection, renderToString]);

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
