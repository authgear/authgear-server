import React, { useCallback, useMemo } from "react";
import { Dialog, DialogFooter, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import { NftCollection } from "./globalTypes.generated";
import styles from "./Web3ConfigurationDetailDialog.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { getNetworkNameId } from "../../util/networkId";
import ActionButton from "../../ActionButton";

interface Web3ConfigurationDetailDialogProps {
  nftCollection: NftCollection;

  isVisible: boolean;
  onDelete: (nftCollection: NftCollection) => void;
  onDismiss: () => void;
}

const Web3ConfigurationDetailDialog: React.VFC<Web3ConfigurationDetailDialogProps> =
  function Web3ConfigurationDetailDialog(props) {
    const { nftCollection, isVisible, onDelete, onDismiss } = props;

    const { themes } = useSystemConfig();

    const dialogContentProps = useMemo(() => {
      return {
        title: nftCollection.name,
      };
    }, [nftCollection]);

    const onRemoveCollection = useCallback(() => {
      onDelete(nftCollection);
    }, [nftCollection, onDelete]);

    return (
      <Dialog
        hidden={!isVisible}
        dialogContentProps={dialogContentProps}
        onDismiss={onDismiss}
      >
        <div className={styles.contentContainer}>
          <div className={styles.fieldContainer}>
            <Text className={styles.fieldTitle} block={true}>
              <FormattedMessage id="Web3ConfigurationScreen.detail-dialog.contract-address" />
            </Text>
            <Text
              className={styles.contractAddressContainer}
              as="p"
              block={true}
            >
              {nftCollection.contractAddress}
            </Text>
            <Text>
              <FormattedMessage
                id="Web3ConfigurationScreen.detail-dialog.view-on-etherscan"
                values={{
                  address: nftCollection.contractAddress,
                }}
              />
            </Text>
          </div>

          <div className={styles.fieldContainer}>
            <Text className={styles.fieldTitle} block={true}>
              <FormattedMessage id="Web3ConfigurationScreen.detail-dialog.network" />
            </Text>
            <Text as="p" block={true}>
              <FormattedMessage id={getNetworkNameId(nftCollection)} />
            </Text>
          </div>
          <div className={styles.fieldContainer}>
            <Text className={styles.fieldTitle} block={true}>
              <FormattedMessage id="Web3ConfigurationScreen.detail-dialog.block-height" />
            </Text>
            <Text as="p" block={true}>
              {nftCollection.blockHeight}
            </Text>
          </div>

          <div className={styles.fieldContainer}>
            <Text className={styles.fieldTitle} block={true}>
              <FormattedMessage id="Web3ConfigurationScreen.detail-dialog.tokens" />
            </Text>
            <Text as="p" block={true}>
              {nftCollection.totalSupply}
            </Text>
          </div>

          <div className={styles.removeCollectionButtonContainer}>
            <ActionButton
              className={styles.removeCollectionButton}
              theme={themes.destructive}
              onClick={onRemoveCollection}
              text={
                <FormattedMessage
                  id={"Web3ConfigurationScreen.detail-dialog.remove-collection"}
                />
              }
            />
          </div>
        </div>

        <DialogFooter>
          <DefaultButton
            onClick={onDismiss}
            theme={themes.inverted}
            text={<FormattedMessage id="dismiss" />}
          />
        </DialogFooter>
      </Dialog>
    );
  };

export default Web3ConfigurationDetailDialog;
