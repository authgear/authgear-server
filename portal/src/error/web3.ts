export interface APIBadNFTCollectionError {
  errorName: string;
  reason: "BadNFTCollection";
  info: {
    type: string;
  };
}
