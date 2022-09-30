const ETHEREUM_ADDRESS_REGEX = /^0x[a-fA-F0-9]{40}$/;

export interface EIP681 {
  chainId: number;
  address: string;
}

export function parseEIP681(
  url: string,
  skipAddressCheck: boolean = false
): EIP681 {
  const protocolURI = url.split(":");

  if (protocolURI.length !== 2) {
    throw new Error(`Invalid URI: ${url}`);
  }

  if (protocolURI[0] !== "ethereum") {
    throw new Error(`Invalid protocol: ${protocolURI[0]}`);
  }

  const addressURI = protocolURI[1].split("@");

  if (addressURI.length !== 2) {
    throw new Error(`Invalid URI: ${url}`);
  }

  const address = addressURI[0];
  if (!skipAddressCheck && !ETHEREUM_ADDRESS_REGEX.test(address)) {
    throw new Error(`Invalid address: ${address}`);
  }

  const chainId = parseInt(addressURI[1], 10);
  if (chainId < 0) {
    throw new Error(`Chain ID cannot be negative: ${chainId}`);
  }

  return {
    chainId,
    address,
  };
}

export function createEIP681URL(
  eip681: EIP681,
  skipAddressCheck: boolean = false
): string {
  const url = `ethereum:${eip681.address}@${eip681.chainId}`;
  // Confirm the format is correct
  parseEIP681(url, skipAddressCheck);
  return url;
}

export function etherscanURL(
  eip681String: string,
  skipAddressCheck: boolean = false
): string {
  const eip681 = parseEIP681(eip681String, skipAddressCheck);

  let prefix: string;
  switch (eip681.chainId) {
    case 1:
      prefix = "https://etherscan.io/";
      break;
    case 5:
      prefix = "https://goerli.etherscan.io/";
      break;
    default:
      prefix = "";
  }

  return prefix;
}

export function etherscanAddress(
  eip681String: string,
  skipAddressCheck: boolean = false
): string {
  const eip681 = parseEIP681(eip681String, skipAddressCheck);
  const prefix = etherscanURL(eip681String, skipAddressCheck) + "address/";

  return prefix + eip681.address;
}

export function etherscanTx(
  eip681String: string,
  tx: string,
  skipAddressCheck: boolean = false
): string {
  const prefix = etherscanURL(eip681String, skipAddressCheck) + "tx/";

  return prefix + tx;
}
