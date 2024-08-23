import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  Text,
  Dropdown,
  IColumn,
  SelectionMode,
  DetailsList,
} from "@fluentui/react";
import { APIError } from "../../error/error";
import { produce } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { PortalAPIAppConfig } from "../../types";
import { useCheckbox, useDropdown } from "../../hook/useInput";
import { clearEmptyObject } from "../../util/misc";
import { useParams } from "react-router-dom";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import FormContainer from "../../FormContainer";

import styles from "./Web3ConfigurationScreen.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useNftCollectionsQuery } from "./query/nftCollectionsQuery";
import { NftCollection } from "./globalTypes.generated";
import { createContractIDURL, parseContractID } from "../../util/contractId";
import { useNftContractMetadataLazyQuery } from "./query/nftContractMetadataQuery";
import { ErrorParseRule, makeReasonErrorParseRule } from "../../error/parse";
import {
  ALL_SUPPORTED_NETWORKS,
  createNetworkIDURL,
  getNetworkNameID,
  NetworkID,
  parseNetworkID,
  sameNetworkID,
} from "../../util/networkId";
import Web3ConfigurationConfirmationDialog from "./Web3ConfigurationConfirmationDialog";
import Web3ConfigurationDetailDialog from "./Web3ConfigurationDetailDialog";
import Web3ConfigurationCollectionDeletionDialog from "./Web3ConfigurationCollectionDeletionDialog";
import Web3ConfigurationAddCollectionSection from "./Web3ConfigurationAddCollectionSection";
import CommandBarButton from "../../CommandBarButton";
import ActionButton from "../../ActionButton";
import Toggle from "../../Toggle";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import HorizontalDivider from "../../HorizontalDivider";
import { useProbeNFTCollectionMutation } from "./mutations/probeNFTCollectionMutation";
import Web3ConfigurationTokenTrackingDialog from "./Web3ConfigurationTokenTrackingDialog";
import { truncateAddress } from "../../util/hex";
import Web3ConfigurationNetworkChangeDialog from "./Web3ConfigurationNetworkChangeDialog";
import { FormModel } from "../../FormContainerBase";

export interface CollectionItem extends NftCollection {
  status: "pending" | "active";
  tokenIDs: string[];
}

export function isNFTCollectionEqual(
  a: NftCollection,
  b: NftCollection
): boolean {
  return (
    a.blockchain === b.blockchain &&
    a.network === b.network &&
    a.contractAddress === b.contractAddress
  );
}
export interface FormState {
  network: NetworkID;
  collections: CollectionItem[];

  siweChecked: boolean;
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.authentication ??= {};
    config.authentication.primary_authenticators ??= [];
    config.identity ??= {};
    config.identity.login_id ??= {};
    config.web3 ??= {};
    config.web3.nft ??= {};
    config.web3.siwe ??= {};

    if (currentState.siweChecked) {
      config.authentication.identities = ["siwe"];
      config.identity.login_id.keys = [];
      config.authentication.primary_authenticators = [];
    } else {
      config.authentication.identities = ["login_id", "oauth"];
      config.identity.login_id.keys = [{ type: "email" }];
      config.authentication.primary_authenticators = ["password"];
    }

    const selectedNetwork = createNetworkIDURL(currentState.network);

    let collections: CollectionItem[] = [];

    if (!currentState.siweChecked) {
      // Clear collection list if SIWE is disabled
      collections = [];
    } else if (!config.web3.siwe.networks?.includes(selectedNetwork)) {
      // Clear unrelated collections if network is changed
      collections = currentState.collections.filter((c) =>
        sameNetworkID(c, currentState.network)
      );
    } else {
      // Proceed with changes
      collections = currentState.collections;
    }

    config.web3.siwe.networks = [selectedNetwork];
    config.web3.nft.collections = collections.map((c) => {
      return createContractIDURL({
        blockchain: c.blockchain,
        network: c.network,
        address: c.contractAddress,
        query:
          c.tokenIDs.length !== 0
            ? new URLSearchParams(c.tokenIDs.map((t) => ["token_ids", t]))
            : undefined,
      });
    });

    clearEmptyObject(config);
  });
}

const DuplicatedContractError: APIError = {
  errorName: "DuplicatedContractError",
  reason: "ValidationFailed",
  info: {
    causes: [
      {
        location: "/contract_address",
        kind: "__local",
        details: {
          error: {
            messageID: "errors.duplicated-contract",
          },
        },
      },
    ],
  },
};

