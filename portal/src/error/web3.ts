export interface APIBadNFTCollectionError {
  errorName: string;
  reason: "BadNFTCollection";
  info: {
    type: string;
  };
}

export interface APIAlchemyProtocolError {
  errorName: string;
  reason: "AlchemyProtocol";
}
