declare interface SIWENonce {
  nonce: string;
  expireAt: Date;
}

declare interface SIWEVerifiedData {
  message: string;
  signature: string;
  encodedPubKey: string;
}
// Define ethereum interface in window
declare interface Window {
  ethereum?: import("ethers").providers.ExternalProvider;
}