const ALL_NETWORK_OPTIONS: string[] = ALL_SUPPORTED_NETWORKS.map((n) =>
  createNetworkIDURL(n)
);

interface Web3ConfigurationContentProps {
  nftCollections: NftCollection[];
  isAddCollectionFieldVisible: boolean;
  showAddCollectionField: () => void;
  hideAddCollectionField: () => void;

  maximumCollections: number;
  fetchMetadata: (contractId: string) => Promise<NftCollection | null>;
  probeCollection: (contractId: string) => Promise<boolean>;
  form: AppConfigFormModel<FormState>;
}

type Web3ConfigurationContentDialogs =
  | "deletionConfirmation"
  | "detail"
  | "networkChange"
  | null;

const Web3ConfigurationContent: React.VFC<Web3ConfigurationContentProps> =
  // eslint-disable-next-line complexity
  function Web3ConfigurationContent(props) {
    const { state, setState } = props.form;
    const { themes } = useSystemConfig();

    const [activeDialog, setActiveDialog] =
      useState<Web3ConfigurationContentDialogs>(null);
    const [selectedCollectionIdx, setSelectedCollectionIdx] =
      useState<number>(-1);
    const [isTokenTrackingDialogVisible, setIsTokenTrackingDialogVisible] =
      useState<boolean>(false);

    const [pendingNetwork, setPendingNetwork] = useState<string | null>(null);

    const { renderToString } = useContext(Context);

    const { onChange: onChangeSIWEChecked } = useCheckbox(
      (checked: boolean) => {
        setState((prev) => ({
          ...prev,
          siweChecked: checked,
        }));
      }
    );

    const renderBlockchainNetwork = useCallback(
      (networkIdUrl: string) => {
        const networkId = parseNetworkID(networkIdUrl);
        return renderToString(getNetworkNameID(networkId));
      },
      [renderToString]
    );

    const onNetworkDropdownOptionChange = useCallback(
      (option: string) => {
        const networkId = parseNetworkID(option);
        if (
          state.collections.length > 0 &&
          !sameNetworkID(networkId, state.network)
        ) {
          setPendingNetwork(option);
          setActiveDialog("networkChange");
          return;
        }

        setState((prev) => ({
          ...prev,
          network: networkId,
        }));
      },
      [state.collections, state.network, setState]
    );

    const { options: blockchainOptions, onChange: onBlockchainChange } =
      useDropdown<string>(
        ALL_NETWORK_OPTIONS,
        onNetworkDropdownOptionChange,
        createNetworkIDURL(state.network),
        renderBlockchainNetwork
      );

    const openDetailDialog = useCallback(() => {
      setActiveDialog("detail");
    }, [setActiveDialog]);

    const openDeleteConfirmationDialog = useCallback(() => {
      setActiveDialog("deletionConfirmation");
    }, [setActiveDialog]);

    const dismissAllDialogs = useCallback(() => {
      setActiveDialog(null);
    }, [setActiveDialog]);

    const onSelectCollection = useCallback(
      (collection: CollectionItem) => {
        const idx = state.collections.findIndex((c) =>
          isNFTCollectionEqual(collection, c)
        );

        setSelectedCollectionIdx(idx);
      },
      [state.collections]
    );

    const onAddNewCollection = useCallback(
      (collection: CollectionItem) => {
        if (
          state.collections.findIndex(
            (c) =>
              c.blockchain === collection.blockchain &&
              c.network === collection.network &&
              c.contractAddress === collection.contractAddress
          ) !== -1
        ) {
          // eslint-disable-next-line @typescript-eslint/no-throw-literal
          throw DuplicatedContractError;
        }

        setState((prev) => {
          const existingCollections = prev.collections;

          return {
            ...prev,
            collections: [collection, ...existingCollections],
          };
        });
      },
      [state, setState]
    );

    const onRequireConfirmRemoveCollection = useCallback(
      (collection: CollectionItem) => {
        onSelectCollection(collection);

        openDeleteConfirmationDialog();
      },
      [onSelectCollection, openDeleteConfirmationDialog]
    );

    const onRemoveCollection = useCallback(
      (collection: NftCollection) => {
        setSelectedCollectionIdx(-1);
        setState((prev) => {
          const collections = prev.collections;
          const index = collections.findIndex((c) =>
            isNFTCollectionEqual(c, collection)
          );

          if (index < 0) {
            return prev;
          }

          return {
            ...prev,
            collections: [
              ...collections.slice(0, index),
              ...collections.slice(index + 1),
            ],
          };
        });

        dismissAllDialogs();
      },
      [setState, dismissAllDialogs]
    );

    const confirmNetworkChange = useCallback(() => {
      if (!pendingNetwork) {
        return;
      }

      setState((prev) => ({
        ...prev,
        network: parseNetworkID(pendingNetwork),
        collections: [],
      }));

      setPendingNetwork(null);
    }, [pendingNetwork, setState]);

    const onCollectionUserActionClick = useCallback(
      (e: React.MouseEvent<unknown>, collection: CollectionItem) => {
        e.preventDefault();
        e.stopPropagation();

        switch (collection.status) {
          case "pending":
            onRemoveCollection(collection);
            break;
          case "active":
            onSelectCollection(collection);

            openDetailDialog();
            break;
        }
      },
      [onRemoveCollection, onSelectCollection, openDetailDialog]
    );

    const showTokenTrackingDialog = useCallback(() => {
      setIsTokenTrackingDialogVisible(true);
    }, []);

    const dismissTokenTrackingDialog = useCallback(() => {
      setIsTokenTrackingDialogVisible(false);
    }, []);

    const onEditSelectedCollectionTokenIds = useCallback(
      (tokenIDs: string[]) => {
        if (selectedCollectionIdx === -1) {
          return;
        }

        setState((prev) => {
          const collections = prev.collections;
          const updatedCollection = {
            ...prev.collections[selectedCollectionIdx],
            tokenIDs: tokenIDs,
          };
          collections.splice(selectedCollectionIdx, 1, updatedCollection);
          return {
            ...prev,
            collections,
          };
        });
        dismissTokenTrackingDialog();
      },
      [selectedCollectionIdx, setState, dismissTokenTrackingDialog]
    );

    const onRenderItemColumn = useCallback(
      (item?: CollectionItem, _index?: number, column?: IColumn) => {
        if (item == null) {
          return null;
        }
        switch (column?.key) {
          case "name":
            return (
              <span style={{ color: themes.main.palette.neutralDark }}>
                {item.name}
              </span>
            );
          case "contract-address":
            return truncateAddress(item.contractAddress);
          case "status":
            if (item.status === "pending") {
              return renderToString(
                "Web3ConfigurationScreen.collection-list.status.pending"
              );
            }
            return "";
          case "action": {
            const theme =
              item.status === "pending"
                ? themes.destructive
                : themes.actionButton;

            const text =
              item.status === "pending" ? (
                <FormattedMessage id="Web3ConfigurationScreen.colleciton-list.action.remove" />
              ) : (
                <FormattedMessage id="Web3ConfigurationScreen.colleciton-list.action.details" />
              );

            return (
              <ActionButton
                className={styles.actionButton}
                type="button"
                theme={theme}
                onClick={(event) => onCollectionUserActionClick(event, item)}
                text={text}
              />
            );
          }

          default:
            return null;
        }
      },
      [onCollectionUserActionClick, renderToString, themes]
    );

    const collectionColumns: IColumn[] = useMemo(
      () => [
        {
          key: "name",
          name: "",
          minWidth: 179,
          maxWidth: 179,
          isMultiline: true,
        },
        {
          key: "contract-address",
          name: "",
          flexGrow: 1,
          minWidth: 103,
          maxWidth: 103,
        },
        {
          key: "status",
          name: "",
          minWidth: 103,
          maxWidth: 103,
        },
        {
          key: "action",
          name: "",
          minWidth: 103,
          maxWidth: 103,
          targetWidthProportion: 1,
        },
      ],
      []
    );

    const collectionLimitReached = useMemo(() => {
      return state.collections.length >= props.maximumCollections;
    }, [props.maximumCollections, state.collections.length]);

    const selectedCollection = useMemo(() => {
      if (selectedCollectionIdx === -1) {
        return null;
      }

      return state.collections[selectedCollectionIdx] ?? null;
    }, [state, selectedCollectionIdx]);

    return (
      <>
        <ScreenContent>
          <ScreenTitle className={styles.widget}>
            <FormattedMessage id="Web3ConfigurationScreen.title" />
          </ScreenTitle>
          <ScreenDescription className={styles.widget}>
            <FormattedMessage id="Web3ConfigurationScreen.description" />
          </ScreenDescription>
          <Widget className={styles.widget} />
          <Widget className={styles.widget}>
            <div>
              <div>
                <Toggle
                  label={renderToString("Web3ConfigurationScreen.siwe.title")}
                  checked={state.siweChecked}
                  onChange={onChangeSIWEChecked}
                  inlineLabel={false}
                  description={
                    <FormattedMessage id="Web3ConfigurationScreen.siwe.description" />
                  }
                />
              </div>
              <div className={styles.networkSection}>
                <Dropdown
                  className={styles.networkDropdown}
                  label={renderToString(
                    "Web3ConfigurationScreen.network-droplist.label"
                  )}
                  disabled={!state.siweChecked}
                  options={blockchainOptions}
                  selectedKey={createNetworkIDURL(state.network)}
                  onChange={onBlockchainChange}
                />
                <Text
                  as="p"
                  variant="small"
                  block={true}
                  className={styles.networkDropdownWarning}
                >
                  <FormattedMessage id="Web3ConfigurationScreen.network-droplist.warning" />
                </Text>
              </div>
              <HorizontalDivider />
              <div className={styles.nftCollectionList}>
                <Text
                  className={styles.nftCollectionTitle}
                  variant="medium"
                  style={
                    !state.siweChecked
                      ? {
                          color: themes.main.palette.neutralTertiary,
                        }
                      : undefined
                  }
                  block={true}
                >
                  <FormattedMessage id="Web3ConfigurationScreen.collection-list.title" />
                </Text>
                <CommandBarButton
                  className={styles.addCollectionButton}
                  iconProps={{ iconName: "Add" }}
                  disabled={!state.siweChecked || collectionLimitReached}
                  text={renderToString(
                    "Web3ConfigurationScreen.collection-list.add-collection"
                  )}
                  onClick={props.showAddCollectionField}
                />
                <HorizontalDivider />
                {props.isAddCollectionFieldVisible &&
                !collectionLimitReached ? (
                  <Web3ConfigurationAddCollectionSection
                    className={styles.addCollectionForm}
                    selectedNetwork={state.network}
                    onAdd={onAddNewCollection}
                    onDismiss={props.hideAddCollectionField}
                    fetchMetadata={props.fetchMetadata}
                    probeCollection={props.probeCollection}
                  />
                ) : null}
                <div className={styles.listWrapper}>
                  {collectionLimitReached ? (
                    <FeatureDisabledMessageBar
                      messageID="FeatureConfig.web3-nft.maximum"
                      messageValues={{
                        maximum: props.maximumCollections,
                      }}
                    />
                  ) : null}
                  <DetailsList
                    className={styles.nftCollectionList}
                    selectionMode={SelectionMode.none}
                    onRenderItemColumn={onRenderItemColumn}
                    isHeaderVisible={false}
                    columns={collectionColumns}
                    items={state.collections}
                  />
                </div>
                {state.siweChecked ? (
                  <Text variant="medium" block={true}>
                    <FormattedMessage id="Web3ConfigurationScreen.collection-list.description" />
                  </Text>
                ) : null}
              </div>
            </div>
          </Widget>
        </ScreenContent>
        {selectedCollection !== null ? (
          <Web3ConfigurationDetailDialog
            nftCollection={selectedCollection}
            isVisible={activeDialog === "detail"}
            onEditTrackedTokens={showTokenTrackingDialog}
            onDismiss={dismissAllDialogs}
            onDelete={onRequireConfirmRemoveCollection}
          />
        ) : null}
        {selectedCollection !== null ? (
          <Web3ConfigurationCollectionDeletionDialog
            nftCollection={selectedCollection}
            isVisible={activeDialog === "deletionConfirmation"}
            onDismiss={dismissAllDialogs}
            onConfirm={onRemoveCollection}
          />
        ) : null}
        {selectedCollection !== null ? (
          <Web3ConfigurationTokenTrackingDialog
            isVisible={isTokenTrackingDialogVisible}
            initialValue={selectedCollection.tokenIDs}
            onContinue={onEditSelectedCollectionTokenIds}
            onDismiss={dismissTokenTrackingDialog}
          />
        ) : null}
        {pendingNetwork !== null ? (
          <Web3ConfigurationNetworkChangeDialog
            isVisible={activeDialog === "networkChange"}
            onConfirm={confirmNetworkChange}
            onDismiss={dismissAllDialogs}
          />
        ) : null}
      </>
    );
  };

