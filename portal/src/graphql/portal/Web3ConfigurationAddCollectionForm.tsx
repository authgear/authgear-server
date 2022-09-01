import React, { useCallback, useContext, useMemo } from "react";
import { PrimaryButton } from "@fluentui/react";
import cn from "classnames";
import { APIError } from "../../error/error";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import styles from "./Web3ConfigurationAddCollectionForm.module.css";
import { createContractIdURL } from "../../util/contractId";
import { FormProvider } from "../../form";
import FormTextField from "../../FormTextField";
import { useSimpleForm } from "../../hook/useSimpleForm";
import { LazyQueryResult, OperationVariables } from "@apollo/client";
import { NftContractMetadataQueryQuery } from "./query/nftContractMetadataQuery.generated";
import { NetworkId } from "../../util/networkId";
import { CollectionItem } from "./Web3ConfigurationScreen";
import { DateTime } from "luxon";

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

interface AddCollectionFormProps {
  className?: string;
  selectedNetwork: NetworkId;
  fetchMetadata: (
    contractId: string
  ) => Promise<
    LazyQueryResult<NftContractMetadataQueryQuery, OperationVariables>
  >;
  onAdd: (collection: CollectionItem) => void;
}

interface AddCollectionSectionFormState {
  contractAddress: string;
}

function makeDefaultAddCollectionSectionFormState(): AddCollectionSectionFormState {
  return {
    contractAddress: "",
  };
}

const Web3ConfigurationAddCollectionForm: React.VFC<AddCollectionFormProps> =
  function Web3ConfigurationAddCollectionForm(props) {
    const { renderToString } = useContext(Context);

    const { onAdd, fetchMetadata, className, selectedNetwork } = props;

    const onSubmit = useCallback(
      async (state: AddCollectionSectionFormState) => {
        const contractId = {
          blockchain: selectedNetwork.blockchain,
          network: selectedNetwork.network,
          address: state.contractAddress,
        };

        const metadataResponse = await fetchMetadata(
          createContractIdURL(contractId)
        );
        if (metadataResponse.error) {
          throw metadataResponse.error;
        }
        const metadata = metadataResponse.data?.nftContractMetadata;

        if (!metadata || metadata.tokenType !== "erc721") {
          // eslint-disable-next-line @typescript-eslint/no-throw-literal
          throw InvalidAddressError;
        }

        onAdd({
          blockchain: contractId.blockchain,
          network: contractId.network,
          contractAddress: metadata.address,
          name: metadata.name,
          blockHeight: 0,
          createdAt: DateTime.now().toISO(),
          totalSupply: parseInt(metadata.totalSupply, 10),
          type: metadata.tokenType,
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
      state: { contractAddress },
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

    return (
      <FormProvider loading={isUpdating} error={updateError}>
        <div className={cn(styles.addCollection, className)}>
          <FormTextField
            className={styles.addCollectionField}
            placeholder={renderToString(
              "Web3ConfigurationScreen.collection-list.add-collection.placeholder"
            )}
            value={contractAddress}
            onChange={onChangeContractAddress}
            parentJSONPointer=""
            fieldName="contract_address"
          />
          <PrimaryButton
            type="submit"
            disabled={!isModified || isUpdating}
            onClick={onSubmitForm}
          >
            {isUpdating ? (
              <FormattedMessage id={"adding"} />
            ) : (
              <FormattedMessage id={"add"} />
            )}
          </PrimaryButton>
        </div>
      </FormProvider>
    );
  };

export default Web3ConfigurationAddCollectionForm;
