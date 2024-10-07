import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import styles from "./Web3ConfigurationAddCollectionSection.module.css";
import { createContractIDURL } from "../../util/contractId";
import { NetworkID } from "../../util/networkId";
import { CollectionItem } from "./Web3ConfigurationScreen";
import { DateTime } from "luxon";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import Web3ConfigurationLargeCollectionDialog from "./Web3ConfigurationLargeCollectionDialog";
import Web3ConfigurationTokenTrackingDialog from "./Web3ConfigurationTokenTrackingDialog";
import Web3ConfigurationContactUsDialog from "./Web3ConfigurationContactUsDialog";
import TextField from "../../TextField";
import { useErrorDialog } from "../../formbinding";
import { makeReasonErrorParseRule } from "../../error/parse";
import { NftCollection } from "./globalTypes.generated";

interface AddCollectionSectionValues {
  contractAddress: string;
}

const defaultValues: AddCollectionSectionValues = {
  contractAddress: "",
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

    const {
      hidden: contactUsDialogHidden,
      onDismiss: onDismissContactUsDialog,
    } = useErrorDialog({
      parentJSONPointer: "",
      fieldName: "",
      rules: [makeReasonErrorParseRule("AlchemyProtocol", "")],
    });

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

    const dismissDialogs = useCallback(() => {
      setActiveDialog(null);
    }, []);

    const handleAddCollection = useCallback(
      async (tokenIDs?: string[]) => {
        setIsLoading(true);
        setValidationErrorId(null);
        dismissDialogs();

        let contractID: string;
        try {
          contractID = buildContractID(values.contractAddress);
        } catch (_: unknown) {
          setValidationErrorId("errors.invalid-address");
          setIsLoading(false);
          return;
        }

        try {
          const probeResult = await probeCollection(contractID);
          if (probeResult) {
            setActiveDialog("largeCollection");
            setIsLoading(false);
            return;
          }

          const metadata = await fetchMetadata(contractID);
          if (!metadata) {
            setIsLoading(false);
            return;
          }

          if (metadata.tokenType === "erc1155" && !tokenIDs?.length) {
            setActiveDialog("tokenTracking");
            setIsLoading(false);
            return;
          }

          onAdd({
            ...metadata,
            createdAt: DateTime.now().toISO(),
            tokenIDs: tokenIDs ?? [],
            status: "pending",
          });
        } catch (_: unknown) {
          // Error handled by parent component
          setIsLoading(false);
          return;
        }

        setIsLoading(false);
        resetValues();
      },
      [
        buildContractID,
        dismissDialogs,
        fetchMetadata,
        onAdd,
        probeCollection,
        resetValues,
        values.contractAddress,
      ]
    );

    const onAddCollection = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();

        handleAddCollection().catch(() => {});
      },
      [handleAddCollection]
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
      <>
        <form
          className={cn(styles.addCollection, className)}
          onSubmit={onAddCollection}
        >
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
          <div className={styles.addCollectionButtonContainer}>
            <PrimaryButton
              type="submit"
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
        </form>
        <Web3ConfigurationLargeCollectionDialog
          isVisible={activeDialog === "largeCollection"}
          onDismiss={dismissDialogs}
        />
        <Web3ConfigurationTokenTrackingDialog
          isVisible={activeDialog === "tokenTracking"}
          onContinue={handleAddCollection}
          onDismiss={dismissDialogs}
        />
        <Web3ConfigurationContactUsDialog
          isVisible={!contactUsDialogHidden}
          onDismiss={onDismissContactUsDialog}
        />
      </>
    );
  };

export default Web3ConfigurationAddCollectionForm;
