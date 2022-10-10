import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { APIError } from "../../error/error";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { Text } from "@fluentui/react";
import styles from "./Web3ConfigurationAddCollectionForm.module.css";
import { createContractIDURL } from "../../util/contractId";
import { FormProvider } from "../../form";
import FormTextField from "../../FormTextField";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { LazyQueryResult, OperationVariables } from "@apollo/client";
import { NftContractMetadataQueryQuery } from "./query/nftContractMetadataQuery.generated";
import { NetworkID } from "../../util/networkId";
import { CollectionItem } from "./Web3ConfigurationScreen";
import { DateTime } from "luxon";
import { parseRawError } from "../../error/parse";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ChoiceButton from "../../ChoiceButton";
import { NFTTokenType } from "../../types";

const InvalidAddressError: APIError = {
  errorName: "InvalidAddressError",
  reason: "ValidationFailed",
  info: {
    causes: [
      {
        location: "/contract_address",
        kind: "__local",
        details: {
          error: {
            messageID: "errors.invalid-address",
          },
        },
      },
    ],
  },
};

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

interface AddCollectionFormProps {
  className?: string;
  selectedNetwork: NetworkID;
  fetchMetadata: (
    contractId: string
  ) => Promise<
    LazyQueryResult<NftContractMetadataQueryQuery, OperationVariables>
  >;
  onAdd: (collection: CollectionItem) => void;
  onCancel: () => void;
}

interface AddCollectionSectionFormState {
  contractAddress: string;
  tokenType: NFTTokenType;
}

function makeDefaultAddCollectionSectionFormState(): AddCollectionSectionFormState {
  return {
    contractAddress: "",
    tokenType: "erc721",
  };
}

const Web3ConfigurationAddCollectionForm: React.VFC<AddCollectionFormProps> =
  function Web3ConfigurationAddCollectionForm(props) {
    const { renderToString } = useContext(Context);

    const { onAdd, onCancel, fetchMetadata, className, selectedNetwork } =
      props;

    const onSubmit = useCallback(
      async (state: AddCollectionSectionFormState) => {
        const contractId = {
          blockchain: selectedNetwork.blockchain,
          network: selectedNetwork.network,
          address: state.contractAddress,
        };

        let contractID = "";
        try {
          contractID = createContractIDURL(contractId);
        } catch (_: unknown) {
          // eslint-disable-next-line @typescript-eslint/no-throw-literal
          throw InvalidAddressError;
        }

        const metadataResponse = await fetchMetadata(contractID);
        if (metadataResponse.error) {
          // eslint-disable-next-line @typescript-eslint/no-throw-literal
          throw parseRawError(metadataResponse.error);
        }
        const metadata = metadataResponse.data?.nftContractMetadata;

        if (!metadata || metadata.tokenType !== "erc721") {
          // eslint-disable-next-line @typescript-eslint/no-throw-literal
          throw InvalidAddressError;
        }

        onAdd({
          ...metadata,
          createdAt: DateTime.now().toISO(),
          status: "pending",
        });
      },
      [fetchMetadata, onAdd, selectedNetwork]
    );

    const form = useSimpleForm({
      stateMode:
        "ConstantInitialStateAndResetCurrentStatetoInitialStateAfterSave",
      defaultState: makeDefaultAddCollectionSectionFormState(),
      submit: onSubmit,
    });

    const {
      updateError,
      save,
      isUpdating,
      state: { contractAddress, tokenType },
      setState,
    } = form;

    const onChangeContractAddress = useCallback(
      (_e, newValue) => {
        if (newValue != null) {
          setState((prev) => ({ ...prev, contractAddress: newValue }));
        }
      },
      [setState]
    );

    const onChangeTokenType = useCallback(
      (type: NFTTokenType) => {
        setState((prev) => ({ ...prev, tokenType: type }));
      },
      [setState]
    );

    const isModified = useMemo(() => {
      return contractAddress !== "";
    }, [contractAddress]);

    const onSubmitForm = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        save().catch(() => {});
      },
      [save]
    );

    const onCancelForm = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        onCancel();
      },
      [onCancel]
    );

    return (
      <FormProvider loading={isUpdating} error={updateError}>
        <div className={cn(styles.addCollection, className)}>
          <FormTextField
            label={renderToString(
              "Web3ConfigurationScreen.collection-list.add-collection.contract-address"
            )}
            className={styles.addCollectionField}
            placeholder={renderToString(
              "Web3ConfigurationScreen.collection-list.add-collection.placeholder"
            )}
            value={contractAddress}
            onChange={onChangeContractAddress}
            parentJSONPointer=""
            fieldName="contract_address"
          />
          <TokenTypeButtons
            disabled={isUpdating}
            selectedType={tokenType}
            onSelect={onChangeTokenType}
          />
          <div className={styles.addCollectionButtonContainer}>
            <PrimaryButton
              type="submit"
              className={styles.addCollectionAddButton}
              disabled={!isModified || isUpdating}
              onClick={onSubmitForm}
              text={
                isUpdating ? (
                  <FormattedMessage id={"adding"} />
                ) : (
                  <FormattedMessage id={"add"} />
                )
              }
            />
            <DefaultButton
              type="reset"
              disabled={isUpdating}
              onClick={onCancelForm}
              text={<FormattedMessage id={"cancel"} />}
            />
          </div>
        </div>
      </FormProvider>
    );
  };

export default Web3ConfigurationAddCollectionForm;
