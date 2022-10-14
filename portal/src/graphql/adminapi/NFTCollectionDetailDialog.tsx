import React, { useCallback, useContext, useMemo } from "react";
import {
  DetailsList,
  Dialog,
  DialogFooter,
  IColumn,
  IDialogContentProps,
  SelectionMode,
} from "@fluentui/react";
import { NFTContract, NFTToken } from "../../types";
import { truncateAddress } from "../../util/hex";

import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ExternalLink from "../../ExternalLink";
import { explorerBlock, explorerTx } from "../../util/eip681";
import DefaultButton from "../../DefaultButton";

const DIALOG_MAX_WIDTH = "80%";

interface NFTCollectionDetailDialogProps {
  contract: NFTContract;
  tokens: NFTToken[];
  eip681String: string;

  isVisible: boolean;
  onDismiss: () => void;
}

const NFTCollectionDetailDialog: React.VFC<NFTCollectionDetailDialogProps> = (
  props
) => {
  const { contract, tokens, eip681String, isVisible, onDismiss } = props;
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: `${contract.name} (${truncateAddress(contract.address)})`,
    };
  }, [contract]);

  const columns = useMemo(() => {
    return [
      {
        key: "token-id",
        name: renderToString(
          "UserDetails.connected-identities.siwe.nft-collections.token-id"
        ),
        minWidth: 100,
        maxWidth: 100,
      },
      {
        key: "transcation-hash",
        name: renderToString(
          "UserDetails.connected-identities.siwe.nft-collections.transaction-hash"
        ),
        minWidth: 221,
        maxWidth: 221,
      },
      {
        key: "balance",
        name: renderToString(
          "UserDetails.connected-identities.siwe.nft-collections.balance"
        ),
        minWidth: 60,
        maxWidth: 60,
      },
      {
        key: "block-index",
        name: renderToString(
          "UserDetails.connected-identities.siwe.nft-collections.block"
        ),
        minWidth: 100,
        maxWidth: 100,
      },
      {
        key: "timestamp",
        name: renderToString(
          "UserDetails.connected-identities.siwe.nft-collections.timestamp"
        ),
        minWidth: 205,
        maxWidth: 205,
      },
    ];
  }, [renderToString]);

  const onRenderItemColumn = useCallback(
    (item?: NFTToken, _index?: number, column?: IColumn) => {
      if (item == null) {
        return null;
      }

      switch (column?.key) {
        case "token-id":
          return (
            <span style={{ color: themes.main.palette.neutralDark }}>
              {truncateAddress(item.token_id)}
            </span>
          );
        case "transcation-hash":
          return (
            <ExternalLink
              href={explorerTx(eip681String, item.transaction_identifier.hash)}
            >
              {truncateAddress(item.transaction_identifier.hash)}
            </ExternalLink>
          );
        case "balance":
          return (
            <span style={{ color: themes.main.palette.neutralSecondary }}>
              {item.balance}
            </span>
          );
        case "block-index":
          return (
            <ExternalLink
              href={explorerBlock(
                eip681String,
                item.block_identifier.index.toString()
              )}
            >
              {item.block_identifier.index}
            </ExternalLink>
          );
        case "timestamp":
          return (
            <span style={{ color: themes.main.palette.neutralSecondary }}>
              {item.block_identifier.timestamp}
            </span>
          );
        default:
          return null;
      }
    },
    [
      eip681String,
      themes.main.palette.neutralDark,
      themes.main.palette.neutralSecondary,
    ]
  );

  return (
    <Dialog
      hidden={!isVisible}
      onDismiss={onDismiss}
      dialogContentProps={dialogContentProps}
      maxWidth={DIALOG_MAX_WIDTH}
    >
      <DetailsList
        columns={columns}
        items={tokens}
        onRenderItemColumn={onRenderItemColumn}
        selectionMode={SelectionMode.none}
      />

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

export default NFTCollectionDetailDialog;
