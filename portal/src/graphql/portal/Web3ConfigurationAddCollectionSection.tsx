import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { Text } from "@fluentui/react";
import styles from "./Web3ConfigurationAddCollectionSection.module.css";
import { createContractIDURL } from "../../util/contractId";
import { NetworkID } from "../../util/networkId";
import { CollectionItem } from "./Web3ConfigurationScreen";
import { DateTime } from "luxon";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ChoiceButton from "../../ChoiceButton";
import { NFTTokenType } from "../../types";
import Web3ConfigurationLargeCollectionDialog from "./Web3ConfigurationLargeCollectionDialog";
import Web3ConfigurationTokenTrackingDialog from "./Web3ConfigurationTokenTrackingDialog";
import TextField from "../../TextField";
import { NftCollection } from "./globalTypes.generated";

interface TokenTypeButtonsProps {
  disabled?: boolean;
  selectedType: NFTTokenType;
  onSelect: (type: NFTTokenType) => void;
}

const TokenTypeButtons: React.VFC<TokenTypeButtonsProps> = (props) => {
  const { selectedType, disabled, onSelect } = props;

  const onSelectERC721 = useCallback(() => {
    onSelect("erc721");
  }, [onSelect]);

  const onSelectERC1155 = useCallback(() => {
    onSelect("erc1155");
  }, [onSelect]);

  return (
    <div className={styles.tokenTypeChoiceButtonContainer}>
      <ChoiceButton
        className={styles.tokenTypeChoiceButton}
        disabled={disabled}
        checked={selectedType === "erc721"}
        onClick={onSelectERC721}
        text={
          <Text className={styles.tokenTypeChoiceButtonText}>
            <FormattedMessage
              id={`Web3ConfigurationScreen.collection-list.add-collection.token-type-button.erc721.title`}
            />
          </Text>
        }
        secondaryText={
          <Text className={styles.tokenTypeChoiceButtonText}>
            <FormattedMessage
              id={`Web3ConfigurationScreen.collection-list.add-collection.token-type-button.erc721.description`}
            />
          </Text>
        }
      />
      <ChoiceButton
        className={styles.tokenTypeChoiceButton}
        disabled={disabled}
        checked={selectedType === "erc1155"}
        onClick={onSelectERC1155}
        text={
          <Text className={styles.tokenTypeChoiceButtonText}>
            <FormattedMessage
              id={`Web3ConfigurationScreen.collection-list.add-collection.token-type-button.erc1155.title`}
            />
          </Text>
        }
        secondaryText={
          <Text className={styles.tokenTypeChoiceButtonText}>
            <FormattedMessage
              id={`Web3ConfigurationScreen.collection-list.add-collection.token-type-button.erc1155.description`}
            />
          </Text>
        }
      />
    </div>
  );
};

interface AddCollectionSectionValues {
  contractAddress: string;
  tokenType: NFTTokenType;
}

const defaultValues: AddCollectionSectionValues = {
  contractAddress: "",
  tokenType: "erc721",
};

interface AddCollectionSectionProps {
  className?: string;
  selectedNetwork: NetworkID;
  disabled?: boolean;
  initialValues?: AddCollectionSectionValues;
  fetchMetadata: (contractId: string) => Promise<NftCollection | null>;
  probeCollection: (contractId: string) => Promise<boolean>;
  onAdd: (collection: CollectionItem) => void;
  onDismiss: () => void;
}