const Web3ConfigurationScreen: React.VFC = function Web3ConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const [isConfirmationDialogVisible, setIsConfirmationDialogVisible] =
    useState<boolean>(false);

  const [isAddCollectionFieldVisible, setIsAddCollectionFieldVisible] =
    useState<boolean>(false);

  const nftCollections = useNftCollectionsQuery(appID);

  const featureConfig = useAppFeatureConfigQuery(appID);

  const { fetch: fetchMetadata, error: fetchMetadataError } =
    useNftContractMetadataLazyQuery();

  const { probeNFTCollection, error: probeNFTCollectionError } =
    useProbeNFTCollectionMutation();

  const constructFormState = useCallback(
    (config: PortalAPIAppConfig) => {
      const siweIndex = config.authentication?.identities?.indexOf("siwe");
      const siweChecked = siweIndex != null && siweIndex >= 0;

      let siweNetworks = (config.web3?.siwe?.networks ?? []).map((n) =>
        parseNetworkID(n)
      );
      if (siweNetworks.length === 0) {
        siweNetworks = [
          {
            blockchain: "ethereum",
            network: "1",
          },
        ];
      }

      // We support 1 chain for now
      const [selectedNetwork] = siweNetworks;

      const contractIDs = config.web3?.nft?.collections ?? [];

      const collectionMap = new Map<string, NftCollection>();
      nftCollections.collections.forEach((c) => {
        collectionMap.set(
          createContractIDURL({
            blockchain: c.blockchain,
            network: c.network,
            address: c.contractAddress,
          }),
          c
        );
      });

      const existingCollections = contractIDs
        .map((c) => parseContractID(c))
        .map<CollectionItem | null>((cid) => {
          const collection = collectionMap.get(
            createContractIDURL({
              blockchain: cid.blockchain,
              network: cid.network,
              address: cid.address,
            })
          );

          const tokens = cid.query?.getAll("token_ids") ?? [];

          if (!collection) {
            return null;
          }

          return {
            ...collection,
            tokenIDs: tokens,
            status: "active",
          };
        })
        .filter((c): c is CollectionItem => c !== null);

      return {
        siweChecked,
        collections: existingCollections,
        network: selectedNetwork,
      };
    },
    [nftCollections]
  );

  const form = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });

  const openConfirmationDialog = useCallback(() => {
    setIsConfirmationDialogVisible(true);
  }, [setIsConfirmationDialogVisible]);

  const dismissConfirmationDialog = useCallback(() => {
    setIsConfirmationDialogVisible(false);
  }, [setIsConfirmationDialogVisible]);

  const showAddCollectionField = useCallback(() => {
    setIsAddCollectionFieldVisible(true);
  }, []);

  const hideAddCollectionField = useCallback(() => {
    setIsAddCollectionFieldVisible(false);
  }, []);

  const saveForm = useCallback(async () => {
    dismissConfirmationDialog();
    hideAddCollectionField();

    await form.save();

    await nftCollections.refetch();
  }, [form, nftCollections, dismissConfirmationDialog, hideAddCollectionField]);

  const onFormSave = useCallback(async () => {
    openConfirmationDialog();
  }, [openConfirmationDialog]);

  const formModel: FormModel = {
    ...form,
    save: onFormSave,
  };

  const errorRules: ErrorParseRule[] = useMemo(() => {
    return [
      makeReasonErrorParseRule(
        "ServiceUnavailable",
        "errors.service-unavailable"
      ),
      makeReasonErrorParseRule("BadNFTCollection", "errors.bad-nft-collection"),
    ];
  }, []);

  const collectionsMaximum = useMemo(() => {
    return featureConfig.effectiveFeatureConfig?.web3?.nft?.maximum ?? 3;
  }, [featureConfig]);

  if (form.isLoading || nftCollections.loading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return (
      <ShowError
        error={form.loadError}
        onRetry={() => {
          form.reload();
        }}
      />
    );
  }

  if (nftCollections.error) {
    return (
      <ShowError
        error={nftCollections.error}
        onRetry={nftCollections.refetch}
      />
    );
  }

  return (
    <FormContainer
      form={formModel}
      errorRules={errorRules}
      localError={fetchMetadataError || probeNFTCollectionError}
    >
      <Web3ConfigurationContent
        form={form}
        isAddCollectionFieldVisible={isAddCollectionFieldVisible}
        maximumCollections={collectionsMaximum}
        nftCollections={nftCollections.collections}
        fetchMetadata={fetchMetadata}
        probeCollection={probeNFTCollection}
        showAddCollectionField={showAddCollectionField}
        hideAddCollectionField={hideAddCollectionField}
      />

      <Web3ConfigurationConfirmationDialog
        isVisible={isConfirmationDialogVisible}
        onDismiss={dismissConfirmationDialog}
        onConfirm={saveForm}
        currentState={form.state}
        initialState={form.initialState}
      />
    </FormContainer>
  );
};

export default Web3ConfigurationScreen;
