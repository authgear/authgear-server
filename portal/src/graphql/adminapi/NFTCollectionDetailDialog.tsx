import React, { useCallback, useMemo } from "react";
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

import { FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ExternalLink from "../../ExternalLink";
import { etherscanBlock, etherscanTx } from "../../util/eip681";
import DefaultButton from "../../DefaultButton";

const DIALOG_MAX_WIDTH = "80%";

interface NFTCollectionDetailDialogProps {
  contract: NFTContract;
  balance: number;
  tokens: NFTToken[];
  eip681String: string;

  isVisible: boolean;
  onDismiss: () => void;
}

const NFTCollectionDetailDialog: React.VFC<NFTCollectionDetailDialogProps> = (
  props
) => {
  const { contract, balance, tokens, eip681String, isVisible, onDismiss } =
    props;
  const { themes } = useSystemConfig();

  const dialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: `${contract.name} (${truncateAddress(contract.address)})`,
    };
  }, [contract]);

  const columns = useMemo(() => {
    return [
      {
        key: "token-id",
        name: "Token ID",
        minWidth: 60,
        maxWidth: 60,
      },
      {
        key: "transcation-hash",
        name: "Txn Hash",
        minWidth: 221,
        maxWidth: 221,
      },
      {
        key: "block-index",
        name: "Block",
        minWidth: 100,
        maxWidth: 100,
      },
      {
        key: "timestamp",
        name: "Timestamp",
        minWidth: 205,
        maxWidth: 205,
      },
    ];
  }, []);

  const onRenderItemColumn = useCallback(
    (item?: NFTToken, _index?: number, column?: IColumn) => {
      if (item == null) {
        return null;
      }

      switch (column?.key) {
        case "token-id":
          return (
            <span style={{ color: themes.main.palette.neutralDark }}>
              {`#${item.token_id}`}
            </span>
          );
        case "transcation-hash":
          return (
            <ExternalLink
              href={etherscanTx(eip681String, item.transaction_identifier.hash)}
            >
              {truncateAddress(item.transaction_identifier.hash)}
            </ExternalLink>
          );
        case "block-index":
          return (
            <ExternalLink
              href={etherscanBlock(
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
      <FormattedMessage
        id="UserDetails.connected-identities.siwe.nft-collections.balance"
        values={{ balance: balance }}
      />
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
