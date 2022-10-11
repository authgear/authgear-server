import React, { useCallback, useMemo } from "react";
import { Dialog, DialogFooter, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import DefaultButton from "../../DefaultButton";
import styles from "./Web3ConfigurationDetailDialog.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { getNetworkNameID } from "../../util/networkId";
import ActionButton from "../../ActionButton";
import { explorerAddress } from "../../util/eip681";
import ExternalLink from "../../ExternalLink";
import { createContractIDURL } from "../../util/contractId";
import { CollectionItem } from "./Web3ConfigurationScreen";

interface Web3ConfigurationDetailDialogProps {
  nftCollection: CollectionItem;

  isVisible: boolean;
  onDelete: (nftCollection: CollectionItem) => void;
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

    const contractID = useMemo(
      () =>
        createContractIDURL({
          address: nftCollection.contractAddress,
          blockchain: nftCollection.blockchain,
          network: nftCollection.network,
        }),
      [nftCollection]
    );

    const displayedTokens = useMemo(() => {
      const totalSupplyNotAvailable =
        !nftCollection.totalSupply || nftCollection.totalSupply === "0";

      // Check if collection is ERC-1155
      if (nftCollection.tokenIDs.length !== 0) {
        // Return tracked token count over total supply if available
        // otherwise just tracked token count
        return totalSupplyNotAvailable
          ? nftCollection.tokenIDs.length
          : `${nftCollection.tokenIDs.length}/${nftCollection.totalSupply}`;
      }

      // Return dash is total supply not available
      return totalSupplyNotAvailable ? "-" : nftCollection.totalSupply;
    }, [nftCollection]);

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
            <ExternalLink href={explorerAddress(contractID)}>
              <FormattedMessage id="Web3ConfigurationScreen.detail-dialog.view-on-explorer" />
            </ExternalLink>
          </div>

          <div className={styles.fieldContainer}>
            <Text className={styles.fieldTitle} block={true}>
              <FormattedMessage id="Web3ConfigurationScreen.detail-dialog.network" />
            </Text>
            <Text as="p" block={true}>
              <FormattedMessage id={getNetworkNameID(nftCollection)} />
            </Text>
          </div>

          <div className={styles.fieldContainer}>
            <Text className={styles.fieldTitle} block={true}>
              <FormattedMessage
                id={
                  nftCollection.totalSupply == null &&
                  nftCollection.tokenIDs.length === 0
                    ? "Web3ConfigurationScreen.detail-dialog.total-supply"
                    : "Web3ConfigurationScreen.detail-dialog.tracked-tokens"
                }
              />
            </Text>
            <Text as="p" block={true}>
              {displayedTokens}
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
