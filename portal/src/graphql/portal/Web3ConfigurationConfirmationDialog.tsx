import React, { useContext, useMemo } from "react";
import { Dialog, DialogFooter, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  FormState as Web3ConfigurationFormState,
  isNFTCollectionEqual,
} from "./Web3ConfigurationScreen";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { NftCollection } from "./globalTypes.generated";
import styles from "./Web3ConfigurationConfirmationDialog.module.css";
import {
  getNetworkNameID,
  NetworkID,
  sameNetworkID,
} from "../../util/networkId";
import { truncateAddress } from "../../util/hex";

interface Web3ConfigurationConfirmationDialogProps {
  initialState: Web3ConfigurationFormState;
  currentState: Web3ConfigurationFormState;

  isVisible: boolean;
  onConfirm: () => void;
  onDismiss: () => void;
}

interface FormChanges {
  siweEnabled: boolean | null;
  networkChange: {
    from: NetworkID;
    to: NetworkID;
  } | null;

  tokenIDChanges: NftCollection[];

  collectionAdded: NftCollection[];
  collectionRemoved: NftCollection[];
}

const Web3ConfigurationConfirmationDialog: React.VFC<Web3ConfigurationConfirmationDialogProps> =
  function Web3ConfigurationConfirmationDialog(props) {
    const { initialState, currentState, isVisible, onConfirm, onDismiss } =
      props;

    const { renderToString } = useContext(Context);

    const dialogContentProps = useMemo(() => {
      return {
        title: (
          <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.title" />
        ),
      };
    }, []);

    const formChanges: FormChanges = useMemo(() => {
      const changes: FormChanges = {
        siweEnabled: null,
        networkChange: null,
        tokenIDChanges: [],
        collectionAdded: [],
        collectionRemoved: [],
      };

      if (initialState.siweChecked !== currentState.siweChecked) {
        changes.siweEnabled = currentState.siweChecked;
      }

      if (!sameNetworkID(initialState.network, currentState.network)) {
        changes.networkChange = {
          from: initialState.network,
          to: currentState.network,
        };
      }

      // We remove all collections if siwe is disabled
      // Remove all collections if siwe is disabled
      if (changes.siweEnabled === false) {
        changes.collectionAdded = [];
        changes.collectionRemoved = initialState.collections;
      } else if (!sameNetworkID(initialState.network, currentState.network)) {
        changes.collectionAdded = currentState.collections.filter((c) =>
          sameNetworkID(c, currentState.network)
        );
        changes.collectionRemoved = initialState.collections;
      } else {
        changes.collectionAdded = currentState.collections.filter(
          (c) =>
            initialState.collections.findIndex((cc) =>
              isNFTCollectionEqual(c, cc)
            ) === -1
        );

        changes.collectionRemoved = initialState.collections.filter(
          (c) =>
            currentState.collections.findIndex((cc) =>
              isNFTCollectionEqual(c, cc)
            ) === -1
        );
      }

      changes.tokenIDChanges = currentState.collections.filter(
        (c) =>
          initialState.collections.findIndex(
            (cc) =>
              isNFTCollectionEqual(c, cc) &&
              !c.tokenIDs.every((ct) => cc.tokenIDs.includes(ct))
          ) !== -1
      );

      return changes;
    }, [initialState, currentState]);

    return (
      <Dialog
        hidden={!isVisible}
        dialogContentProps={dialogContentProps}
        onDismiss={onDismiss}
      >
        <div className={styles.changesContainer}>
          {formChanges.siweEnabled !== null ? (
            <div className={styles.changesSectionContainer}>
              <Text className={styles.changesSectionTitle}>
                <FormattedMessage
                  id={
                    formChanges.siweEnabled
                      ? "Web3ConfigurationScreen.confirmation-dialog.siwe-enabled.title"
                      : "Web3ConfigurationScreen.confirmation-dialog.siwe-disabled.title"
                  }
                />
              </Text>
              <Text>
                <FormattedMessage
                  id={
                    formChanges.siweEnabled
                      ? "Web3ConfigurationScreen.confirmation-dialog.siwe-enabled.description"
                      : "Web3ConfigurationScreen.confirmation-dialog.siwe-disabled.description"
                  }
                />
              </Text>
            </div>
          ) : null}

          {formChanges.siweEnabled !== false &&
          formChanges.networkChange !== null ? (
            <div className={styles.changesSectionContainer}>
              <Text className={styles.changesSectionTitle}>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.network-changed.title" />
              </Text>
              <Text>
                <FormattedMessage
                  id="Web3ConfigurationScreen.confirmation-dialog.network-changed.description"
                  values={{
                    old: renderToString(
                      getNetworkNameID(formChanges.networkChange.from)
                    ),
                    new: renderToString(
                      getNetworkNameID(formChanges.networkChange.to)
                    ),
                  }}
                />
              </Text>
            </div>
          ) : null}

          {formChanges.tokenIDChanges.length > 0 ? (
            <div className={styles.changesSectionContainer}>
              <Text className={styles.changesSectionTitle}>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.tracked-token-changed.title" />
              </Text>
              <Text>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.tracked-token-changed.description" />
              </Text>
              <ul className={styles.changesSectionCollectionList}>
                {formChanges.tokenIDChanges.map((c) => {
                  const networkNameId = getNetworkNameID(c);
                  return (
                    <li key={c.contractAddress}>
                      <FormattedMessage
                        id="NftCollection.item.identifier"
                        values={{
                          name: c.name,
                          network: renderToString(networkNameId),
                          address: truncateAddress(c.contractAddress),
                        }}
                      />
                    </li>
                  );
                })}
              </ul>
            </div>
          ) : null}

          {formChanges.collectionRemoved.length > 0 ? (
            <div className={styles.changesSectionContainer}>
              <Text className={styles.changesSectionTitle}>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.collection-removed.title" />
              </Text>
              <Text>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.collection-removed.description" />
              </Text>
              <ul className={styles.changesSectionCollectionList}>
                {formChanges.collectionRemoved.map((c) => {
                  const networkNameId = getNetworkNameID(c);
                  return (
                    <li key={c.contractAddress}>
                      <FormattedMessage
                        id="NftCollection.item.identifier"
                        values={{
                          name: c.name,
                          network: renderToString(networkNameId),
                          address: truncateAddress(c.contractAddress),
                        }}
                      />
                    </li>
                  );
                })}
              </ul>
            </div>
          ) : null}

          {formChanges.collectionAdded.length > 0 ? (
            <div className={styles.changesSectionContainer}>
              <Text className={styles.changesSectionTitle}>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.collection-added.title" />
              </Text>
              <Text>
                <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.collection-added.description" />
              </Text>
              <ul className={styles.changesSectionCollectionList}>
                {formChanges.collectionAdded.map((c) => {
                  const networkNameId = getNetworkNameID(c);
                  return (
                    <li key={c.contractAddress}>
                      <FormattedMessage
                        id="NftCollection.item.identifier"
                        values={{
                          name: c.name,
                          network: renderToString(networkNameId),
                          address: truncateAddress(c.contractAddress),
                        }}
                      />
                    </li>
                  );
                })}
              </ul>
            </div>
          ) : null}
        </div>

        <DialogFooter>
          <PrimaryButton
            onClick={onConfirm}
            text={
              <FormattedMessage id="Web3ConfigurationScreen.confirmation-dialog.confirm" />
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

export default Web3ConfigurationConfirmationDialog;