const Web3ConfigurationAddCollectionForm: React.VFC<AddCollectionSectionProps> =
  function Web3ConfigurationAddCollectionForm(props) {
    const { renderToString } = useContext(Context);

    const {
      onAdd,
      onDismiss,
      fetchMetadata,
      probeCollection,
      initialValues = defaultValues,
      className,
      selectedNetwork,
    } = props;

    const [validationErrorId, setValidationErrorId] = useState<string | null>(
      null
    );
    const [activeDialog, setActiveDialog] = useState<
      "largeCollection" | "tokenTracking" | null
    >(null);
    const [isLoading, setIsLoading] = useState<boolean>(false);
    const [values, setValues] =
      useState<AddCollectionSectionValues>(initialValues);

    const buildContractID = useCallback(
      (address: string): string => {
        const contractId = {
          blockchain: selectedNetwork.blockchain,
          network: selectedNetwork.network,
          address: address,
        };

        return createContractIDURL(contractId);
      },
      [selectedNetwork]
    );

    const isModified = useMemo(() => {
      return values.contractAddress !== "";
    }, [values]);

    const resetValues = useCallback(() => {
      setValues(initialValues);
    }, [initialValues, setValues]);

    const onChangeContractAddress = useCallback(
      (_e, newValue) => {
        if (newValue != null) {
          setValues((prev) => ({ ...prev, contractAddress: newValue }));
        }
      },
      [setValues]
    );

    const onChangeTokenType = useCallback(
      (type: NFTTokenType) => {
        setValues((prev) => ({ ...prev, tokenType: type }));
      },
      [setValues]
    );

    const dismissDialogs = useCallback(() => {
      setActiveDialog(null);
    }, []);

    const handleAddCollection = useCallback(
      async (tokenIDs?: string[]) => {
        setIsLoading(true);
        setValidationErrorId(null);
        dismissDialogs();

        let contractID = "";
        try {
          contractID = buildContractID(values.contractAddress);
        } catch (_: unknown) {
          setValidationErrorId("errors.invalid-address");
          setIsLoading(false);
          return;
        }

        const metadata = await fetchMetadata(contractID);
        if (!metadata) {
          setIsLoading(false);
          return;
        }

        if (metadata.tokenType !== values.tokenType) {
          setValidationErrorId("errors.invalid-address");
          setIsLoading(false);
          return;
        }

        onAdd({
          ...metadata,
          createdAt: DateTime.now().toISO(),
          tokenIDs: tokenIDs ?? [],
          status: "pending",
        });
        setIsLoading(false);
        resetValues();
      },
      [
        buildContractID,
        dismissDialogs,
        fetchMetadata,
        onAdd,
        resetValues,
        values.contractAddress,
        values.tokenType,
      ]
    );

    const onAddCollection = useCallback(
      async (e) => {
        e.preventDefault();
        e.stopPropagation();

        setIsLoading(true);
        setValidationErrorId(null);

        let contractID = "";
        try {
          contractID = buildContractID(values.contractAddress);
        } catch (_: unknown) {
          setValidationErrorId("errors.invalid-address");
          setIsLoading(false);
          return;
        }

        const probeResult = await probeCollection(contractID);
        if (probeResult) {
          setActiveDialog("largeCollection");
          setIsLoading(false);
          return;
        }

        if (values.tokenType === "erc1155") {
          setActiveDialog("tokenTracking");
          setIsLoading(false);
          return;
        }

        // Add ERC-721 collection directly
        await handleAddCollection();
      },
      [
        probeCollection,
        values.tokenType,
        values.contractAddress,
        handleAddCollection,
        buildContractID,
      ]
    );

    const onCancel = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        onDismiss();
      },
      [onDismiss]
    );

    return (
      <div className={cn(styles.addCollection, className)}>
        <TextField
          label={renderToString(
            "Web3ConfigurationScreen.collection-list.add-collection.contract-address"
          )}
          className={styles.addCollectionField}
          placeholder={renderToString(
            "Web3ConfigurationScreen.collection-list.add-collection.placeholder"
          )}
          errorMessage={
            validationErrorId != null
              ? renderToString(validationErrorId)
              : undefined
          }
          value={values.contractAddress}
          onChange={onChangeContractAddress}
        />
        <TokenTypeButtons
          disabled={isLoading}
          selectedType={values.tokenType}
          onSelect={onChangeTokenType}
        />
        <div className={styles.addCollectionButtonContainer}>
          <PrimaryButton
            type="button"
            className={styles.addCollectionAddButton}
            disabled={!isModified || isLoading}
            onClick={onAddCollection}
            text={
              isLoading ? (
                <FormattedMessage id={"adding"} />
              ) : (
                <FormattedMessage id={"add"} />
              )
            }
          />
          <DefaultButton
            type="reset"
            disabled={isLoading}
            onClick={onCancel}
            text={<FormattedMessage id={"cancel"} />}
          />
        </div>
        <Web3ConfigurationLargeCollectionDialog
          isVisible={activeDialog === "largeCollection"}
          onDismiss={dismissDialogs}
        />
        <Web3ConfigurationTokenTrackingDialog
          isVisible={activeDialog === "tokenTracking"}
          onContinue={handleAddCollection}
          onDismiss={dismissDialogs}
        />
      </div>
    );
  };

export default Web3ConfigurationAddCollectionForm;
